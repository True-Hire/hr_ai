package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	usersdb "github.com/ruziba3vich/hr-ai/db/sqlc/users"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type UserRepository struct {
	q *usersdb.Queries
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{
		q: usersdb.New(pool),
	}
}

func (r *UserRepository) Create(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := r.q.CreateUser(ctx, usersdb.CreateUserParams{
		ID:              uuidToPgtype(user.ID),
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Patronymic:      textToPgtype(user.Patronymic),
		Phone:           textToPgtype(user.Phone),
		Telegram:        textToPgtype(user.Telegram),
		TelegramID:      textToPgtype(user.TelegramID),
		Email:           textToPgtype(user.Email),
		Gender:          textToPgtype(user.Gender),
		Country:         textToPgtype(user.Country),
		Region:          textToPgtype(user.Region),
		Nationality:     textToPgtype(user.Nationality),
		ProfilePicUrl:   textToPgtype(user.ProfilePicURL),
		Status:          user.Status,
		TariffType:      user.TariffType,
		JobStatus:       textToPgtype(user.JobStatus),
		ActivityType:    textToPgtype(user.ActivityType),
		Specializations: user.Specializations,
	})
	if err != nil {
		return nil, fmt.Errorf("create user: %w", err)
	}
	return userToDomain(row), nil
}

func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	row, err := r.q.GetUserByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return userToDomain(row), nil
}

func (r *UserRepository) List(ctx context.Context, limit, offset int32) ([]domain.User, error) {
	rows, err := r.q.ListUsers(ctx, usersdb.ListUsersParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	users := make([]domain.User, 0, len(rows))
	for _, row := range rows {
		users = append(users, *userToDomain(row))
	}
	return users, nil
}

func (r *UserRepository) Count(ctx context.Context) (int64, error) {
	count, err := r.q.CountUsers(ctx)
	if err != nil {
		return 0, fmt.Errorf("count users: %w", err)
	}
	return count, nil
}

func (r *UserRepository) Update(ctx context.Context, user *domain.User) (*domain.User, error) {
	row, err := r.q.UpdateUser(ctx, usersdb.UpdateUserParams{
		ID:              uuidToPgtype(user.ID),
		FirstName:       user.FirstName,
		LastName:        user.LastName,
		Patronymic:      user.Patronymic,
		Phone:           user.Phone,
		Telegram:        user.Telegram,
		TelegramID:      user.TelegramID,
		Email:           user.Email,
		Gender:          user.Gender,
		Country:         user.Country,
		Region:          user.Region,
		Nationality:     user.Nationality,
		ProfilePicUrl:   user.ProfilePicURL,
		Status:          user.Status,
		TariffType:      user.TariffType,
		JobStatus:       user.JobStatus,
		ActivityType:    user.ActivityType,
		Specializations: user.Specializations,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("update user: %w", err)
	}
	return userToDomain(row), nil
}

func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	err := r.q.DeleteUser(ctx, uuidToPgtype(id))
	if err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (r *UserRepository) GetByPhone(ctx context.Context, phone string) (*domain.User, error) {
	row, err := r.q.GetUserByPhone(ctx, textToPgtype(phone))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by phone: %w", err)
	}
	return userToDomain(row), nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, textToPgtype(email))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("get user by email: %w", err)
	}
	return userToDomain(row), nil
}

func (r *UserRepository) SetPassword(ctx context.Context, id uuid.UUID, hash string) error {
	err := r.q.SetUserPassword(ctx, usersdb.SetUserPasswordParams{
		ID:           uuidToPgtype(id),
		PasswordHash: textToPgtype(hash),
	})
	if err != nil {
		return fmt.Errorf("set user password: %w", err)
	}
	return nil
}

func userToDomain(row usersdb.User) *domain.User {
	return &domain.User{
		ID:              pgtypeToUUID(row.ID),
		FirstName:       row.FirstName,
		LastName:        row.LastName,
		Patronymic:      pgtypeToString(row.Patronymic),
		Phone:           pgtypeToString(row.Phone),
		Telegram:        pgtypeToString(row.Telegram),
		TelegramID:      pgtypeToString(row.TelegramID),
		Email:           pgtypeToString(row.Email),
		Gender:          pgtypeToString(row.Gender),
		Country:         pgtypeToString(row.Country),
		Region:          pgtypeToString(row.Region),
		Nationality:     pgtypeToString(row.Nationality),
		ProfilePicURL:   pgtypeToString(row.ProfilePicUrl),
		Status:          row.Status,
		TariffType:      row.TariffType,
		JobStatus:       pgtypeToString(row.JobStatus),
		ActivityType:    pgtypeToString(row.ActivityType),
		Specializations: row.Specializations,
		PasswordHash:    pgtypeToString(row.PasswordHash),
		CreatedAt:       row.CreatedAt.Time,
	}
}
