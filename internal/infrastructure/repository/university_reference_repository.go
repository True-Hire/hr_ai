package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type UniversityReferenceRepository struct {
	pool *pgxpool.Pool
}

func NewUniversityReferenceRepository(pool *pgxpool.Pool) *UniversityReferenceRepository {
	return &UniversityReferenceRepository{pool: pool}
}

func (r *UniversityReferenceRepository) GetByName(ctx context.Context, name string) (*domain.UniversityReference, error) {
	query := `
		SELECT institution_name, normalized_name, education_score, category, updated_at
		FROM university_references
		WHERE LOWER(institution_name) = LOWER($1)
	`
	row := r.pool.QueryRow(ctx, query, name)
	return scanUniversityReference(row)
}

func (r *UniversityReferenceRepository) GetByNormalizedName(ctx context.Context, normalizedName string) (*domain.UniversityReference, error) {
	query := `
		SELECT institution_name, normalized_name, education_score, category, updated_at
		FROM university_references
		WHERE normalized_name = $1
	`
	row := r.pool.QueryRow(ctx, query, normalizedName)
	return scanUniversityReference(row)
}

func (r *UniversityReferenceRepository) ListAll(ctx context.Context) ([]domain.UniversityReference, error) {
	query := `
		SELECT institution_name, normalized_name, education_score, category, updated_at
		FROM university_references
		ORDER BY institution_name
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list university references: %w", err)
	}
	defer rows.Close()

	var results []domain.UniversityReference
	for rows.Next() {
		var u domain.UniversityReference
		err := rows.Scan(
			&u.InstitutionName,
			&u.NormalizedName,
			&u.EducationScore,
			&u.Category,
			&u.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan university reference row: %w", err)
		}
		results = append(results, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate university references: %w", err)
	}
	return results, nil
}

func scanUniversityReference(row pgx.Row) (*domain.UniversityReference, error) {
	var u domain.UniversityReference
	err := row.Scan(
		&u.InstitutionName,
		&u.NormalizedName,
		&u.EducationScore,
		&u.Category,
		&u.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("university reference not found")
		}
		return nil, fmt.Errorf("scan university reference: %w", err)
	}
	return &u, nil
}
