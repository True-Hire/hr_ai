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
	q    *vacanciesdb.Queries
	pool *pgxpool.Pool
}

func NewVacancyRepository(pool *pgxpool.Pool) *VacancyRepository {
	return &VacancyRepository{q: vacanciesdb.New(pool), pool: pool}
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
		MainCategoryID: uuidToPgtype(v.MainCategoryID),
		SubCategoryID:  uuidToPgtype(v.SubCategoryID),
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

func buildVacancyFilterQuery(prefix string, filter domain.VacancyFilter, limit, offset int32) (string, []any) {
	var conditions []string
	var args []any
	argIdx := 1

	if filter.HRID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("hr_id = $%d", argIdx))
		args = append(args, filter.HRID)
		argIdx++
	}
	if filter.Status != "" {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argIdx))
		args = append(args, filter.Status)
		argIdx++
	}
	if filter.Format != "" {
		conditions = append(conditions, fmt.Sprintf("format = $%d", argIdx))
		args = append(args, filter.Format)
		argIdx++
	}
	if filter.Schedule != "" {
		conditions = append(conditions, fmt.Sprintf("schedule = $%d", argIdx))
		args = append(args, filter.Schedule)
		argIdx++
	}
	if filter.SalaryCurrency != "" {
		conditions = append(conditions, fmt.Sprintf("salary_currency = $%d", argIdx))
		args = append(args, filter.SalaryCurrency)
		argIdx++
	}
	if filter.SalaryMin > 0 {
		conditions = append(conditions, fmt.Sprintf("salary_max >= $%d", argIdx))
		args = append(args, filter.SalaryMin)
		argIdx++
	}
	if filter.SalaryMax > 0 {
		conditions = append(conditions, fmt.Sprintf("salary_min <= $%d", argIdx))
		args = append(args, filter.SalaryMax)
		argIdx++
	}
	if filter.ExperienceMin > 0 {
		conditions = append(conditions, fmt.Sprintf("(experience_max >= $%d OR experience_max = 0 OR experience_max IS NULL)", argIdx))
		args = append(args, filter.ExperienceMin)
		argIdx++
	}
	if filter.ExperienceMax > 0 {
		conditions = append(conditions, fmt.Sprintf("experience_min <= $%d", argIdx))
		args = append(args, filter.ExperienceMax)
		argIdx++
	}
	if filter.CountryID != uuid.Nil {
		conditions = append(conditions, fmt.Sprintf("country_id = $%d", argIdx))
		args = append(args, filter.CountryID)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + conditions[0]
		for _, c := range conditions[1:] {
			where += " AND " + c
		}
	}

	query := prefix + where

	if limit > 0 {
		query += " ORDER BY created_at DESC"
		query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIdx, argIdx+1)
		args = append(args, limit, offset)
	}

	return query, args
}

func (r *VacancyRepository) ListFiltered(ctx context.Context, filter domain.VacancyFilter, limit, offset int32) ([]domain.Vacancy, error) {
	query, args := buildVacancyFilterQuery(
		`SELECT id, hr_id, company_data, country_id, salary_min, salary_max, salary_currency,
		    experience_min, experience_max, format, schedule,
		    phone, telegram, email, address, status, source_lang, 
		    main_category_id, sub_category_id, created_at
		FROM vacancies`,
		filter, limit, offset,
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list filtered vacancies: %w", err)
	}
	defer rows.Close()

	var result []domain.Vacancy
	for rows.Next() {
		var v domain.Vacancy
		var id, hrID, countryID, mainCatID, subCatID pgtype.UUID
		var salaryMin, salaryMax, expMin, expMax pgtype.Int4
		var phone, telegram, email, address pgtype.Text
		var companyData []byte
		var createdAt pgtype.Timestamp

		if err := rows.Scan(
			&id, &hrID, &companyData, &countryID,
			&salaryMin, &salaryMax, &v.SalaryCurrency,
			&expMin, &expMax, &v.Format, &v.Schedule,
			&phone, &telegram, &email, &address,
			&v.Status, &v.SourceLang, 
			&mainCatID, &subCatID, &createdAt,
		); err != nil {
			return nil, fmt.Errorf("scan filtered vacancy: %w", err)
		}

		v.ID = pgtypeToUUID(id)
		v.HRID = pgtypeToUUID(hrID)
		v.CountryID = pgtypeUUIDToNullable(countryID)
		v.SalaryMin = pgtypeToInt32(salaryMin)
		v.SalaryMax = pgtypeToInt32(salaryMax)
		v.ExperienceMin = pgtypeToInt32(expMin)
		v.ExperienceMax = pgtypeToInt32(expMax)
		v.Phone = pgtypeToString(phone)
		v.Telegram = pgtypeToString(telegram)
		v.Email = pgtypeToString(email)
		v.Address = pgtypeToString(address)
		v.MainCategoryID = pgtypeToUUID(mainCatID)
		v.SubCategoryID = pgtypeToUUID(subCatID)
		v.CreatedAt = createdAt.Time

		cd, err := unmarshalCompanyData(companyData)
		if err != nil {
			return nil, err
		}
		v.CompanyData = cd

		result = append(result, v)
	}

	return result, nil
}

func (r *VacancyRepository) CountFiltered(ctx context.Context, filter domain.VacancyFilter) (int64, error) {
	query, args := buildVacancyFilterQuery("SELECT COUNT(*) FROM vacancies", filter, 0, 0)
	var count int64
	if err := r.pool.QueryRow(ctx, query, args...).Scan(&count); err != nil {
		return 0, fmt.Errorf("count filtered vacancies: %w", err)
	}
	return count, nil
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
		MainCategoryID: uuidToPgtype(v.MainCategoryID),
		SubCategoryID:  uuidToPgtype(v.SubCategoryID),
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
		MainCategoryID: pgtypeToUUID(row.MainCategoryID),
		SubCategoryID:  pgtypeToUUID(row.SubCategoryID),
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
		MainCategoryID: pgtypeToUUID(row.MainCategoryID),
		SubCategoryID:  pgtypeToUUID(row.SubCategoryID),
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
		MainCategoryID: pgtypeToUUID(row.MainCategoryID),
		SubCategoryID:  pgtypeToUUID(row.SubCategoryID),
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
		MainCategoryID: pgtypeToUUID(row.MainCategoryID),
		SubCategoryID:  pgtypeToUUID(row.SubCategoryID),
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
		MainCategoryID: pgtypeToUUID(row.MainCategoryID),
		SubCategoryID:  pgtypeToUUID(row.SubCategoryID),
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
		MainCategoryID: pgtypeToUUID(row.MainCategoryID),
		SubCategoryID:  pgtypeToUUID(row.SubCategoryID),
		CreatedAt:      row.CreatedAt.Time,
	}, nil
}
