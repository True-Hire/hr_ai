package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	countriesdb "github.com/ruziba3vich/hr-ai/db/sqlc/countries"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CountryRepository struct {
	q *countriesdb.Queries
}

func NewCountryRepository(pool *pgxpool.Pool) *CountryRepository {
	return &CountryRepository{q: countriesdb.New(pool)}
}

func (r *CountryRepository) Create(ctx context.Context, c *domain.Country) (*domain.Country, error) {
	row, err := r.q.CreateCountry(ctx, countriesdb.CreateCountryParams{
		ID:        uuidToPgtype(c.ID),
		Name:      c.Name,
		ShortCode: c.ShortCode,
	})
	if err != nil {
		return nil, fmt.Errorf("create country: %w", err)
	}
	return countryToDomain(row), nil
}

func (r *CountryRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Country, error) {
	row, err := r.q.GetCountryByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCountryNotFound
		}
		return nil, fmt.Errorf("get country by id: %w", err)
	}
	return countryToDomain(row), nil
}

func (r *CountryRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.Country, error) {
	row, err := r.q.GetCountryByShortCode(ctx, shortCode)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCountryNotFound
		}
		return nil, fmt.Errorf("get country by short code: %w", err)
	}
	return countryToDomain(row), nil
}

func (r *CountryRepository) List(ctx context.Context) ([]domain.Country, error) {
	rows, err := r.q.ListCountries(ctx)
	if err != nil {
		return nil, fmt.Errorf("list countries: %w", err)
	}
	result := make([]domain.Country, 0, len(rows))
	for _, row := range rows {
		result = append(result, *countryToDomain(row))
	}
	return result, nil
}

func (r *CountryRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteCountry(ctx, uuidToPgtype(id))
}

func countryToDomain(row countriesdb.Country) *domain.Country {
	return &domain.Country{
		ID:        pgtypeToUUID(row.ID),
		Name:      row.Name,
		ShortCode: row.ShortCode,
		CreatedAt: row.CreatedAt.Time,
	}
}
