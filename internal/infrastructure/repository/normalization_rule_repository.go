package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type NormalizationRuleRepository struct {
	pool *pgxpool.Pool
}

func NewNormalizationRuleRepository(pool *pgxpool.Pool) *NormalizationRuleRepository {
	return &NormalizationRuleRepository{pool: pool}
}

func (r *NormalizationRuleRepository) Create(ctx context.Context, rule *domain.NormalizationRule) (*domain.NormalizationRule, error) {
	metadataJSON, err := json.Marshal(rule.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO normalization_rules (id, category, source_value, normalized_value, metadata)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, category, source_value, normalized_value, metadata, created_at
	`
	row := r.pool.QueryRow(ctx, query, rule.ID, rule.Category, rule.SourceValue, rule.NormalizedValue, metadataJSON)
	return scanNormalizationRule(row)
}

func (r *NormalizationRuleRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NormalizationRule, error) {
	query := `
		SELECT id, category, source_value, normalized_value, metadata, created_at
		FROM normalization_rules
		WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanNormalizationRule(row)
}

func (r *NormalizationRuleRepository) Update(ctx context.Context, rule *domain.NormalizationRule) (*domain.NormalizationRule, error) {
	metadataJSON, err := json.Marshal(rule.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		UPDATE normalization_rules
		SET category = $2, source_value = $3, normalized_value = $4, metadata = $5
		WHERE id = $1
		RETURNING id, category, source_value, normalized_value, metadata, created_at
	`
	row := r.pool.QueryRow(ctx, query, rule.ID, rule.Category, rule.SourceValue, rule.NormalizedValue, metadataJSON)
	return scanNormalizationRule(row)
}

func (r *NormalizationRuleRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM normalization_rules WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete normalization rule: %w", err)
	}
	return nil
}

func (r *NormalizationRuleRepository) ListAll(ctx context.Context) ([]domain.NormalizationRule, error) {
	query := `
		SELECT id, category, source_value, normalized_value, metadata, created_at
		FROM normalization_rules
		ORDER BY category, source_value
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("list normalization rules: %w", err)
	}
	defer rows.Close()

	return scanNormalizationRules(rows)
}

func (r *NormalizationRuleRepository) ListByCategory(ctx context.Context, category string) ([]domain.NormalizationRule, error) {
	query := `
		SELECT id, category, source_value, normalized_value, metadata, created_at
		FROM normalization_rules
		WHERE category = $1
		ORDER BY source_value
	`
	rows, err := r.pool.Query(ctx, query, category)
	if err != nil {
		return nil, fmt.Errorf("list normalization rules by category: %w", err)
	}
	defer rows.Close()

	return scanNormalizationRules(rows)
}

func (r *NormalizationRuleRepository) Search(ctx context.Context, category, query string, limit, offset int) ([]domain.NormalizationRule, error) {
	var sqlQuery string
	var args []any

	if query == "" {
		sqlQuery = `
			SELECT id, category, source_value, normalized_value, metadata, created_at
			FROM normalization_rules
			WHERE (category = $1 OR $1 = '')
			ORDER BY category, source_value
			LIMIT $2 OFFSET $3
		`
		args = []any{category, limit, offset}
	} else {
		sqlQuery = `
			SELECT id, category, source_value, normalized_value, metadata, created_at
			FROM normalization_rules
			WHERE (category = $1 OR $1 = '')
				AND (source_value ILIKE '%' || $2 || '%' OR normalized_value ILIKE '%' || $2 || '%')
			ORDER BY category, source_value
			LIMIT $3 OFFSET $4
		`
		args = []any{category, query, limit, offset}
	}

	rows, err := r.pool.Query(ctx, sqlQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("search normalization rules: %w", err)
	}
	defer rows.Close()

	return scanNormalizationRules(rows)
}

func (r *NormalizationRuleRepository) Count(ctx context.Context, category, query string) (int64, error) {
	var sqlQuery string
	var args []any

	if query == "" {
		sqlQuery = `
			SELECT COUNT(*)
			FROM normalization_rules
			WHERE (category = $1 OR $1 = '')
		`
		args = []any{category}
	} else {
		sqlQuery = `
			SELECT COUNT(*)
			FROM normalization_rules
			WHERE (category = $1 OR $1 = '')
				AND (source_value ILIKE '%' || $2 || '%' OR normalized_value ILIKE '%' || $2 || '%')
		`
		args = []any{category, query}
	}

	var count int64
	err := r.pool.QueryRow(ctx, sqlQuery, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("count normalization rules: %w", err)
	}
	return count, nil
}

func (r *NormalizationRuleRepository) GetByCategoryAndSource(ctx context.Context, category, sourceValue string) (*domain.NormalizationRule, error) {
	query := `
		SELECT id, category, source_value, normalized_value, metadata, created_at
		FROM normalization_rules
		WHERE category = $1 AND source_value = $2
	`
	row := r.pool.QueryRow(ctx, query, category, sourceValue)
	return scanNormalizationRule(row)
}

func (r *NormalizationRuleRepository) Upsert(ctx context.Context, rule *domain.NormalizationRule) (*domain.NormalizationRule, error) {
	metadataJSON, err := json.Marshal(rule.Metadata)
	if err != nil {
		return nil, fmt.Errorf("marshal metadata: %w", err)
	}

	query := `
		INSERT INTO normalization_rules (id, category, source_value, normalized_value, metadata)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (category, source_value)
		DO UPDATE SET normalized_value = EXCLUDED.normalized_value, metadata = EXCLUDED.metadata
		RETURNING id, category, source_value, normalized_value, metadata, created_at
	`
	row := r.pool.QueryRow(ctx, query, rule.ID, rule.Category, rule.SourceValue, rule.NormalizedValue, metadataJSON)
	return scanNormalizationRule(row)
}

func scanNormalizationRule(row pgx.Row) (*domain.NormalizationRule, error) {
	var n domain.NormalizationRule
	var metadataJSON []byte
	err := row.Scan(
		&n.ID,
		&n.Category,
		&n.SourceValue,
		&n.NormalizedValue,
		&metadataJSON,
		&n.CreatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("normalization rule not found")
		}
		return nil, fmt.Errorf("scan normalization rule: %w", err)
	}
	if metadataJSON != nil {
		if err := json.Unmarshal(metadataJSON, &n.Metadata); err != nil {
			return nil, fmt.Errorf("unmarshal metadata: %w", err)
		}
	}
	return &n, nil
}

func scanNormalizationRules(rows pgx.Rows) ([]domain.NormalizationRule, error) {
	var results []domain.NormalizationRule
	for rows.Next() {
		var n domain.NormalizationRule
		var metadataJSON []byte
		err := rows.Scan(
			&n.ID,
			&n.Category,
			&n.SourceValue,
			&n.NormalizedValue,
			&metadataJSON,
			&n.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("scan normalization rule row: %w", err)
		}
		if metadataJSON != nil {
			if err := json.Unmarshal(metadataJSON, &n.Metadata); err != nil {
				return nil, fmt.Errorf("unmarshal metadata: %w", err)
			}
		}
		results = append(results, n)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate normalization rules: %w", err)
	}
	return results, nil
}
