package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	// Auto-migrate missing category columns
	_, _ = pool.Exec(ctx, `
		ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS main_category_id UUID REFERENCES main_category(id) ON DELETE SET NULL;
		ALTER TABLE vacancies ADD COLUMN IF NOT EXISTS sub_category_id UUID REFERENCES sub_category(id) ON DELETE SET NULL;
	`)

	return pool, nil
}
