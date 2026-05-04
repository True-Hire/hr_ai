package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrVacancyNotFound = errors.New("vacancy not found")

const (
	VacancyStatusActive = "active"
	VacancyStatusDraft  = "draft"
)

type Vacancy struct {
	ID             uuid.UUID
	HRID           uuid.UUID
	CompanyData    *CompanyData
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
	MainCategoryID uuid.UUID
	SubCategoryID  uuid.UUID
	CreatedAt      time.Time
}

// VacancyFilter contains optional filter fields for listing vacancies.
// Zero values are ignored (not filtered on).
type VacancyFilter struct {
	HRID           uuid.UUID
	Status         string
	Format         string
	Schedule       string
	SalaryCurrency string
	SalaryMin      int32 // vacancies with salary_max >= this value
	SalaryMax      int32 // vacancies with salary_min <= this value
	ExperienceMin  int32 // vacancies with experience_max >= this value
	ExperienceMax  int32 // vacancies with experience_min <= this value
	CountryID      uuid.UUID
}

type VacancyRepository interface {
	Create(ctx context.Context, v *Vacancy) (*Vacancy, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Vacancy, error)
	List(ctx context.Context, limit, offset int32) ([]Vacancy, error)
	ListByHR(ctx context.Context, hrID uuid.UUID, limit, offset int32) ([]Vacancy, error)
	ListFiltered(ctx context.Context, filter VacancyFilter, limit, offset int32) ([]Vacancy, error)
	CountFiltered(ctx context.Context, filter VacancyFilter) (int64, error)
	Search(ctx context.Context, lang, query string, limit, offset int32) ([]Vacancy, error)
	Count(ctx context.Context) (int64, error)
	CountByHR(ctx context.Context, hrID uuid.UUID) (int64, error)
	CountSearch(ctx context.Context, lang, query string) (int64, error)
	Update(ctx context.Context, v *Vacancy) (*Vacancy, error)
	Delete(ctx context.Context, id uuid.UUID) error
	NullifyHRID(ctx context.Context, hrID uuid.UUID) error
	ListIDsByHR(ctx context.Context, hrID uuid.UUID) ([]uuid.UUID, error)
}
