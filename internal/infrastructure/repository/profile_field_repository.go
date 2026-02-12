package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pfdb "github.com/ruziba3vich/hr-ai/db/sqlc/profile_fields"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ProfileFieldRepository struct {
	q *pfdb.Queries
}

func NewProfileFieldRepository(pool *pgxpool.Pool) *ProfileFieldRepository {
	return &ProfileFieldRepository{
		q: pfdb.New(pool),
	}
}

func (r *ProfileFieldRepository) Create(ctx context.Context, field *domain.ProfileField) (*domain.ProfileField, error) {
	row, err := r.q.CreateProfileField(ctx, pfdb.CreateProfileFieldParams{
		ID:         uuidToPgtype(field.ID),
		UserID:     uuidToPgtype(field.UserID),
		FieldName:  field.FieldName,
		SourceLang: field.SourceLang,
	})
	if err != nil {
		return nil, fmt.Errorf("create profile field: %w", err)
	}
	return profileFieldToDomain(row), nil
}

func (r *ProfileFieldRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.ProfileField, error) {
	row, err := r.q.GetProfileFieldByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileFieldNotFound
		}
		return nil, fmt.Errorf("get profile field by id: %w", err)
	}
	return profileFieldToDomain(row), nil
}

func (r *ProfileFieldRepository) GetByUserAndName(ctx context.Context, userID uuid.UUID, fieldName string) (*domain.ProfileField, error) {
	row, err := r.q.GetProfileFieldByUserAndName(ctx, pfdb.GetProfileFieldByUserAndNameParams{
		UserID:    uuidToPgtype(userID),
		FieldName: fieldName,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileFieldNotFound
		}
		return nil, fmt.Errorf("get profile field by user and name: %w", err)
	}
	return profileFieldToDomain(row), nil
}

func (r *ProfileFieldRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]domain.ProfileField, error) {
	rows, err := r.q.ListProfileFieldsByUser(ctx, uuidToPgtype(userID))
	if err != nil {
		return nil, fmt.Errorf("list profile fields by user: %w", err)
	}
	fields := make([]domain.ProfileField, 0, len(rows))
	for _, row := range rows {
		fields = append(fields, *profileFieldToDomain(row))
	}
	return fields, nil
}

func (r *ProfileFieldRepository) Update(ctx context.Context, field *domain.ProfileField) (*domain.ProfileField, error) {
	row, err := r.q.UpdateProfileField(ctx, pfdb.UpdateProfileFieldParams{
		ID:         uuidToPgtype(field.ID),
		FieldName:  field.FieldName,
		SourceLang: field.SourceLang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileFieldNotFound
		}
		return nil, fmt.Errorf("update profile field: %w", err)
	}
	return profileFieldToDomain(row), nil
}

func (r *ProfileFieldRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteProfileField(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete profile field: %w", err)
	}
	return nil
}

func (r *ProfileFieldRepository) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	err := r.q.DeleteProfileFieldsByUser(ctx, uuidToPgtype(userID))
	if err != nil {
		return fmt.Errorf("delete profile fields by user: %w", err)
	}
	return nil
}

func profileFieldToDomain(row pfdb.ProfileField) *domain.ProfileField {
	return &domain.ProfileField{
		ID:         pgtypeToUUID(row.ID),
		UserID:     pgtypeToUUID(row.UserID),
		FieldName:  row.FieldName,
		SourceLang: row.SourceLang,
		UpdatedAt:  row.UpdatedAt.Time,
	}
}
