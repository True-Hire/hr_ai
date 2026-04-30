package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruziba3vich/hr-ai/internal/config"
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

	log.Println("--- Seeding Main Categories ---")
	mainCatMap := make(map[string]uuid.UUID)
	for _, name := range MainCategories {
		id := generateID(name)
		mainCatMap[name] = id
		now := time.Now()

		_, err = pool.Exec(ctx, `
			INSERT INTO main_category (id, name, created_at, updated_at)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at`,
			id, name, now, now)
		if err != nil {
			log.Printf("Failed main category %s: %v", name, err)
		}
	}

	log.Println("--- Seeding Sub Categories ---")
	subCatMap := make(map[string]uuid.UUID)
	for _, d := range SubCategories {
		mainID := mainCatMap[d.MainCategory]
		for _, subName := range d.Names {
			compositeKey := d.MainCategory + ":" + subName
			subID := generateID(compositeKey)
			subCatMap[compositeKey] = subID
			now := time.Now()

			_, err = pool.Exec(ctx, `
				INSERT INTO sub_category (id, name, main_category_id, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5)
				ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, updated_at = EXCLUDED.updated_at`,
				subID, subName, mainID, now, now)
			if err != nil {
				log.Printf("  Failed sub category %s: %v", subName, err)
			}
		}
	}

	log.Println("--- Seeding Technologies ---")
	for _, t := range Technologies {
		techID := generateID("tech:" + t.Name)
		var subIDs []uuid.UUID
		for _, key := range t.SubCategoryKeys {
			if id, ok := subCatMap[key]; ok {
				subIDs = append(subIDs, id)
			}
		}
		now := time.Now()

		_, err = pool.Exec(ctx, `
			INSERT INTO technologies (id, name, sub_category_ids, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET 
				name = EXCLUDED.name, 
				sub_category_ids = EXCLUDED.sub_category_ids, 
				updated_at = EXCLUDED.updated_at`,
			techID, t.Name, subIDs, now, now)
		if err != nil {
			log.Printf("Failed technology %s: %v", t.Name, err)
		}
	}

	log.Println("--- Seeding Skills ---")
	for _, s := range Skills {
		skillID := generateID("skill:" + s.Name)
		var subIDs []uuid.UUID
		for _, key := range s.SubCategoryKeys {
			if id, ok := subCatMap[key]; ok {
				subIDs = append(subIDs, id)
			}
		}
		now := time.Now()

		_, err = pool.Exec(ctx, `
			INSERT INTO skills (id, name, sub_category_ids, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5)
			ON CONFLICT (id) DO UPDATE SET 
				name = EXCLUDED.name, 
				sub_category_ids = EXCLUDED.sub_category_ids, 
				updated_at = EXCLUDED.updated_at`,
			skillID, s.Name, subIDs, now, now)
		if err != nil {
			log.Printf("Failed skill %s: %v", s.Name, err)
		} else {
			log.Printf("Seeded Skill: %s", s.Name)
		}
	}

	log.Println("Seeding completed successfully!")
}
