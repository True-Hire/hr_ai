package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type SearchSessionRepository struct {
	pool *pgxpool.Pool
}

func NewSearchSessionRepository(pool *pgxpool.Pool) *SearchSessionRepository {
	return &SearchSessionRepository{pool: pool}
}

func (r *SearchSessionRepository) Create(ctx context.Context, session *domain.SearchSession) (*domain.SearchSession, error) {
	parsedQueryJSON, err := json.Marshal(session.ParsedQuery)
	if err != nil {
		return nil, fmt.Errorf("marshal parsed_query: %w", err)
	}
	filtersJSON, err := json.Marshal(session.Filters)
	if err != nil {
		return nil, fmt.Errorf("marshal filters: %w", err)
	}

	query := `
		INSERT INTO search_sessions (
			id, hr_id, query_text, parsed_query, filters,
			total_results, status, created_at, expires_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, now(), $8
		)
		RETURNING id, hr_id, query_text, parsed_query, filters,
			total_results, status, created_at, expires_at
	`

	row := r.pool.QueryRow(ctx, query,
		session.ID,
		session.HRID,
		session.QueryText,
		parsedQueryJSON,
		filtersJSON,
		session.TotalResults,
		session.Status,
		session.ExpiresAt,
	)

	return scanSearchSession(row)
}

func (r *SearchSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.SearchSession, error) {
	query := `
		SELECT id, hr_id, query_text, parsed_query, filters,
			total_results, status, created_at, expires_at
		FROM search_sessions
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanSearchSession(row)
}

func (r *SearchSessionRepository) InsertResults(ctx context.Context, searchID uuid.UUID, results []domain.SearchSessionResult) error {
	if len(results) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("INSERT INTO search_session_results (search_id, rank, user_id, final_score, score_breakdown) VALUES ")

	args := make([]interface{}, 0, len(results)*5)
	for i, res := range results {
		if i > 0 {
			sb.WriteString(", ")
		}
		base := i*5 + 1
		fmt.Fprintf(&sb, "($%d, $%d, $%d, $%d, $%d)", base, base+1, base+2, base+3, base+4)

		breakdownJSON, err := json.Marshal(res.ScoreBreakdown)
		if err != nil {
			return fmt.Errorf("marshal score_breakdown for rank %d: %w", res.Rank, err)
		}

		args = append(args, searchID, res.Rank, res.UserID, res.FinalScore, breakdownJSON)
	}

	_, err := r.pool.Exec(ctx, sb.String(), args...)
	if err != nil {
		return fmt.Errorf("insert search session results: %w", err)
	}
	return nil
}

func (r *SearchSessionRepository) GetResultsPage(ctx context.Context, searchID uuid.UUID, afterRank int, pageSize int) ([]domain.SearchSessionResult, error) {
	query := `
		SELECT search_id, rank, user_id, final_score, score_breakdown
		FROM search_session_results
		WHERE search_id = $1 AND rank > $2
		ORDER BY rank
		LIMIT $3
	`

	rows, err := r.pool.Query(ctx, query, searchID, afterRank, pageSize)
	if err != nil {
		return nil, fmt.Errorf("get results page: %w", err)
	}
	defer rows.Close()

	var results []domain.SearchSessionResult
	for rows.Next() {
		var res domain.SearchSessionResult
		var breakdownJSON []byte

		err := rows.Scan(
			&res.SearchID,
			&res.Rank,
			&res.UserID,
			&res.FinalScore,
			&breakdownJSON,
		)
		if err != nil {
			return nil, fmt.Errorf("scan search session result row: %w", err)
		}

		if len(breakdownJSON) > 0 {
			if err := json.Unmarshal(breakdownJSON, &res.ScoreBreakdown); err != nil {
				return nil, fmt.Errorf("unmarshal score_breakdown: %w", err)
			}
		}

		results = append(results, res)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate search session results: %w", err)
	}
	return results, nil
}

func (r *SearchSessionRepository) DeleteExpired(ctx context.Context) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM search_sessions WHERE expires_at < now()")
	if err != nil {
		return fmt.Errorf("delete expired search sessions: %w", err)
	}
	return nil
}

func scanSearchSession(row pgx.Row) (*domain.SearchSession, error) {
	var s domain.SearchSession
	var parsedQueryJSON, filtersJSON []byte

	err := row.Scan(
		&s.ID,
		&s.HRID,
		&s.QueryText,
		&parsedQueryJSON,
		&filtersJSON,
		&s.TotalResults,
		&s.Status,
		&s.CreatedAt,
		&s.ExpiresAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("search session not found")
		}
		return nil, fmt.Errorf("scan search session: %w", err)
	}

	if len(parsedQueryJSON) > 0 {
		if err := json.Unmarshal(parsedQueryJSON, &s.ParsedQuery); err != nil {
			return nil, fmt.Errorf("unmarshal parsed_query: %w", err)
		}
	}
	if len(filtersJSON) > 0 {
		if err := json.Unmarshal(filtersJSON, &s.Filters); err != nil {
			return nil, fmt.Errorf("unmarshal filters: %w", err)
		}
	}

	return &s, nil
}
