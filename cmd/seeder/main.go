package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruziba3vich/hr-ai/internal/config"
)

const (
	EnableSeeder     = false
	DeleteBeforeSeed = true
)

func generateID(name string) uuid.UUID {
	return uuid.NewSHA1(uuid.NameSpaceDNS, []byte(name))
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	var pool *pgxpool.Pool
	for i := 0; i < 15; i++ {
		pool, err = pgxpool.New(ctx, cfg.DatabaseURL)
		if err == nil {
			if err = pool.Ping(ctx); err == nil {
				break
			}
		}
		log.Printf("Waiting for database... retry %d/15", i+1)
		time.Sleep(2 * time.Second)
	}

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer pool.Close()

	if DeleteBeforeSeed {
		ClearSeeder(ctx, pool)
	}

	if EnableSeeder {
		RunSeeder(ctx, pool)
	} else {
		log.Println("Seeding is currently DISABLED via EnableSeeder=false flag.")
	}
}
