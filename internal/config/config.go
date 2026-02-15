package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	ServerPort       string
	GeminiAPIKey     string
	JWTSecret        string
	QdrantURL        string
	QdrantAPIKey     string
	RedisURL         string
	TelegramBotToken string
	MinioEndpoint    string
	MinioAccessKey   string
	MinioSecretKey   string
	MinioBucket      string
	MinioUseSSL      bool
	WebAppURL        string
}

func Load() (*Config, error) {
	_ = godotenv.Overload()

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "11911"
	}

	geminiKey := os.Getenv("GEMINI_API_KEY")
	if geminiKey == "" {
		return nil, fmt.Errorf("GEMINI_API_KEY is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return nil, fmt.Errorf("JWT_SECRET is required")
	}

	qdrantURL := os.Getenv("QDRANT_URL")
	if qdrantURL == "" {
		return nil, fmt.Errorf("QDRANT_URL is required")
	}

	qdrantAPIKey := os.Getenv("QDRANT_API_KEY")
	if qdrantAPIKey == "" {
		return nil, fmt.Errorf("QDRANT_API_KEY is required")
	}

	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		return nil, fmt.Errorf("REDIS_URL is required")
	}

	tgBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")

	minioEndpoint := os.Getenv("MINIO_ENDPOINT")
	if minioEndpoint == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT is required")
	}
	minioAccessKey := os.Getenv("MINIO_ACCESS_KEY")
	if minioAccessKey == "" {
		return nil, fmt.Errorf("MINIO_ACCESS_KEY is required")
	}
	minioSecretKey := os.Getenv("MINIO_SECRET_KEY")
	if minioSecretKey == "" {
		return nil, fmt.Errorf("MINIO_SECRET_KEY is required")
	}
	minioBucket := os.Getenv("MINIO_BUCKET")
	if minioBucket == "" {
		minioBucket = "hr-ai"
	}
	minioUseSSL := os.Getenv("MINIO_USE_SSL") == "true"
	webAppURL := os.Getenv("WEBAPP_URL")

	return &Config{
		DatabaseURL:      dbURL,
		ServerPort:       port,
		GeminiAPIKey:     geminiKey,
		JWTSecret:        jwtSecret,
		QdrantURL:        qdrantURL,
		QdrantAPIKey:     qdrantAPIKey,
		RedisURL:         redisURL,
		TelegramBotToken: tgBotToken,
		MinioEndpoint:    minioEndpoint,
		MinioAccessKey:   minioAccessKey,
		MinioSecretKey:   minioSecretKey,
		MinioBucket:      minioBucket,
		MinioUseSSL:      minioUseSSL,
		WebAppURL:        webAppURL,
	}, nil
}
