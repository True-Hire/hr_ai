package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	vacanciesdb "github.com/ruziba3vich/hr-ai/db/sqlc/vacancies"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type VacancyRepository struct {
	q *vacanciesdb.Queries
}

func NewVacancyRepository(pool *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{q: vacanciesdb.New(pool)}
}

func (r *VacancyRepository) Create(ctx context.Context, v *domain.Vacancy) (*domain.Vacancy, error) {
	row, err := r.q.CreateVacancy(ctx, vacanciesdb.CreateVacancyParams{
		ID:             uuidToPgtype(v.ID),
		HrID:           uuidToPgtype(v.HRID),
		CompanyID:      uuidToPgtype(v.CompanyID),
		SalaryMin:      int4ToPgtype(v.SalaryMin),
		SalaryMax:      int4ToPgtype(v.SalaryMax),
		SalaryCurrency: v.SalaryCurrency,
		ExperienceMin:  int4ToPgtype(v.ExperienceMin),
		ExperienceMax:  int4ToPgtype(v.ExperienceMax),
		Format:         v.Format,
		Schedule:       v.Schedule,
		Phone:          textToPgtype(v.Phone),
		Telegram:       textToPgtype(v.Telegram),
		Email:          textToPgtype(v.Email),
		Address:        textToPgtype(v.Address),
		Status:         v.Status,
		SourceLang:     v.SourceLang,
	})
	if err != nil {
		return nil, fmt.Errorf("create vacancy: %w", err)
	}
	return vacancyToDomain(row), nil
}

func (r *VacancyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Vacancy, error) {
	row, err := r.q.GetVacancyByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyNotFound
		}
		return nil, fmt.Errorf("get vacancy by id: %w", err)
	}
	return vacancyToDomain(row), nil
}

func (r *VacancyRepository) List(ctx context.Context, limit, offset int32) ([]domain.Vacancy, error) {
	rows, err := r.q.ListVacancies(ctx, vacanciesdb.ListVacanciesParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list vacancies: %w", err)
	}
	result := make([]domain.Vacancy, 0, len(rows))
	for _, row := range rows {
		result = append(result, *vacancyToDomain(row))
	}
	return result, nil
}

func (r *VacancyRepository) ListByCompany(ctx context.Context, companyID uuid.UUID, limit, offset int32) ([]domain.Vacancy, error) {
	rows, err := r.q.ListVacanciesByCompany(ctx, vacanciesdb.ListVacanciesByCompanyParams{
		CompanyID: uuidToPgtype(companyID),
		Limit:     limit,
		Offset:    offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list vacancies by company: %w", err)
	}
	result := make([]domain.Vacancy, 0, len(rows))
	for _, row := range rows {
		result = append(result, *vacancyToDomain(row))
	}
	return result, nil
}

func (r *VacancyRepository) ListByHR(ctx context.Context, hrID uuid.UUID, limit, offset int32) ([]domain.Vacancy, error) {
	rows, err := r.q.ListVacanciesByHR(ctx, vacanciesdb.ListVacanciesByHRParams{
		HrID:   uuidToPgtype(hrID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list vacancies by hr: %w", err)
	}
	result := make([]domain.Vacancy, 0, len(rows))
	for _, row := range rows {
		result = append(result, *vacancyToDomain(row))
	}
	return result, nil
}

func (r *VacancyRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountVacancies(ctx)
}

func (r *VacancyRepository) CountByCompany(ctx context.Context, companyID uuid.UUID) (int64, error) {
	return r.q.CountVacanciesByCompany(ctx, uuidToPgtype(companyID))
}

func (r *VacancyRepository) Update(ctx context.Context, v *domain.Vacancy) (*domain.Vacancy, error) {
	row, err := r.q.UpdateVacancy(ctx, vacanciesdb.UpdateVacancyParams{
		ID:             uuidToPgtype(v.ID),
		SalaryMin:      v.SalaryMin,
		SalaryMax:      v.SalaryMax,
		SalaryCurrency: v.SalaryCurrency,
		ExperienceMin:  v.ExperienceMin,
		ExperienceMax:  v.ExperienceMax,
		Format:         v.Format,
		Schedule:       v.Schedule,
		Phone:          v.Phone,
		Telegram:       v.Telegram,
		Email:          v.Email,
		Address:        v.Address,
		Status:         v.Status,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyNotFound
		}
		return nil, fmt.Errorf("update vacancy: %w", err)
	}
	return vacancyToDomain(row), nil
}

func (r *VacancyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteVacancy(ctx, uuidToPgtype(id))
}

func vacancyToDomain(row vacanciesdb.Vacancy) *domain.Vacancy {
	return &domain.Vacancy{
		ID:             pgtypeToUUID(row.ID),
		HRID:           pgtypeToUUID(row.HrID),
		CompanyID:      pgtypeToUUID(row.CompanyID),
		SalaryMin:      pgtypeToInt32(row.SalaryMin),
		SalaryMax:      pgtypeToInt32(row.SalaryMax),
		SalaryCurrency: row.SalaryCurrency,
		ExperienceMin:  pgtypeToInt32(row.ExperienceMin),
		ExperienceMax:  pgtypeToInt32(row.ExperienceMax),
		Format:         row.Format,
		Schedule:       row.Schedule,
		Phone:          pgtypeToString(row.Phone),
		Telegram:       pgtypeToString(row.Telegram),
		Email:          pgtypeToString(row.Email),
		Address:        pgtypeToString(row.Address),
		Status:         row.Status,
		SourceLang:     row.SourceLang,
		CreatedAt:      row.CreatedAt.Time,
	}
}
