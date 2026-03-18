package domain

import (
	"context"
	"time"
)

type CompanyReference struct {
	CompanyName      string
	NormalizedName   string
	PrestigeScore    float64
	EngineeringScore float64
	ScaleScore       float64
	HiringBarScore   float64
	Category         string
	UpdatedAt        time.Time
}

type CompanyReferenceRepository interface {
	GetByName(ctx context.Context, name string) (*CompanyReference, error)
	GetByNormalizedName(ctx context.Context, normalizedName string) (*CompanyReference, error)
	ListAll(ctx context.Context) ([]CompanyReference, error)
}
