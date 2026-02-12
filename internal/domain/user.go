package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID            uuid.UUID
	Phone         string
	Email         string
	ProfilePicURL string
	CreatedAt     time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, limit, offset int32) ([]User, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
