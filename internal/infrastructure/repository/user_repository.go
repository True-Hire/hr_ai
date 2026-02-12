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
		ID:            uuidToPgtype(user.ID),
		Phone:         textToPgtype(user.Phone),
		Email:         textToPgtype(user.Email),
		ProfilePicUrl: textToPgtype(user.ProfilePicURL),
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
		ID:            uuidToPgtype(user.ID),
		Phone:         user.Phone,
		Email:         user.Email,
		ProfilePicUrl: user.ProfilePicURL,
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

func userToDomain(row usersdb.User) *domain.User {
	return &domain.User{
		ID:            pgtypeToUUID(row.ID),
		Phone:         pgtypeToString(row.Phone),
		Email:         pgtypeToString(row.Email),
		ProfilePicURL: pgtypeToString(row.ProfilePicUrl),
		CreatedAt:     row.CreatedAt.Time,
	}
}
