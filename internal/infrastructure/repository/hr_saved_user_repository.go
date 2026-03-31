package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	sudb "github.com/ruziba3vich/hr-ai/db/sqlc/hr_saved_users"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type HRSavedUserRepository struct {
	q    *sudb.Queries
	pool *pgxpool.Pool
}

func NewHRSavedUserRepository(pool *pgxpool.Pool) *HRSavedUserRepository {
	return &HRSavedUserRepository{q: sudb.New(pool), pool: pool}
}

func (r *HRSavedUserRepository) Save(ctx context.Context, hrID, userID uuid.UUID, note string) (*domain.HRSavedUser, error) {
	row, err := r.q.SaveUser(ctx, sudb.SaveUserParams{
		HrID:   uuidToPgtype(hrID),
		UserID: uuidToPgtype(userID),
		Note:   textToPgtype(note),
	})
	if err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return savedUserFromRow(row), nil
}

func (r *HRSavedUserRepository) Unsave(ctx context.Context, hrID, userID uuid.UUID) error {
	return r.q.UnsaveUser(ctx, sudb.UnsaveUserParams{
		HrID:   uuidToPgtype(hrID),
		UserID: uuidToPgtype(userID),
	})
}

func (r *HRSavedUserRepository) IsSaved(ctx context.Context, hrID, userID uuid.UUID) (bool, error) {
	count, err := r.q.IsSaved(ctx, sudb.IsSavedParams{
		HrID:   uuidToPgtype(hrID),
		UserID: uuidToPgtype(userID),
	})
	if err != nil {
		return false, fmt.Errorf("check saved: %w", err)
	}
	return count > 0, nil
}

func (r *HRSavedUserRepository) ListByHR(ctx context.Context, hrID uuid.UUID, limit, offset int32) ([]domain.HRSavedUser, error) {
	rows, err := r.q.ListSavedByHR(ctx, sudb.ListSavedByHRParams{
		HrID:   uuidToPgtype(hrID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list saved users: %w", err)
	}
	result := make([]domain.HRSavedUser, 0, len(rows))
	for _, row := range rows {
		result = append(result, *savedUserFromRow(row))
	}
	return result, nil
}

func (r *HRSavedUserRepository) CountByHR(ctx context.Context, hrID uuid.UUID) (int64, error) {
	return r.q.CountSavedByHR(ctx, uuidToPgtype(hrID))
}

// ListByHRFiltered returns saved users with dynamic filtering on user fields and skills.
func (r *HRSavedUserRepository) ListByHRFiltered(ctx context.Context, hrID uuid.UUID, nameQuery string, skills []string, limit, offset int32) ([]domain.HRSavedUser, int64, error) {
	query := `
		SELECT s.hr_id, s.user_id, s.note, s.created_at
		FROM hr_saved_users s
		JOIN users u ON u.id = s.user_id
		WHERE s.hr_id = $1`
	countQuery := `
		SELECT count(*)
		FROM hr_saved_users s
		JOIN users u ON u.id = s.user_id
		WHERE s.hr_id = $1`

	args := []interface{}{uuidToPgtype(hrID)}
	argIdx := 2

	if nameQuery != "" {
		filter := fmt.Sprintf(` AND (u.first_name ILIKE $%d OR u.last_name ILIKE $%d)`, argIdx, argIdx)
		query += filter
		countQuery += filter
		args = append(args, "%"+nameQuery+"%")
		argIdx++
	}

	if len(skills) > 0 {
		filter := fmt.Sprintf(` AND EXISTS (
			SELECT 1 FROM user_skills us
			JOIN skills sk ON sk.id = us.skill_id
			WHERE us.user_id = u.id AND sk.name ILIKE ANY($%d)
		)`, argIdx)
		query += filter
		countQuery += filter
		patterns := make([]string, len(skills))
		for i, s := range skills {
			patterns[i] = "%" + s + "%"
		}
		args = append(args, patterns)
		argIdx++
	}

	// Get count
	var total int64
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count filtered saved users: %w", err)
	}

	query += fmt.Sprintf(` ORDER BY s.created_at DESC LIMIT $%d OFFSET $%d`, argIdx, argIdx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list filtered saved users: %w", err)
	}
	defer rows.Close()

	var result []domain.HRSavedUser
	for rows.Next() {
		var row sudb.HrSavedUser
		if err := rows.Scan(&row.HrID, &row.UserID, &row.Note, &row.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan saved user row: %w", err)
		}
		result = append(result, *savedUserFromRow(row))
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate saved user rows: %w", err)
	}

	return result, total, nil
}

func savedUserFromRow(row sudb.HrSavedUser) *domain.HRSavedUser {
	return &domain.HRSavedUser{
		HRID:      pgtypeToUUID(row.HrID),
		UserID:    pgtypeToUUID(row.UserID),
		Note:      pgtypeToString(row.Note),
		CreatedAt: row.CreatedAt.Time,
	}
}
