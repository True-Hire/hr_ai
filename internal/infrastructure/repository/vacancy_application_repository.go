package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	vadb "github.com/ruziba3vich/hr-ai/db/sqlc/vacancy_applications"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyApplicationRepository struct {
	q *vadb.Queries
}

func NewVacancyApplicationRepository(pool *pgxpool.Pool) *VacancyApplicationRepository {
	return &VacancyApplicationRepository{q: vadb.New(pool)}
}

func (r *VacancyApplicationRepository) Create(ctx context.Context, va *domain.VacancyApplication) (*domain.VacancyApplication, error) {
	row, err := r.q.CreateVacancyApplication(ctx, vadb.CreateVacancyApplicationParams{
		ID:          uuidToPgtype(va.ID),
		UserID:      uuidToPgtype(va.UserID),
		VacancyID:   uuidToPgtype(va.VacancyID),
		Status:      va.Status,
		CoverLetter: textToPgtype(va.CoverLetter),
	})
	if err != nil {
		return nil, fmt.Errorf("create vacancy application: %w", err)
	}
	return vacancyApplicationFromRow(row), nil
}

func (r *VacancyApplicationRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.VacancyApplication, error) {
	row, err := r.q.GetVacancyApplicationByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyApplicationNotFound
		}
		return nil, fmt.Errorf("get vacancy application by id: %w", err)
	}
	return vacancyApplicationFromRow(row), nil
}

func (r *VacancyApplicationRepository) GetByUserAndVacancy(ctx context.Context, userID, vacancyID uuid.UUID) (*domain.VacancyApplication, error) {
	row, err := r.q.GetVacancyApplicationByUserAndVacancy(ctx, vadb.GetVacancyApplicationByUserAndVacancyParams{
		UserID:    uuidToPgtype(userID),
		VacancyID: uuidToPgtype(vacancyID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyApplicationNotFound
		}
		return nil, fmt.Errorf("get vacancy application by user and vacancy: %w", err)
	}
	return vacancyApplicationFromRow(row), nil
}

func (r *VacancyApplicationRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.VacancyApplication, error) {
	rows, err := r.q.ListVacancyApplicationsByUser(ctx, vadb.ListVacancyApplicationsByUserParams{
		UserID: uuidToPgtype(userID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list vacancy applications by user: %w", err)
	}
	result := make([]domain.VacancyApplication, 0, len(rows))
	for _, row := range rows {
		result = append(result, *vacancyApplicationFromRow(row))
	}
	return result, nil
}

func (r *VacancyApplicationRepository) ListByVacancy(ctx context.Context, vacancyID uuid.UUID, limit, offset int32) ([]domain.VacancyApplication, error) {
	rows, err := r.q.ListVacancyApplicationsByVacancy(ctx, vadb.ListVacancyApplicationsByVacancyParams{
		VacancyID: uuidToPgtype(vacancyID),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list vacancy applications by vacancy: %w", err)
	}
	result := make([]domain.VacancyApplication, 0, len(rows))
	for _, row := range rows {
		result = append(result, *vacancyApplicationFromRow(row))
	}
	return result, nil
}

func (r *VacancyApplicationRepository) CountByUser(ctx context.Context, userID uuid.UUID) (int64, error) {
	return r.q.CountVacancyApplicationsByUser(ctx, uuidToPgtype(userID))
}

func (r *VacancyApplicationRepository) CountByVacancy(ctx context.Context, vacancyID uuid.UUID) (int64, error) {
	return r.q.CountVacancyApplicationsByVacancy(ctx, uuidToPgtype(vacancyID))
}

func (r *VacancyApplicationRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string) (*domain.VacancyApplication, error) {
	row, err := r.q.UpdateVacancyApplicationStatus(ctx, vadb.UpdateVacancyApplicationStatusParams{
		ID:     uuidToPgtype(id),
		Status: status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyApplicationNotFound
		}
		return nil, fmt.Errorf("update vacancy application status: %w", err)
	}
	return vacancyApplicationFromRow(row), nil
}

func (r *VacancyApplicationRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteVacancyApplication(ctx, uuidToPgtype(id))
}

func vacancyApplicationFromRow(row vadb.VacancyApplication) *domain.VacancyApplication {
	return &domain.VacancyApplication{
		ID:          pgtypeToUUID(row.ID),
		UserID:      pgtypeToUUID(row.UserID),
		VacancyID:   pgtypeToUUID(row.VacancyID),
		Status:      row.Status,
		CoverLetter: pgtypeToString(row.CoverLetter),
		CreatedAt:   row.CreatedAt.Time,
		UpdatedAt:   row.UpdatedAt.Time,
	}
}
