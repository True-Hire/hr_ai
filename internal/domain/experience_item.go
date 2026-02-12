package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrExperienceItemNotFound = errors.New("experience item not found")

type ExperienceItem struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Company   string
	Position  string
	StartDate string
	EndDate   string
	Projects  string
	WebSite   string
	ItemOrder int32
	UpdatedAt time.Time
}

type ExperienceItemRepository interface {
	Create(ctx context.Context, item *ExperienceItem) (*ExperienceItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ExperienceItem, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]ExperienceItem, error)
	Update(ctx context.Context, item *ExperienceItem) (*ExperienceItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
}
