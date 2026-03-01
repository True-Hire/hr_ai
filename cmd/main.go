package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ruziba3vich/hr-ai/internal/app"
	"github.com/ruziba3vich/hr-ai/internal/config"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/postgres"
	httphandler "github.com/ruziba3vich/hr-ai/internal/interfaces/http"
	"github.com/ruziba3vich/hr-ai/internal/telegram"

	_ "github.com/ruziba3vich/hr-ai/docs"
)

// @title HR AI API
// @version 1.0
// @description HR AI platform API with multilingual profile parsing powered by Gemini
// @BasePath /api/v1
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	services, err := app.NewServices(pool, cfg.GeminiAPIKey, cfg.JWTSecret, cfg.DatabaseURL, cfg.QdrantURL, cfg.QdrantAPIKey, cfg.RedisURL, cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL, cfg.TelegramBotToken, cfg.TelegramHRBotToken)
	if err != nil {
		log.Fatalf("failed to init services: %v", err)
	}
	defer services.RedisClient.Close()

	// Start Telegram bot in background
	var tgBot *telegram.Bot
	if cfg.TelegramBotToken != "" {
		tgBot, err = telegram.NewBot(cfg.TelegramBotToken, services.Bot, cfg.WebAppURL)
		if err != nil {
			log.Fatalf("failed to init telegram bot: %v", err)
		}
		go tgBot.Start()
	} else {
		log.Println("TELEGRAM_BOT_TOKEN not set, skipping bot")
	}

	// Start HR Telegram bot in background
	var hrBot *telegram.HRBot
	if cfg.TelegramHRBotToken != "" {
		hrBot, err = telegram.NewHRBot(cfg.TelegramHRBotToken, services.HRBot)
		if err != nil {
			log.Fatalf("failed to init hr telegram bot: %v", err)
		}
		go hrBot.Start()
	} else {
		log.Println("TELEGRAM_HR_BOT_TOKEN not set, skipping HR bot")
	}

	// Start HTTP server
	router := httphandler.NewRouter(services)
	srv := &http.Server{
		Addr:    ":" + cfg.ServerPort,
		Handler: router,
	}

	go func() {
		log.Printf("server starting on port %s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	if tgBot != nil {
		tgBot.Stop()
	}
	if hrBot != nil {
		hrBot.Stop()
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
