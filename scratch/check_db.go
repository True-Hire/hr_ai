package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	connStr := os.Getenv("DATABASE_URL")
	if connStr == "" {
		connStr = "postgres://postgres:postgres@localhost:5432/hr_ai?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), connStr)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	fmt.Println("--- TECHNOLOGIES ---")
	rows, _ := pool.Query(context.Background(), "SELECT name FROM technologies LIMIT 50")
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("- %s\n", name)
	}
	rows.Close()

	fmt.Println("\n--- SKILLS ---")
	rows, _ = pool.Query(context.Background(), "SELECT name FROM skills LIMIT 50")
	for rows.Next() {
		var name string
		rows.Scan(&name)
		fmt.Printf("- %s\n", name)
	}
	rows.Close()
}
