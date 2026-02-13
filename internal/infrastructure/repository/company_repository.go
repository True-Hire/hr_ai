package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	companiesdb "github.com/ruziba3vich/hr-ai/db/sqlc/companies"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyRepository struct {
	q *companiesdb.Queries
}

func NewCompanyRepository(pool *pgxpool.Pool) *CompanyRepository {
	return &CompanyRepository{
		q: companiesdb.New(pool),
	}
}

func (r *CompanyRepository) Create(ctx context.Context, c *domain.Company) (*domain.Company, error) {
	row, err := r.q.CreateCompany(ctx, companiesdb.CreateCompanyParams{
		ID:              uuidToPgtype(c.ID),
		EmployeeCount:   int4ToPgtype(c.EmployeeCount),
		Country:         textToPgtype(c.Country),
		Address:         textToPgtype(c.Address),
		Phone:           textToPgtype(c.Phone),
		Telegram:        textToPgtype(c.Telegram),
		TelegramChannel: textToPgtype(c.TelegramChannel),
		Email:           textToPgtype(c.Email),
		LogoUrl:         textToPgtype(c.LogoURL),
		WebSite:         textToPgtype(c.WebSite),
		Instagram:       textToPgtype(c.Instagram),
		SourceLang:      c.SourceLang,
	})
	if err != nil {
		return nil, fmt.Errorf("create company: %w", err)
	}
	return companyToDomain(row), nil
}

func (r *CompanyRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Company, error) {
	row, err := r.q.GetCompanyByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyNotFound
		}
		return nil, fmt.Errorf("get company by id: %w", err)
	}
	return companyToDomain(row), nil
}

func (r *CompanyRepository) List(ctx context.Context, limit, offset int32) ([]domain.Company, error) {
	rows, err := r.q.ListCompanies(ctx, companiesdb.ListCompaniesParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list companies: %w", err)
	}
	companies := make([]domain.Company, 0, len(rows))
	for _, row := range rows {
		companies = append(companies, *companyToDomain(row))
	}
	return companies, nil
}

func (r *CompanyRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.q.CountCompanies(ctx)
	if err != nil {
		return 0, fmt.Errorf("count companies: %w", err)
	}
	return count, nil
}

func (r *CompanyRepository) Update(ctx context.Context, c *domain.Company) (*domain.Company, error) {
	row, err := r.q.UpdateCompany(ctx, companiesdb.UpdateCompanyParams{
		ID:              uuidToPgtype(c.ID),
		EmployeeCount:   c.EmployeeCount,
		Country:         c.Country,
		Address:         c.Address,
		Phone:           c.Phone,
		Telegram:        c.Telegram,
		TelegramChannel: c.TelegramChannel,
		Email:           c.Email,
		LogoUrl:         c.LogoURL,
		WebSite:         c.WebSite,
		Instagram:       c.Instagram,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyNotFound
		}
		return nil, fmt.Errorf("update company: %w", err)
	}
	return companyToDomain(row), nil
}

func (r *CompanyRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteCompany(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete company: %w", err)
	}
	return nil
}

func companyToDomain(row companiesdb.Company) *domain.Company {
	return &domain.Company{
		ID:              pgtypeToUUID(row.ID),
		EmployeeCount:   pgtypeToInt32(row.EmployeeCount),
		Country:         pgtypeToString(row.Country),
		Address:         pgtypeToString(row.Address),
		Phone:           pgtypeToString(row.Phone),
		Telegram:        pgtypeToString(row.Telegram),
		TelegramChannel: pgtypeToString(row.TelegramChannel),
		Email:           pgtypeToString(row.Email),
		LogoURL:         pgtypeToString(row.LogoUrl),
		WebSite:         pgtypeToString(row.WebSite),
		Instagram:       pgtypeToString(row.Instagram),
		SourceLang:      row.SourceLang,
		CreatedAt:       row.CreatedAt.Time,
	}
}
