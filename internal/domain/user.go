package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrUserNotFound = errors.New("user not found")

type User struct {
	ID                      uuid.UUID
	FirstName               string
	LastName                string
	Patronymic              string
	Phone                   string
	Telegram                string
	TelegramID              string
	Email                   string
	Gender                  string
	Country                 string
	Region                  string
	Nationality             string
	ProfilePicURL           string
	Status                  string
	TariffType              string
	JobStatus               string
	ActivityType            string
	Specializations         []string
	PasswordHash            string
	Language                string
	ProfileScore            int32
	EstimatedSalaryMin      int32
	EstimatedSalaryMax      int32
	EstimatedSalaryCurrency string
	MainCategoryID          uuid.UUID
	SubCategoryID           uuid.UUID
	CreatedAt               time.Time
}

type UserRepository interface {
	Create(ctx context.Context, user *User) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	List(ctx context.Context, limit, offset int32) ([]User, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, user *User) (*User, error)
	Delete(ctx context.Context, id uuid.UUID) error
	GetByPhone(ctx context.Context, phone string) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	GetByTelegramID(ctx context.Context, telegramID string) (*User, error)
	CountMatchingUsers(ctx context.Context, mainCatID, subCatID uuid.UUID) (int64, error)
	SetPassword(ctx context.Context, id uuid.UUID, hash string) error
	SetProfileScore(ctx context.Context, id uuid.UUID, score int32) error
	SetEstimatedSalary(ctx context.Context, id uuid.UUID, min, max int32, currency string) error
}
