package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	vacancytextsdb "github.com/ruziba3vich/hr-ai/db/sqlc/vacancy_texts"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyTextRepository struct {
	q *vacancytextsdb.Queries
}

func NewVacancyTextRepository(pool *pgxpool.Pool) *VacancyTextRepository {
	return &VacancyTextRepository{q: vacancytextsdb.New(pool)}
}

func (r *VacancyTextRepository) Create(ctx context.Context, vt *domain.VacancyText) (*domain.VacancyText, error) {
	row, err := r.q.CreateVacancyText(ctx, vacancytextsdb.CreateVacancyTextParams{
		VacancyID:        uuidToPgtype(vt.VacancyID),
		Lang:             vt.Lang,
		Title:            vt.Title,
		Description:      vt.Description,
		Responsibilities: vt.Responsibilities,
		Requirements:     vt.Requirements,
		Benefits:         vt.Benefits,
		IsSource:         vt.IsSource,
		ModelVersion:     textToPgtype(vt.ModelVersion),
	})
	if err != nil {
		return nil, fmt.Errorf("create vacancy text: %w", err)
	}
	return vacancyTextToDomain(row), nil
}

func (r *VacancyTextRepository) Get(ctx context.Context, vacancyID uuid.UUID, lang string) (*domain.VacancyText, error) {
	row, err := r.q.GetVacancyText(ctx, vacancytextsdb.GetVacancyTextParams{
		VacancyID: uuidToPgtype(vacancyID),
		Lang:      lang,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyNotFound
		}
		return nil, fmt.Errorf("get vacancy text: %w", err)
	}
	return vacancyTextToDomain(row), nil
}

func (r *VacancyTextRepository) ListByVacancy(ctx context.Context, vacancyID uuid.UUID) ([]domain.VacancyText, error) {
	rows, err := r.q.ListVacancyTextsByVacancy(ctx, uuidToPgtype(vacancyID))
	if err != nil {
		return nil, fmt.Errorf("list vacancy texts: %w", err)
	}
	result := make([]domain.VacancyText, 0, len(rows))
	for _, row := range rows {
		result = append(result, *vacancyTextToDomain(row))
	}
	return result, nil
}

func (r *VacancyTextRepository) Update(ctx context.Context, vt *domain.VacancyText) (*domain.VacancyText, error) {
	row, err := r.q.UpdateVacancyText(ctx, vacancytextsdb.UpdateVacancyTextParams{
		VacancyID:        uuidToPgtype(vt.VacancyID),
		Lang:             vt.Lang,
		Title:            vt.Title,
		Description:      vt.Description,
		Responsibilities: vt.Responsibilities,
		Requirements:     vt.Requirements,
		Benefits:         vt.Benefits,
		IsSource:         vt.IsSource,
		ModelVersion:     textToPgtype(vt.ModelVersion),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyNotFound
		}
		return nil, fmt.Errorf("update vacancy text: %w", err)
	}
	return vacancyTextToDomain(row), nil
}

func (r *VacancyTextRepository) Delete(ctx context.Context, vacancyID uuid.UUID, lang string) error {
	return r.q.DeleteVacancyText(ctx, vacancytextsdb.DeleteVacancyTextParams{
		VacancyID: uuidToPgtype(vacancyID),
		Lang:      lang,
	})
}

func (r *VacancyTextRepository) DeleteByVacancy(ctx context.Context, vacancyID uuid.UUID) error {
	return r.q.DeleteVacancyTextsByVacancy(ctx, uuidToPgtype(vacancyID))
}

func vacancyTextToDomain(row vacancytextsdb.VacancyText) *domain.VacancyText {
	return &domain.VacancyText{
		VacancyID:        pgtypeToUUID(row.VacancyID),
		Lang:             row.Lang,
		Title:            row.Title,
		Description:      row.Description,
		Responsibilities: row.Responsibilities,
		Requirements:     row.Requirements,
		Benefits:         row.Benefits,
		IsSource:         row.IsSource,
		ModelVersion:     pgtypeToString(row.ModelVersion),
		UpdatedAt:        row.UpdatedAt.Time,
	}
}
