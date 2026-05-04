package main

import (
	"context"
	"fmt"
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

	var vID, mID, sID string
	err = conn.QueryRow(ctx, "SELECT id, main_category_id, sub_category_id FROM vacancies WHERE id::text LIKE '88b1fbba%'").Scan(&vID, &mID, &sID)
	if err != nil {
		log.Fatalf("Vacancy not found or error: %v", err)
	}

	fmt.Printf("Vacancy ID: %s\nMain Category: %s\nSub Category: %s\n", vID, mID, sID)

	var count int
	err = conn.QueryRow(ctx, "SELECT count(*) FROM candidate_search_profiles WHERE main_category_id = $1 AND sub_category_id = $2", mID, sID).Scan(&count)
	if err != nil {
		log.Fatalf("Count error: %v", err)
	}

	fmt.Printf("Matching candidates in candidate_search_profiles: %d\n", count)
}
