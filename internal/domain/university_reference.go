package domain

import (
	"context"
	"time"
)

type UniversityReference struct {
	InstitutionName string
	NormalizedName  string
	EducationScore  float64
	Category        string
	UpdatedAt       time.Time
}

type UniversityReferenceRepository interface {
	GetByName(ctx context.Context, name string) (*UniversityReference, error)
	GetByNormalizedName(ctx context.Context, normalizedName string) (*UniversityReference, error)
	ListAll(ctx context.Context) ([]UniversityReference, error)
}
