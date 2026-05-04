package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgresql://hr_ai:kR9mXw4vJpL2nT7s@hr-ai-db.compile-me.uz:5455/hr_ai_db"
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `
		-- Add match details to vacancy_workers
		ALTER TABLE vacancy_workers ADD COLUMN IF NOT EXISTS match_percentage INT NOT NULL DEFAULT 0;
		ALTER TABLE vacancy_workers ADD COLUMN IF NOT EXISTS match_score NUMERIC(6,3) NOT NULL DEFAULT 0;
		ALTER TABLE vacancy_workers ADD COLUMN IF NOT EXISTS rank INT NOT NULL DEFAULT 0;

		-- Ensure uniqueness to avoid duplicate matches for the same vacancy/user
		DO $$ 
		BEGIN 
			IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_vacancy_workers_vacancy_user') THEN
				CREATE UNIQUE INDEX idx_vacancy_workers_vacancy_user ON vacancy_workers(vacancy_id, user_id);
			END IF;
		END $$;
	`)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration successful: Added match details to vacancy_workers table.")
}
