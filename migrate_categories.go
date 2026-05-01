package main

import (
	"context"
	"log"

	"github.com/jackc/pgx/v5"
)

func main() {
	dbURL := "postgresql://hr_ai:kR9mXw4vJpL2nT7s@hr-ai-db.compile-me.uz:5455/hr_ai_db"
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dbURL)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	_, err = conn.Exec(ctx, `
		ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS main_category_id UUID REFERENCES main_category(id) ON DELETE SET NULL;
		ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS sub_category_id UUID REFERENCES sub_category(id) ON DELETE SET NULL;
	`)
	if err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	log.Println("Migration successful: Added main_category_id and sub_category_id to vacancies table.")
}
