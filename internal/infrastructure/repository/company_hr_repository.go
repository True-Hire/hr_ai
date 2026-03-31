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

	companyhrsdb "github.com/ruziba3vich/hr-ai/db/sqlc/company_hrs"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type CompanyHRRepository struct {
	q *companyhrsdb.Queries
}

func NewCompanyHRRepository(pool *pgxpool.Pool) *CompanyHRRepository {
	return &CompanyHRRepository{
		q: companyhrsdb.New(pool),
	}
}

// companyHRRow is a normalised view of the fields shared by every SQLC row
// type returned from company_hrs queries.  Each per-query row type is
// converted to this struct before mapping to the domain model so that a
// single companyHRToDomain helper can serve all callers.
type companyHRRow struct {
	ID           pgtype.UUID
	FirstName    string
	LastName     string
	Patronymic   pgtype.Text
	Phone        pgtype.Text
	Telegram     pgtype.Text
	TelegramID   pgtype.Text
	Email        pgtype.Text
	Position     pgtype.Text
	Status       string
	PasswordHash pgtype.Text
	CreatedAt    pgtype.Timestamp
	CompanyData  []byte
	Language     string
}

func (r *CompanyHRRepository) Create(ctx context.Context, hr *domain.CompanyHR) (*domain.CompanyHR, error) {
	companyDataBytes, err := json.Marshal(hr.CompanyData)
	if err != nil {
		return nil, fmt.Errorf("create company hr: marshal company data: %w", err)
	}

	raw, err := r.q.CreateCompanyHR(ctx, companyhrsdb.CreateCompanyHRParams{
		ID:          uuidToPgtype(hr.ID),
		FirstName:   hr.FirstName,
		LastName:    hr.LastName,
		Patronymic:  textToPgtype(hr.Patronymic),
		Phone:       textToPgtype(hr.Phone),
		Telegram:    textToPgtype(hr.Telegram),
		TelegramID:  textToPgtype(hr.TelegramID),
		Email:       textToPgtype(hr.Email),
		Position:    textToPgtype(hr.Position),
		Status:      hr.Status,
		CompanyData: companyDataBytes,
		Language:    hr.Language,
	})
	if err != nil {
		return nil, fmt.Errorf("create company hr: %w", err)
	}

	return companyHRToDomain(companyHRRow{
		ID:           raw.ID,
		FirstName:    raw.FirstName,
		LastName:     raw.LastName,
		Patronymic:   raw.Patronymic,
		Phone:        raw.Phone,
		Telegram:     raw.Telegram,
		TelegramID:   raw.TelegramID,
		Email:        raw.Email,
		Position:     raw.Position,
		Status:       raw.Status,
		PasswordHash: raw.PasswordHash,
		CreatedAt:    raw.CreatedAt,
		CompanyData:  raw.CompanyData,
		Language:     raw.Language,
	})
}

func (r *CompanyHRRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.CompanyHR, error) {
	raw, err := r.q.GetCompanyHRByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("get company hr by id: %w", err)
	}
	return companyHRToDomain(companyHRRow{
		ID:           raw.ID,
		FirstName:    raw.FirstName,
		LastName:     raw.LastName,
		Patronymic:   raw.Patronymic,
		Phone:        raw.Phone,
		Telegram:     raw.Telegram,
		TelegramID:   raw.TelegramID,
		Email:        raw.Email,
		Position:     raw.Position,
		Status:       raw.Status,
		PasswordHash: raw.PasswordHash,
		CreatedAt:    raw.CreatedAt,
		CompanyData:  raw.CompanyData,
		Language:     raw.Language,
	})
}

func (r *CompanyHRRepository) List(ctx context.Context, limit, offset int32) ([]domain.CompanyHR, error) {
	rows, err := r.q.ListCompanyHRs(ctx, companyhrsdb.ListCompanyHRsParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list company hrs: %w", err)
	}
	hrs := make([]domain.CompanyHR, 0, len(rows))
	for _, raw := range rows {
		hr, err := companyHRToDomain(companyHRRow{
			ID:           raw.ID,
			FirstName:    raw.FirstName,
			LastName:     raw.LastName,
			Patronymic:   raw.Patronymic,
			Phone:        raw.Phone,
			Telegram:     raw.Telegram,
			TelegramID:   raw.TelegramID,
			Email:        raw.Email,
			Position:     raw.Position,
			Status:       raw.Status,
			PasswordHash: raw.PasswordHash,
			CreatedAt:    raw.CreatedAt,
			CompanyData:  raw.CompanyData,
			Language:     raw.Language,
		})
		if err != nil {
			return nil, err
		}
		hrs = append(hrs, *hr)
	}
	return hrs, nil
}

func (r *CompanyHRRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.q.CountCompanyHRs(ctx)
	if err != nil {
		return 0, fmt.Errorf("count company hrs: %w", err)
	}
	return count, nil
}

