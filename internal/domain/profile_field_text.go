package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrProfileFieldTextNotFound = errors.New("profile field text not found")

type ProfileFieldText struct {
	ProfileFieldID uuid.UUID
	Lang           string
	Content        string
	IsSource       bool
	ModelVersion   string
	UpdatedAt      time.Time
}

type ProfileFieldTextRepository interface {
	Create(ctx context.Context, text *ProfileFieldText) (*ProfileFieldText, error)
	Get(ctx context.Context, profileFieldID uuid.UUID, lang string) (*ProfileFieldText, error)
	ListByField(ctx context.Context, profileFieldID uuid.UUID) ([]ProfileFieldText, error)
	Update(ctx context.Context, text *ProfileFieldText) (*ProfileFieldText, error)
	Delete(ctx context.Context, profileFieldID uuid.UUID, lang string) error
	DeleteByField(ctx context.Context, profileFieldID uuid.UUID) error
}
