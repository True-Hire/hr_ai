package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrProfileFieldNotFound = errors.New("profile field not found")

type ProfileField struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	FieldName  string
	SourceLang string
	UpdatedAt  time.Time
}

type ProfileFieldRepository interface {
	Create(ctx context.Context, field *ProfileField) (*ProfileField, error)
	GetByID(ctx context.Context, id uuid.UUID) (*ProfileField, error)
	GetByUserAndName(ctx context.Context, userID uuid.UUID, fieldName string) (*ProfileField, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]ProfileField, error)
	Update(ctx context.Context, field *ProfileField) (*ProfileField, error)
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
}
