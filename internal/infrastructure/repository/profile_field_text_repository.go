package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	pftdb "github.com/ruziba3vich/hr-ai/db/sqlc/profile_field_texts"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type ProfileFieldTextRepository struct {
	q *pftdb.Queries
}

func NewProfileFieldTextRepository(pool *pgxpool.Pool) *ProfileFieldTextRepository {
	return &ProfileFieldTextRepository{
		q: pftdb.New(pool),
	}
}

func (r *ProfileFieldTextRepository) Create(ctx context.Context, text *domain.ProfileFieldText) (*domain.ProfileFieldText, error) {
	row, err := r.q.CreateProfileFieldText(ctx, pftdb.CreateProfileFieldTextParams{
		ProfileFieldID: uuidToPgtype(text.ProfileFieldID),
		Lang:           text.Lang,
		Content:        text.Content,
		IsSource:       text.IsSource,
		ModelVersion:   textToPgtype(text.ModelVersion),
	})
	if err != nil {
		return nil, fmt.Errorf("create profile field text: %w", err)
	}
	return profileFieldTextToDomain(row), nil
}

func (r *ProfileFieldTextRepository) Get(ctx context.Context, profileFieldID uuid.UUID, lang string) (*domain.ProfileFieldText, error) {
	row, err := r.q.GetProfileFieldText(ctx, pftdb.GetProfileFieldTextParams{
		ProfileFieldID: uuidToPgtype(profileFieldID),
		Lang:           lang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileFieldTextNotFound
		}
		return nil, fmt.Errorf("get profile field text: %w", err)
	}
	return profileFieldTextToDomain(row), nil
}

func (r *ProfileFieldTextRepository) ListByField(ctx context.Context, profileFieldID uuid.UUID) ([]domain.ProfileFieldText, error) {
	rows, err := r.q.ListProfileFieldTexts(ctx, uuidToPgtype(profileFieldID))
	if err != nil {
		return nil, fmt.Errorf("list profile field texts: %w", err)
	}
	texts := make([]domain.ProfileFieldText, 0, len(rows))
	for _, row := range rows {
		texts = append(texts, *profileFieldTextToDomain(row))
	}
	return texts, nil
}

func (r *ProfileFieldTextRepository) Update(ctx context.Context, text *domain.ProfileFieldText) (*domain.ProfileFieldText, error) {
	row, err := r.q.UpdateProfileFieldText(ctx, pftdb.UpdateProfileFieldTextParams{
		ProfileFieldID: uuidToPgtype(text.ProfileFieldID),
		Lang:           text.Lang,
		Content:        text.Content,
		ModelVersion:   textToPgtype(text.ModelVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrProfileFieldTextNotFound
		}
		return nil, fmt.Errorf("update profile field text: %w", err)
	}
	return profileFieldTextToDomain(row), nil
}

func (r *ProfileFieldTextRepository) Delete(ctx context.Context, profileFieldID uuid.UUID, lang string) error {
	err := r.q.DeleteProfileFieldText(ctx, pftdb.DeleteProfileFieldTextParams{
		ProfileFieldID: uuidToPgtype(profileFieldID),
		Lang:           lang,
	})
	if err != nil {
		return fmt.Errorf("delete profile field text: %w", err)
	}
	return nil
}

func (r *ProfileFieldTextRepository) DeleteByField(ctx context.Context, profileFieldID uuid.UUID) error {
	err := r.q.DeleteProfileFieldTextsByField(ctx, uuidToPgtype(profileFieldID))
	if err != nil {
		return fmt.Errorf("delete profile field texts by field: %w", err)
	}
	return nil
}

func profileFieldTextToDomain(row pftdb.ProfileFieldText) *domain.ProfileFieldText {
	return &domain.ProfileFieldText{
		ProfileFieldID: pgtypeToUUID(row.ProfileFieldID),
		Lang:           row.Lang,
		Content:        row.Content,
		IsSource:       row.IsSource,
		ModelVersion:   pgtypeToString(row.ModelVersion),
		UpdatedAt:      row.UpdatedAt.Time,
	}
}
