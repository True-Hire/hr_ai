package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrVacancyApplicationNotFound = errors.New("vacancy application not found")
	ErrAlreadyApplied             = errors.New("user already applied to this vacancy")
)

type VacancyApplication struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	VacancyID   uuid.UUID
	Status      string
	CoverLetter string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	SeenAt      *time.Time
}

type VacancyApplicationRepository interface {
	Create(ctx context.Context, va *VacancyApplication) (*VacancyApplication, error)
	GetByID(ctx context.Context, id uuid.UUID) (*VacancyApplication, error)
	GetByUserAndVacancy(ctx context.Context, userID, vacancyID uuid.UUID) (*VacancyApplication, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]VacancyApplication, error)
	ListByVacancy(ctx context.Context, vacancyID uuid.UUID, limit, offset int32) ([]VacancyApplication, error)
	CountByUser(ctx context.Context, userID uuid.UUID) (int64, error)
	CountByVacancy(ctx context.Context, vacancyID uuid.UUID) (int64, error)
	CountUnseenByVacancy(ctx context.Context, vacancyID uuid.UUID) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*VacancyApplication, error)
	MarkSeen(ctx context.Context, id uuid.UUID) (*VacancyApplication, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	DeleteByVacancy(ctx context.Context, vacancyID uuid.UUID) error
}
