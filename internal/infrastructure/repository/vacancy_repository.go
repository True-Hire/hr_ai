package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
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
	companyDataBytes, err := json.Marshal(v.CompanyData)
	if err != nil {
		return nil, fmt.Errorf("create vacancy: marshal company data: %w", err)
	}
	row, err := r.q.CreateVacancy(ctx, vacanciesdb.CreateVacancyParams{
		ID:             uuidToPgtype(v.ID),
		HrID:           uuidToPgtype(v.HRID),
		CompanyData:    companyDataBytes,
		CountryID:      uuidToPgtypeNullable(v.CountryID),
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
	return vacancyFromCreateRow(row)
}

func (r *VacancyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Vacancy, error) {
	row, err := r.q.GetVacancyByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyNotFound
		}
		return nil, fmt.Errorf("get vacancy by id: %w", err)
	}
	return vacancyFromGetRow(row)
}

func (r *VacancyRepository) List(ctx context.Context, limit, offset int32) ([]domain.Vacancy, error) {
	rows, err := r.q.ListVacancies(ctx, vacanciesdb.ListVacanciesParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, fmt.Errorf("list vacancies: %w", err)
	}
	result := make([]domain.Vacancy, 0, len(rows))
	for _, row := range rows {
		v, err := vacancyFromListRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, *v)
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
		v, err := vacancyFromListByHRRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, *v)
	}
	return result, nil
}

func (r *VacancyRepository) Search(ctx context.Context, lang, query string, limit, offset int32) ([]domain.Vacancy, error) {
	rows, err := r.q.SearchVacancies(ctx, vacanciesdb.SearchVacanciesParams{
		Lang:  lang,
		Query: textToPgtype(query),
		Lim:   limit,
		Off:   offset,
	})
	if err != nil {
		return nil, fmt.Errorf("search vacancies: %w", err)
	}
	result := make([]domain.Vacancy, 0, len(rows))
	for _, row := range rows {
		v, err := vacancyFromSearchRow(row)
		if err != nil {
			return nil, err
		}
		result = append(result, *v)
	}
	return result, nil
}

func (r *VacancyRepository) Count(ctx context.Context) (int64, error) {
	return r.q.CountVacancies(ctx)
}

func (r *VacancyRepository) CountByHR(ctx context.Context, hrID uuid.UUID) (int64, error) {
	return r.q.CountVacanciesByHR(ctx, uuidToPgtype(hrID))
}

func (r *VacancyRepository) CountSearch(ctx context.Context, lang, query string) (int64, error) {
	return r.q.CountSearchVacancies(ctx, vacanciesdb.CountSearchVacanciesParams{
		Lang:  lang,
		Query: textToPgtype(query),
	})
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
		CountryID:      uuidToPgtypeNullable(v.CountryID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrVacancyNotFound
		}
		return nil, fmt.Errorf("update vacancy: %w", err)
	}
	return vacancyFromUpdateRow(row)
}

func (r *VacancyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.q.DeleteVacancy(ctx, uuidToPgtype(id))
}

func (r *VacancyRepository) NullifyHRID(ctx context.Context, hrID uuid.UUID) error {
	return r.q.NullifyVacancyHRID(ctx, uuidToPgtype(hrID))
}

func (r *VacancyRepository) ListIDsByHR(ctx context.Context, hrID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := r.q.ListVacancyIDsByHR(ctx, uuidToPgtype(hrID))
	if err != nil {
		return nil, fmt.Errorf("list vacancy ids by hr: %w", err)
	}
	result := make([]uuid.UUID, 0, len(rows))
	for _, row := range rows {
		result = append(result, pgtypeToUUID(row))
	}
	return result, nil
}

func pgtypeUUIDToNullable(id pgtype.UUID) uuid.UUID {
	if !id.Valid {
		return uuid.Nil
	}
	return uuid.UUID(id.Bytes)
}

func unmarshalCompanyData(b []byte) (*domain.CompanyData, error) {
	if len(b) == 0 {
		return nil, nil
	}
	var cd domain.CompanyData
	if err := json.Unmarshal(b, &cd); err != nil {
		return nil, fmt.Errorf("unmarshal company data: %w", err)
	}
	return &cd, nil
}

func vacancyFromCreateRow(row vacanciesdb.CreateVacancyRow) (*domain.Vacancy, error) {
	cd, err := unmarshalCompanyData(row.CompanyData)
	if err != nil {
		return nil, err
	}
	return &domain.Vacancy{
		ID:             pgtypeToUUID(row.ID),
		HRID:           pgtypeToUUID(row.HrID),
		CompanyData:    cd,
		CountryID:      pgtypeUUIDToNullable(row.CountryID),
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
	}, nil
}

func vacancyFromGetRow(row vacanciesdb.GetVacancyByIDRow) (*domain.Vacancy, error) {
	cd, err := unmarshalCompanyData(row.CompanyData)
	if err != nil {
		return nil, err
	}
	return &domain.Vacancy{
		ID:             pgtypeToUUID(row.ID),
		HRID:           pgtypeToUUID(row.HrID),
		CompanyData:    cd,
		CountryID:      pgtypeUUIDToNullable(row.CountryID),
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
	}, nil
}

func vacancyFromListRow(row vacanciesdb.ListVacanciesRow) (*domain.Vacancy, error) {
	cd, err := unmarshalCompanyData(row.CompanyData)
	if err != nil {
		return nil, err
	}
	return &domain.Vacancy{
		ID:             pgtypeToUUID(row.ID),
		HRID:           pgtypeToUUID(row.HrID),
		CompanyData:    cd,
		CountryID:      pgtypeUUIDToNullable(row.CountryID),
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
	}, nil
}

func vacancyFromListByHRRow(row vacanciesdb.ListVacanciesByHRRow) (*domain.Vacancy, error) {
	cd, err := unmarshalCompanyData(row.CompanyData)
	if err != nil {
		return nil, err
	}
	return &domain.Vacancy{
		ID:             pgtypeToUUID(row.ID),
		HRID:           pgtypeToUUID(row.HrID),
		CompanyData:    cd,
		CountryID:      pgtypeUUIDToNullable(row.CountryID),
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
	}, nil
}

func vacancyFromSearchRow(row vacanciesdb.SearchVacanciesRow) (*domain.Vacancy, error) {
	cd, err := unmarshalCompanyData(row.CompanyData)
	if err != nil {
		return nil, err
	}
	return &domain.Vacancy{
		ID:             pgtypeToUUID(row.ID),
		HRID:           pgtypeToUUID(row.HrID),
		CompanyData:    cd,
		CountryID:      pgtypeUUIDToNullable(row.CountryID),
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
	}, nil
}

func vacancyFromUpdateRow(row vacanciesdb.UpdateVacancyRow) (*domain.Vacancy, error) {
	cd, err := unmarshalCompanyData(row.CompanyData)
	if err != nil {
		return nil, err
	}
	return &domain.Vacancy{
		ID:             pgtypeToUUID(row.ID),
		HRID:           pgtypeToUUID(row.HrID),
		CompanyData:    cd,
		CountryID:      pgtypeUUIDToNullable(row.CountryID),
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
	}, nil
}
