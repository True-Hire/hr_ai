package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrVacancyNotFound = errors.New("vacancy not found")

type Vacancy struct {
	ID             uuid.UUID
	HRID           uuid.UUID
	CompanyID      uuid.UUID
	CountryID      uuid.UUID
	SalaryMin      int32
	SalaryMax      int32
	SalaryCurrency string
	ExperienceMin  int32
	ExperienceMax  int32
	Format         string
	Schedule       string
	Phone          string
	Telegram       string
	Email          string
	Address        string
	Status         string
	SourceLang     string
	CreatedAt      time.Time
}

type VacancyRepository interface {
	Create(ctx context.Context, v *Vacancy) (*Vacancy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Vacancy, error)
	List(ctx context.Context, limit, offset int32) ([]Vacancy, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int32) ([]Vacancy, error)
	ListByHR(ctx context.Context, hrID uuid.UUID, limit, offset int32) ([]Vacancy, error)
	Search(ctx context.Context, lang, query string, limit, offset int32) ([]Vacancy, error)
	Count(ctx context.Context) (int64, error)
	CountByCompany(ctx context.Context, companyID uuid.UUID) (int64, error)
	CountSearch(ctx context.Context, lang, query string) (int64, error)
	Update(ctx context.Context, v *Vacancy) (*Vacancy, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
