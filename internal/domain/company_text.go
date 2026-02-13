package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CompanyText struct {
	CompanyID    uuid.UUID
	Lang         string
	Name         string
	ActivityType string
	CompanyType  string
	About        string
	Market       string
	IsSource     bool
	ModelVersion string
	UpdatedAt    time.Time
}

type CompanyTextRepository interface {
	Create(ctx context.Context, ct *CompanyText) (*CompanyText, error)
	Get(ctx context.Context, companyID uuid.UUID, lang string) (*CompanyText, error)
	ListByCompany(ctx context.Context, companyID uuid.UUID) ([]CompanyText, error)
	Update(ctx context.Context, ct *CompanyText) (*CompanyText, error)
	Delete(ctx context.Context, companyID uuid.UUID, lang string) error
	DeleteByCompany(ctx context.Context, companyID uuid.UUID) error
}
