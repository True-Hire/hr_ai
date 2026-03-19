package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type NormalizationRule struct {
	ID              uuid.UUID
	Category        string
	SourceValue     string
	NormalizedValue string
	Metadata        map[string]any
	CreatedAt       time.Time
}

type NormalizationRuleRepository interface {
	Create(ctx context.Context, rule *NormalizationRule) (*NormalizationRule, error)
	GetByID(ctx context.Context, id uuid.UUID) (*NormalizationRule, error)
	Update(ctx context.Context, rule *NormalizationRule) (*NormalizationRule, error)
	Delete(ctx context.Context, id uuid.UUID) error
	ListAll(ctx context.Context) ([]NormalizationRule, error)
	ListByCategory(ctx context.Context, category string) ([]NormalizationRule, error)
	Search(ctx context.Context, category, query string, limit, offset int) ([]NormalizationRule, error)
	Count(ctx context.Context, category, query string) (int64, error)
	GetByCategoryAndSource(ctx context.Context, category, sourceValue string) (*NormalizationRule, error)
	Upsert(ctx context.Context, rule *NormalizationRule) (*NormalizationRule, error)
}
