package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var ErrCompanyNotFound = errors.New("company not found")

type Company struct {
	ID               uuid.UUID
	EmployeeCount    int32
	Country          string
	Address          string
	Phone            string
	Telegram         string
	TelegramChannel  string
	Email            string
	LogoURL          string
	WebSite          string
	Instagram        string
	SourceLang       string
	CreatedAt        time.Time
}

type CompanyRepository interface {
	Create(ctx context.Context, company *Company) (*Company, error)
	GetByID(ctx context.Context, id uuid.UUID) (*Company, error)
	List(ctx context.Context, limit, offset int32) ([]Company, error)
	Count(ctx context.Context) (int64, error)
	Update(ctx context.Context, company *Company) (*Company, error)
	Delete(ctx context.Context, id uuid.UUID) error
}
