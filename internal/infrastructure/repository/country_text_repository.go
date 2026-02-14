package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	countrytextsdb "github.com/ruziba3vich/hr-ai/db/sqlc/country_texts"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CountryTextRepository struct {
	q *countrytextsdb.Queries
}

func NewCountryTextRepository(pool *pgxpool.Pool) *CountryTextRepository {
	return &CountryTextRepository{q: countrytextsdb.New(pool)}
}

func (r *CountryTextRepository) Create(ctx context.Context, ct *domain.CountryText) (*domain.CountryText, error) {
	row, err := r.q.CreateCountryText(ctx, countrytextsdb.CreateCountryTextParams{
		CountryID:    uuidToPgtype(ct.CountryID),
		Lang:         ct.Lang,
		Name:         ct.Name,
		IsSource:     ct.IsSource,
		ModelVersion: textToPgtype(ct.ModelVersion),
	})
	if err != nil {
		return nil, fmt.Errorf("create country text: %w", err)
	}
	return countryTextToDomain(row), nil
}

func (r *CountryTextRepository) Get(ctx context.Context, countryID uuid.UUID, lang string) (*domain.CountryText, error) {
	row, err := r.q.GetCountryText(ctx, countrytextsdb.GetCountryTextParams{
		CountryID: uuidToPgtype(countryID),
		Lang:      lang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCountryNotFound
		}
		return nil, fmt.Errorf("get country text: %w", err)
	}
	return countryTextToDomain(row), nil
}

func (r *CountryTextRepository) ListByCountry(ctx context.Context, countryID uuid.UUID) ([]domain.CountryText, error) {
	rows, err := r.q.ListCountryTextsByCountry(ctx, uuidToPgtype(countryID))
	if err != nil {
		return nil, fmt.Errorf("list country texts: %w", err)
	}
	texts := make([]domain.CountryText, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, *countryTextToDomain(row))
	}
	return texts, nil
}

func (r *CountryTextRepository) DeleteByCountry(ctx context.Context, countryID uuid.UUID) error {
	return r.q.DeleteCountryTextsByCountry(ctx, uuidToPgtype(countryID))
}

func countryTextToDomain(row countrytextsdb.CountryText) *domain.CountryText {
	return &domain.CountryText{
		CountryID:    pgtypeToUUID(row.CountryID),
		Lang:         row.Lang,
		Name:         row.Name,
		IsSource:     row.IsSource,
		ModelVersion: pgtypeToString(row.ModelVersion),
		UpdatedAt:    row.UpdatedAt.Time,
	}
}
