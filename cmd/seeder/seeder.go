package main

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ClearSeeder removes all existing data from taxonomy tables.
func ClearSeeder(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("--- Deleting Existing Seeder Data ---")
	_, err := pool.Exec(ctx, `TRUNCATE main_category, sub_category, technologies, skills CASCADE;`)
	if err != nil {
		log.Printf("Failed to clear seeder data: %v", err)
	} else {
		log.Println("Data cleared successfully.")
	}
}

// RunSeeder populates the database with default taxonomies.
func RunSeeder(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("--- Seeding Main Categories ---")
	mainCatMap := make(map[string]uuid.UUID)
	for _, name := range MainCategories {
		id := generateID(name)
		mainCatMap[name] = id
		now := time.Now()

		_, err := pool.Exec(ctx, `
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

			_, err := pool.Exec(ctx, `
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

		_, err := pool.Exec(ctx, `
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

		_, err := pool.Exec(ctx, `
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
