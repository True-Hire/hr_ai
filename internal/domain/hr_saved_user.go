package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type HRSavedUser struct {
	HRID      uuid.UUID
	UserID    uuid.UUID
	Note      string
	CreatedAt time.Time
}

type HRSavedUserRepository interface {
	Save(ctx context.Context, hrID, userID uuid.UUID, note string) (*HRSavedUser, error)
	Unsave(ctx context.Context, hrID, userID uuid.UUID) error
	IsSaved(ctx context.Context, hrID, userID uuid.UUID) (bool, error)
	ListByHR(ctx context.Context, hrID uuid.UUID, limit, offset int32) ([]HRSavedUser, error)
	CountByHR(ctx context.Context, hrID uuid.UUID) (int64, error)
}
