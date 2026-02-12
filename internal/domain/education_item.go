package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrEducationItemNotFound = errors.New("education item not found")

type EducationItem struct {
	ID           uuid.UUID
	UserID       uuid.UUID
	Institution  string
	Degree       string
	FieldOfStudy string
	StartDate    string
	EndDate      string
	Location     string
	ItemOrder    int32
	UpdatedAt    time.Time
}

type EducationItemRepository interface {
	Create(ctx context.Context, item *EducationItem) (*EducationItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (*EducationItem, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]EducationItem, error)
	Update(ctx context.Context, item *EducationItem) (*EducationItem, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
}
