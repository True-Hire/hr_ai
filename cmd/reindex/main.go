package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"

	"github.com/ruziba3vich/hr-ai/internal/application"
	"github.com/ruziba3vich/hr-ai/internal/infrastructure/repository"
)

func main() {
	_ = godotenv.Load()
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepository(pool)
	pfRepo := repository.NewProfileFieldRepository(pool)
	pftRepo := repository.NewProfileFieldTextRepository(pool)
	expRepo := repository.NewExperienceItemRepository(pool)
	eduRepo := repository.NewEducationItemRepository(pool)
	itRepo := repository.NewItemTextRepository(pool)
	skillRepo := repository.NewSkillRepository(pool)
	cspRepo := repository.NewCandidateSearchProfileRepository(pool)
	companyRefRepo := repository.NewCompanyReferenceRepository(pool)
	universityRefRepo := repository.NewUniversityReferenceRepository(pool)

	userSvc := application.NewUserService(userRepo)
	pfSvc := application.NewProfileFieldService(pfRepo)
	pftSvc := application.NewProfileFieldTextService(pftRepo, nil)
	expSvc := application.NewExperienceItemService(expRepo)
	eduSvc := application.NewEducationItemService(eduRepo)
	itSvc := application.NewItemTextService(itRepo, nil)
	skillSvc := application.NewSkillService(skillRepo)

	indexingSvc := application.NewCandidateIndexingService(
		cspRepo, companyRefRepo, universityRefRepo,
		userSvc, pfSvc, pftSvc, expSvc, eduSvc, itSvc, skillSvc,
	)

	// Get unindexed users
	rows, err := pool.Query(ctx, `SELECT u.id FROM users u LEFT JOIN candidate_search_profiles csp ON u.id = csp.user_id WHERE csp.user_id IS NULL`)
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			continue
		}
		ids = append(ids, id)
	}

	fmt.Printf("Found %d unindexed users\n", len(ids))

	indexed := 0
	for _, id := range ids {
		if err := indexingSvc.IndexCandidate(ctx, id); err != nil {
			log.Printf("index %s: %v", id, err)
		} else {
			fmt.Printf("indexed %s\n", id)
			indexed++
		}
	}

	var total int
	_ = pool.QueryRow(ctx, "SELECT COUNT(*) FROM candidate_search_profiles").Scan(&total)
	fmt.Printf("\nDone! Indexed %d/%d users. Total profiles: %d\n", indexed, len(ids), total)
}
