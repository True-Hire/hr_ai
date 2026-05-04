package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyWorkerRepository struct {
	pool *pgxpool.Pool
}

func NewVacancyWorkerRepository(pool *pgxpool.Pool) *VacancyWorkerRepository {
	return &VacancyWorkerRepository{pool: pool}
}

func (r *VacancyWorkerRepository) Create(ctx context.Context, worker *domain.VacancyWorker) error {
	query := `
		INSERT INTO vacancy_workers (id, vacancy_id, user_id, match_percentage, match_score, rank, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (vacancy_id, user_id) DO UPDATE SET
			match_percentage = EXCLUDED.match_percentage,
			match_score = EXCLUDED.match_score,
			rank = EXCLUDED.rank,
			updated_at = now()
	`
	_, err := r.pool.Exec(ctx, query,
		worker.ID,
		worker.VacancyID,
		worker.UserID,
		worker.MatchPercentage,
		worker.MatchScore,
		worker.Rank,
		worker.CreatedAt,
		worker.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("create vacancy worker: %w", err)
	}
	return nil
}

func (r *VacancyWorkerRepository) BulkCreate(ctx context.Context, workers []domain.VacancyWorker) error {
	if len(workers) == 0 {
		return nil
	}

	// Simple implementation using single query for now, can be optimized with CopyFrom if needed
	for _, w := range workers {
		if err := r.Create(ctx, &w); err != nil {
			return err
		}
	}
	return nil
}

func (r *VacancyWorkerRepository) ListByVacancy(ctx context.Context, vacancyID uuid.UUID) ([]domain.VacancyWorker, error) {
	query := `
		SELECT id, vacancy_id, user_id, match_percentage, match_score, rank, created_at, updated_at
		FROM vacancy_workers
		WHERE vacancy_id = $1
		ORDER BY rank ASC
	`
	rows, err := r.pool.Query(ctx, query, vacancyID)
	if err != nil {
		return nil, fmt.Errorf("query vacancy workers: %w", err)
	}
	defer rows.Close()

	var workers []domain.VacancyWorker
	for rows.Next() {
		var w domain.VacancyWorker
		err := rows.Scan(
			&w.ID,
			&w.VacancyID,
			&w.UserID,
			&w.MatchPercentage,
			&w.MatchScore,
			&w.Rank,
			&w.CreatedAt,
			&w.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan vacancy worker: %w", err)
		}
		workers = append(workers, w)
	}
	return workers, nil
}

func (r *VacancyWorkerRepository) DeleteByVacancy(ctx context.Context, vacancyID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM vacancy_workers WHERE vacancy_id = $1", vacancyID)
	if err != nil {
		return fmt.Errorf("delete vacancy workers: %w", err)
	}
	return nil
}
