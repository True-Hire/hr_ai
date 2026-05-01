package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	// Database URL from .env
	dbURL := "postgresql://hr_ai:kR9mXw4vJpL2nT7s@hr-ai-db.compile-me.uz:5455/hr_ai_db"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	log.Println("Starting migration...")

	_, err = conn.Exec(ctx, `
		-- Add category columns to vacancies table
		ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS main_category_id UUID REFERENCES main_category(id) ON DELETE SET NULL;
		ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS sub_category_id UUID REFERENCES sub_category(id) ON DELETE SET NULL;

		-- Add category columns to users table
		ALTER TABLE users ADD COLUMN IF NOT EXISTS main_category_id UUID REFERENCES main_category(id) ON DELETE SET NULL;
		ALTER TABLE users ADD COLUMN IF NOT EXISTS sub_category_id UUID REFERENCES sub_category(id) ON DELETE SET NULL;

		CREATE INDEX IF NOT EXISTS idx_users_main_category_id ON users(main_category_id);
		CREATE INDEX IF NOT EXISTS idx_users_sub_category_id ON users(sub_category_id);
	`)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration successful: Added main_category_id and sub_category_id to vacancies and users tables.")
}
