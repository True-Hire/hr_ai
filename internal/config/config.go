package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL  string
	ServerPort   string
	GeminiAPIKey string
	JWTSecret    string
	QdrantURL    string
	QdrantAPIKey string
	RedisURL     string
}

func Load() (*Config, error) {
	_ = godotenv.Load()

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

	return &Config{
		DatabaseURL:  dbURL,
		ServerPort:   port,
		GeminiAPIKey: geminiKey,
		JWTSecret:    jwtSecret,
		QdrantURL:    qdrantURL,
		QdrantAPIKey: qdrantAPIKey,
		RedisURL:     redisURL,
	}, nil
}
