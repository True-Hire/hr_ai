package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrCompanyHRNotFound = errors.New("company hr not found")

type CompanyHR struct {
	ID           uuid.UUID
	FirstName    string
	LastName     string
	Patronymic   string
	Phone        string
	Telegram     string
	TelegramID   string
	Email        string
	Position     string
	Status       string
	PasswordHash string
	CompanyID    uuid.UUID
	CreatedAt    time.Time
}

type CompanyHRRepository interface {
	Create(ctx context.Context, hr *CompanyHR) (*CompanyHR, error)
	GetByID(ctx context.Context, id uuid.UUID) (*CompanyHR, error)
	List(ctx context.Context, limit, offset int32) ([]CompanyHR, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, hr *CompanyHR) (*CompanyHR, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByPhone(ctx context.Context, phone string) (*CompanyHR, error)
	GetByEmail(ctx context.Context, email string) (*CompanyHR, error)
	SetPassword(ctx context.Context, id uuid.UUID, hash string) error
}
