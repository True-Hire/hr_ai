package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type SearchSession struct {
	ID           uuid.UUID
	HRID         uuid.UUID
	QueryText    string
	ParsedQuery  map[string]interface{}
	Filters      map[string]interface{}
	TotalResults int
	Status       string
	CreatedAt    time.Time
	ExpiresAt    time.Time
}

type SearchSessionResult struct {
	SearchID       uuid.UUID
	Rank           int
	UserID         uuid.UUID
	FinalScore     float64
	ScoreBreakdown map[string]interface{}
}

type SearchSessionRepository interface {
	Create(ctx context.Context, session *SearchSession) (*SearchSession, error)
	GetByID(ctx context.Context, id uuid.UUID) (*SearchSession, error)
	InsertResults(ctx context.Context, searchID uuid.UUID, results []SearchSessionResult) error
	GetResultsPage(ctx context.Context, searchID uuid.UUID, afterRank int, pageSize int) ([]SearchSessionResult, error)
	DeleteExpired(ctx context.Context) error
}
