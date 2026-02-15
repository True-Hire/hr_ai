package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/ruziba3vich/hr-ai/internal/app"
	"github.com/ruziba3vich/hr-ai/internal/config"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/postgres"

	tele "gopkg.in/telebot.v3"
)

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

	services, err := app.NewServices(pool, cfg.GeminiAPIKey, cfg.JWTSecret, cfg.DatabaseURL, cfg.QdrantURL, cfg.QdrantAPIKey, cfg.RedisURL, cfg.MinioEndpoint, cfg.MinioAccessKey, cfg.MinioSecretKey, cfg.MinioBucket, cfg.MinioUseSSL)
	if err != nil {
		log.Fatalf("failed to init services: %v", err)
	}
	defer services.RedisClient.Close()

	bot, err := tele.NewBot(tele.Settings{
		Token:  cfg.TelegramBotToken,
		Poller: &tele.LongPoller{Timeout: 10 * time.Second},
	})
	if err != nil {
		log.Fatalf("failed to create telegram bot: %v", err)
	}

	botSvc := services.Bot

	bot.Handle("/start", func(c tele.Context) error {
		sender := c.Sender()

		var photoData []byte
		photos, err := bot.ProfilePhotosOf(sender)
		if err == nil && len(photos) > 0 {
			reader, err := bot.File(&photos[0].File)
			if err == nil {
				photoData, _ = io.ReadAll(reader)
				reader.Close()
			}
		}

		user, isNew, err := botSvc.HandleStart(ctx, sender.ID, sender.FirstName, sender.LastName, sender.Username, photoData)
		if err != nil {
			log.Printf("handle /start error for %d: %v", sender.ID, err)
			return c.Send("Something went wrong. Please try again.")
		}

		if isNew {
			return c.Send("Welcome, " + user.FirstName + "! Your profile has been created.\n\nSend me your resume as text or a file (PDF, image) and I'll parse it for you.")
		}
		return c.Send("Welcome back, " + user.FirstName + "!\n\nSend me your resume as text or a file (PDF, image) to update your profile.")
	})

	bot.Handle(tele.OnText, func(c tele.Context) error {
		sender := c.Sender()

		user, _, err := botSvc.HandleStart(ctx, sender.ID, sender.FirstName, sender.LastName, sender.Username, nil)
		if err != nil {
			log.Printf("ensure user error for %d: %v", sender.ID, err)
			return c.Send("Something went wrong. Please try /start first.")
		}

		_ = c.Send("Parsing your resume... This may take a moment.")

		result, err := botSvc.HandleResumeText(ctx, user.ID, c.Text())
		if err != nil {
			log.Printf("parse resume text error for %s: %v", user.ID, err)
			return c.Send("Failed to parse your resume. Please try again.")
		}

		return c.Send("Your resume has been parsed successfully!\n\n" +
			"Source language: " + result.SourceLang + "\n" +
			"Profile fields: " + itoa(len(result.Fields)) + "\n" +
			"Experience items: " + itoa(len(result.Experience)) + "\n" +
			"Education items: " + itoa(len(result.Education)))
	})

	bot.Handle(tele.OnDocument, func(c tele.Context) error {
		sender := c.Sender()
		doc := c.Message().Document

		user, _, err := botSvc.HandleStart(ctx, sender.ID, sender.FirstName, sender.LastName, sender.Username, nil)
		if err != nil {
			log.Printf("ensure user error for %d: %v", sender.ID, err)
			return c.Send("Something went wrong. Please try /start first.")
		}

		mimeType := doc.MIME
		if !isAllowedMIME(mimeType) {
			return c.Send("Unsupported file type. Please send a PDF, image (PNG/JPG), or text file.")
		}

		reader, err := bot.File(&doc.File)
		if err != nil {
			log.Printf("download file error: %v", err)
			return c.Send("Failed to download your file. Please try again.")
		}
		defer reader.Close()

		fileData, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("read file error: %v", err)
			return c.Send("Failed to read your file. Please try again.")
		}

		_ = c.Send("Parsing your document... This may take a moment.")

		result, err := botSvc.HandleResumeFile(ctx, user.ID, fileData, mimeType)
		if err != nil {
			log.Printf("parse resume file error for %s: %v", user.ID, err)
			return c.Send("Failed to parse your document. Please try again.")
		}

		return c.Send("Your document has been parsed successfully!\n\n" +
			"Source language: " + result.SourceLang + "\n" +
			"Profile fields: " + itoa(len(result.Fields)) + "\n" +
			"Experience items: " + itoa(len(result.Experience)) + "\n" +
			"Education items: " + itoa(len(result.Education)))
	})

	bot.Handle(tele.OnPhoto, func(c tele.Context) error {
		sender := c.Sender()
		photo := c.Message().Photo

		user, _, err := botSvc.HandleStart(ctx, sender.ID, sender.FirstName, sender.LastName, sender.Username, nil)
		if err != nil {
			log.Printf("ensure user error for %d: %v", sender.ID, err)
			return c.Send("Something went wrong. Please try /start first.")
		}

		reader, err := bot.File(&photo.File)
		if err != nil {
			log.Printf("download photo error: %v", err)
			return c.Send("Failed to download your photo. Please try again.")
		}
		defer reader.Close()

		fileData, err := io.ReadAll(reader)
		if err != nil {
			log.Printf("read photo error: %v", err)
			return c.Send("Failed to read your photo. Please try again.")
		}

		_ = c.Send("Parsing your image... This may take a moment.")

		result, err := botSvc.HandleResumeFile(ctx, user.ID, fileData, "image/jpeg")
		if err != nil {
			log.Printf("parse resume photo error for %s: %v", user.ID, err)
			return c.Send("Failed to parse your image. Please try again.")
		}

		return c.Send("Your image has been parsed successfully!\n\n" +
			"Source language: " + result.SourceLang + "\n" +
			"Profile fields: " + itoa(len(result.Fields)) + "\n" +
			"Experience items: " + itoa(len(result.Experience)) + "\n" +
			"Education items: " + itoa(len(result.Education)))
	})

	go func() {
		log.Println("telegram bot starting...")
		bot.Start()
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down bot...")
	bot.Stop()
	log.Println("bot stopped")
}

func isAllowedMIME(mime string) bool {
	switch mime {
	case "application/pdf", "image/png", "image/jpeg", "text/plain":
		return true
	}
	return false
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
