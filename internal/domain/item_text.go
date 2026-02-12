package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrItemTextNotFound = errors.New("item text not found")

type ItemText struct {
	ItemID       uuid.UUID
	ItemType     string
	Lang         string
	Description  string
	IsSource     bool
	ModelVersion string
	UpdatedAt    time.Time
}

type ItemTextRepository interface {
	Create(ctx context.Context, text *ItemText) (*ItemText, error)
	Get(ctx context.Context, itemID uuid.UUID, itemType string, lang string) (*ItemText, error)
	ListByItem(ctx context.Context, itemID uuid.UUID, itemType string) ([]ItemText, error)
	Update(ctx context.Context, text *ItemText) (*ItemText, error)
	Delete(ctx context.Context, itemID uuid.UUID, itemType string, lang string) error
	DeleteByItem(ctx context.Context, itemID uuid.UUID, itemType string) error
	DeleteByItemID(ctx context.Context, itemID uuid.UUID) error
}