func (r *CompanyHRRepository) Update(ctx context.Context, hr *domain.CompanyHR) (*domain.CompanyHR, error) {
	companyDataBytes, err := json.Marshal(hr.CompanyData)
	if err != nil {
		return nil, fmt.Errorf("update company hr: marshal company data: %w", err)
	}

	raw, err := r.q.UpdateCompanyHR(ctx, companyhrsdb.UpdateCompanyHRParams{
		ID:          uuidToPgtype(hr.ID),
		FirstName:   hr.FirstName,
		LastName:    hr.LastName,
		Patronymic:  hr.Patronymic,
		Phone:       hr.Phone,
		Telegram:    hr.Telegram,
		TelegramID:  hr.TelegramID,
		Email:       hr.Email,
		Position:    hr.Position,
		Status:      hr.Status,
		CompanyData: companyDataBytes,
		Language:    hr.Language,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("update company hr: %w", err)
	}
	return companyHRToDomain(companyHRRow{
		ID:           raw.ID,
		FirstName:    raw.FirstName,
		LastName:     raw.LastName,
		Patronymic:   raw.Patronymic,
		Phone:        raw.Phone,
		Telegram:     raw.Telegram,
		TelegramID:   raw.TelegramID,
		Email:        raw.Email,
		Position:     raw.Position,
		Status:       raw.Status,
		PasswordHash: raw.PasswordHash,
		CreatedAt:    raw.CreatedAt,
		CompanyData:  raw.CompanyData,
		Language:     raw.Language,
	})
}

func (r *CompanyHRRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteCompanyHR(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete company hr: %w", err)
	}
	return nil
}

func (r *CompanyHRRepository) GetByPhone(ctx context.Context, phone string) (*domain.CompanyHR, error) {
	raw, err := r.q.GetCompanyHRByPhone(ctx, textToPgtype(phone))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("get company hr by phone: %w", err)
	}
	return companyHRToDomain(companyHRRow{
		ID:           raw.ID,
		FirstName:    raw.FirstName,
		LastName:     raw.LastName,
		Patronymic:   raw.Patronymic,
		Phone:        raw.Phone,
		Telegram:     raw.Telegram,
		TelegramID:   raw.TelegramID,
		Email:        raw.Email,
		Position:     raw.Position,
		Status:       raw.Status,
		PasswordHash: raw.PasswordHash,
		CreatedAt:    raw.CreatedAt,
		CompanyData:  raw.CompanyData,
		Language:     raw.Language,
	})
}

func (r *CompanyHRRepository) GetByEmail(ctx context.Context, email string) (*domain.CompanyHR, error) {
	raw, err := r.q.GetCompanyHRByEmail(ctx, textToPgtype(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("get company hr by email: %w", err)
	}
	return companyHRToDomain(companyHRRow{
		ID:           raw.ID,
		FirstName:    raw.FirstName,
		LastName:     raw.LastName,
		Patronymic:   raw.Patronymic,
		Phone:        raw.Phone,
		Telegram:     raw.Telegram,
		TelegramID:   raw.TelegramID,
		Email:        raw.Email,
		Position:     raw.Position,
		Status:       raw.Status,
		PasswordHash: raw.PasswordHash,
		CreatedAt:    raw.CreatedAt,
		CompanyData:  raw.CompanyData,
		Language:     raw.Language,
	})
}

func (r *CompanyHRRepository) GetByTelegramID(ctx context.Context, telegramID string) (*domain.CompanyHR, error) {
	raw, err := r.q.GetCompanyHRByTelegramID(ctx, textToPgtype(telegramID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrCompanyHRNotFound
		}
		return nil, fmt.Errorf("get company hr by telegram id: %w", err)
	}
	return companyHRToDomain(companyHRRow{
		ID:           raw.ID,
		FirstName:    raw.FirstName,
		LastName:     raw.LastName,
		Patronymic:   raw.Patronymic,
		Phone:        raw.Phone,
		Telegram:     raw.Telegram,
		TelegramID:   raw.TelegramID,
		Email:        raw.Email,
		Position:     raw.Position,
		Status:       raw.Status,
		PasswordHash: raw.PasswordHash,
		CreatedAt:    raw.CreatedAt,
		CompanyData:  raw.CompanyData,
		Language:     raw.Language,
	})
}

func (r *CompanyHRRepository) SetPassword(ctx context.Context, id uuid.UUID, hash string) error {
	err := r.q.SetCompanyHRPassword(ctx, companyhrsdb.SetCompanyHRPasswordParams{
		ID:           uuidToPgtype(id),
		PasswordHash: textToPgtype(hash),
	})
	if err != nil {
		return fmt.Errorf("set company hr password: %w", err)
	}
	return nil
}

// companyHRToDomain converts the normalised companyHRRow into a domain.CompanyHR.
// It unmarshals the JSON-encoded CompanyData field; a nil or empty byte slice
// results in a nil *domain.CompanyData pointer.
func companyHRToDomain(row companyHRRow) (*domain.CompanyHR, error) {
	var companyData *domain.CompanyData
	if len(row.CompanyData) > 0 {
		companyData = new(domain.CompanyData)
		if err := json.Unmarshal(row.CompanyData, companyData); err != nil {
			return nil, fmt.Errorf("company hr to domain: unmarshal company data: %w", err)
		}
	}

	return &domain.CompanyHR{
		ID:           pgtypeToUUID(row.ID),
		FirstName:    row.FirstName,
		LastName:     row.LastName,
		Patronymic:   pgtypeToString(row.Patronymic),
		Phone:        pgtypeToString(row.Phone),
		Telegram:     pgtypeToString(row.Telegram),
		TelegramID:   pgtypeToString(row.TelegramID),
		Email:        pgtypeToString(row.Email),
		Position:     pgtypeToString(row.Position),
		Status:       row.Status,
		PasswordHash: pgtypeToString(row.PasswordHash),
		CompanyData:  companyData,
		Language:     row.Language,
		CreatedAt:    row.CreatedAt.Time,
	}, nil
}
