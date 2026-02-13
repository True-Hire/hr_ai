package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrSessionNotFound    = errors.New("session not found")
)

type UserSession struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	DeviceID         string
	RefreshTokenHash string
	FcmToken         string
	IPAddress        string
	CreatedAt        time.Time
	Deleted          bool
}

type SessionRepository interface {
	Create(ctx context.Context, session *UserSession) (*UserSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*UserSession, error)
	GetByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) (*UserSession, error)
	SoftDelete(ctx context.Context, id uuid.UUID) error
	SoftDeleteByUser(ctx context.Context, userID uuid.UUID) error
	UpdateRefreshToken(ctx context.Context, id uuid.UUID, hash string) error
}
