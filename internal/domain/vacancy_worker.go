package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type VacancyWorker struct {
	ID              uuid.UUID
	VacancyID       uuid.UUID
	UserID          uuid.UUID
	MatchPercentage int
	MatchScore      float64
	Rank            int
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type VacancyWorkerRepository interface {
	Create(ctx context.Context, worker *VacancyWorker) error
	BulkCreate(ctx context.Context, workers []VacancyWorker) error
	ListByVacancy(ctx context.Context, vacancyID uuid.UUID) ([]VacancyWorker, error)
	DeleteByVacancy(ctx context.Context, vacancyID uuid.UUID) error
}
