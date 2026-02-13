package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type VacancyText struct {
	VacancyID        uuid.UUID
	Lang             string
	Title            string
	Description      string
	Responsibilities string
	Requirements     string
	Benefits         string
	IsSource         bool
	ModelVersion     string
	UpdatedAt        time.Time
}

type VacancyTextRepository interface {
	Create(ctx context.Context, vt *VacancyText) (*VacancyText, error)
	Get(ctx context.Context, vacancyID uuid.UUID, lang string) (*VacancyText, error)
	ListByVacancy(ctx context.Context, vacancyID uuid.UUID) ([]VacancyText, error)
	Update(ctx context.Context, vt *VacancyText) (*VacancyText, error)
	Delete(ctx context.Context, vacancyID uuid.UUID, lang string) error
	DeleteByVacancy(ctx context.Context, vacancyID uuid.UUID) error
}
