package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrHRSessionNotFound = errors.New("hr session not found")

type HRSession struct {
	ID               uuid.UUID
	HRID             uuid.UUID
	DeviceID         string
	RefreshTokenHash string
	FcmToken         string
	IPAddress        string
	CreatedAt        time.Time
	Deleted          bool
}

type HRSessionRepository interface {
	Create(ctx context.Context, session *HRSession) (*HRSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*HRSession, error)
	GetByDeviceID(ctx context.Context, hrID uuid.UUID, deviceID string) (*HRSession, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	SoftDeleteByHR(ctx context.Context, hrID uuid.UUID) error
	HardDeleteByHR(ctx context.Context, hrID uuid.UUID) error
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, hash string) error
}
