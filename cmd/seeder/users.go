package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func SeedFakeUsers(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("--- Seeding Fake Users ---")

	mainCatID, err := uuid.Parse("e9569147-ee1f-58de-956c-88e6c5438c42")
	if err != nil {
		log.Printf("Invalid main_category_id: %v", err)
		return
	}

	subCatID, err := uuid.Parse("9c5c6118-5dab-520d-938d-c796d1859ee2")
	if err != nil {
		log.Printf("Invalid sub_category_id: %v", err)
		return
	}

	firstNames := []string{"Ali", "Vali", "Gani", "Sami", "Karim", "Zarif", "Botir", "Sobir", "Murod", "Umid", "Jasur", "Sardor", "Farhod", "Rustam", "Olim", "Aziz", "Bekzod", "Dilshod", "Sherzod", "Sanjar"}
	lastNames := []string{"Aliyev", "Valiyev", "Ganiyev", "Samiyev", "Karimov", "Zarifov", "Botirov", "Sobirov", "Murodov", "Umidov", "Jasurov", "Sardorov", "Farhodov", "Rustamov", "Olimov", "Azizov", "Bekzodov", "Dilshodov", "Sherzodov", "Sanjarov"}

	for i := 0; i < 20; i++ {
		id := uuid.New()
		firstName := firstNames[i%len(firstNames)]
		lastName := lastNames[i%len(lastNames)]
		phone := fmt.Sprintf("+99890%07d", 1234567+i)         // Unique phone
		email := fmt.Sprintf("user%d@example.com", 1234567+i) // Unique email
		now := time.Now()

		_, err := pool.Exec(ctx, `
			INSERT INTO users (
				id, first_name, last_name, phone, email, 
				status, tariff_type, main_category_id, sub_category_id, created_at
			) VALUES (
				$1, $2, $3, $4, $5, 
				'active', 'free', $6, $7, $8
			) ON CONFLICT (id) DO NOTHING`,
			id, firstName, lastName, phone, email, mainCatID, subCatID, now)

		if err != nil {
			log.Printf("Failed to seed user %s %s: %v", firstName, lastName, err)
		} else {
			log.Printf("Seeded user: %s %s", firstName, lastName)
		}
	}

	log.Println("Fake Users seeding completed.")
}
