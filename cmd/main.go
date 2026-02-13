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

	services, err := app.NewServices(pool, cfg.GeminiAPIKey, cfg.JWTSecret, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("failed to init services: %v", err)
	}

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
	log.Println("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}

	log.Println("server exited")
}
