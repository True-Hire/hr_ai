package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyReferenceRepository struct {
	pool *pgxpool.Pool
}

func NewCompanyReferenceRepository(pool *pgxpool.Pool) *CompanyReferenceRepository {
	return &CompanyReferenceRepository{pool: pool}
}

func (r *CompanyReferenceRepository) GetByName(ctx context.Context, name string) (*domain.CompanyReference, error) {
	query := `
		SELECT company_name, normalized_name, prestige_score, engineering_score,
			scale_score, hiring_bar_score, category, updated_at
		FROM company_references
		WHERE LOWER(company_name) = LOWER($1)
	`
	row := r.pool.QueryRow(ctx, query, name)
	return scanCompanyReference(row)
}

func (r *CompanyReferenceRepository) GetByNormalizedName(ctx context.Context, normalizedName string) (*domain.CompanyReference, error) {
	query := `
		SELECT company_name, normalized_name, prestige_score, engineering_score,
			scale_score, hiring_bar_score, category, updated_at
		FROM company_references
		WHERE normalized_name = $1
	`
	row := r.pool.QueryRow(ctx, query, normalizedName)
	return scanCompanyReference(row)
}

func (r *CompanyReferenceRepository) ListAll(ctx context.Context) ([]domain.CompanyReference, error) {
	query := `
		SELECT company_name, normalized_name, prestige_score, engineering_score,
			scale_score, hiring_bar_score, category, updated_at
		FROM company_references
		ORDER BY company_name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list company references: %w", err)
	}
	defer rows.Close()

	var results []domain.CompanyReference
	for rows.Next() {
		var c domain.CompanyReference
		err := rows.Scan(
			&c.CompanyName,
			&c.NormalizedName,
			&c.PrestigeScore,
			&c.EngineeringScore,
			&c.ScaleScore,
			&c.HiringBarScore,
			&c.Category,
			&c.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan company reference row: %w", err)
		}
		results = append(results, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate company references: %w", err)
	}
	return results, nil
}

func scanCompanyReference(row pgx.Row) (*domain.CompanyReference, error) {
	var c domain.CompanyReference
	err := row.Scan(
		&c.CompanyName,
		&c.NormalizedName,
		&c.PrestigeScore,
		&c.EngineeringScore,
		&c.ScaleScore,
		&c.HiringBarScore,
		&c.Category,
		&c.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("company reference not found")
		}
		return nil, fmt.Errorf("scan company reference: %w", err)
	}
	return &c, nil
}
