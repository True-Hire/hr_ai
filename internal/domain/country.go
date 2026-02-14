package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrCountryNotFound = errors.New("country not found")

type Country struct {
	ID        uuid.UUID
	Name      string
	ShortCode string
	CreatedAt time.Time
}

type CountryText struct {
	CountryID    uuid.UUID
	Lang         string
	Name         string
	IsSource     bool
	ModelVersion string
	UpdatedAt    time.Time
}

type CountryRepository interface {
	Create(ctx context.Context, country *Country) (*Country, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Country, error)
	GetByShortCode(ctx context.Context, shortCode string) (*Country, error)
	List(ctx context.Context) ([]Country, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

type CountryTextRepository interface {
	Create(ctx context.Context, text *CountryText) (*CountryText, error)
	Get(ctx context.Context, countryID uuid.UUID, lang string) (*CountryText, error)
	ListByCountry(ctx context.Context, countryID uuid.UUID) ([]CountryText, error)
	DeleteByCountry(ctx context.Context, countryID uuid.UUID) error
}
