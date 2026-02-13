package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	sessionsdb "github.com/ruziba3vich/hr-ai/db/sqlc/sessions"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type SessionRepository struct {
	q *sessionsdb.Queries
}

func NewSessionRepository(pool *pgxpool.Pool) *SessionRepository {
	return &SessionRepository{
		q: sessionsdb.New(pool),
	}
}

func (r *SessionRepository) Create(ctx context.Context, session *domain.UserSession) (*domain.UserSession, error) {
	row, err := r.q.CreateSession(ctx, sessionsdb.CreateSessionParams{
		ID:               uuidToPgtype(session.ID),
		UserID:           uuidToPgtype(session.UserID),
		DeviceID:         session.DeviceID,
		RefreshTokenHash: session.RefreshTokenHash,
		FcmToken:         textToPgtype(session.FcmToken),
		IpAddress:        textToPgtype(session.IPAddress),
	})
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}
	return sessionToDomain(row), nil
}

func (r *SessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.UserSession, error) {
	row, err := r.q.GetSessionByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session by id: %w", err)
	}
	return sessionToDomain(row), nil
}

func (r *SessionRepository) GetByDeviceID(ctx context.Context, userID uuid.UUID, deviceID string) (*domain.UserSession, error) {
	row, err := r.q.GetSessionByDeviceID(ctx, sessionsdb.GetSessionByDeviceIDParams{
		UserID:   uuidToPgtype(userID),
		DeviceID: deviceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrSessionNotFound
		}
		return nil, fmt.Errorf("get session by device id: %w", err)
	}
	return sessionToDomain(row), nil
}

func (r *SessionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.SoftDeleteSession(ctx, uuidToPgtype(id)); err != nil {
		return fmt.Errorf("soft delete session: %w", err)
	}
	return nil
}

func (r *SessionRepository) SoftDeleteByUser(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.SoftDeleteUserSessions(ctx, uuidToPgtype(userID)); err != nil {
		return fmt.Errorf("soft delete user sessions: %w", err)
	}
	return nil
}

func (r *SessionRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, hash string) error {
	if err := r.q.UpdateSessionRefreshToken(ctx, sessionsdb.UpdateSessionRefreshTokenParams{
		ID:               uuidToPgtype(id),
		RefreshTokenHash: hash,
	}); err != nil {
		return fmt.Errorf("update session refresh token: %w", err)
	}
	return nil
}

func sessionToDomain(row sessionsdb.UserSession) *domain.UserSession {
	return &domain.UserSession{
		ID:               pgtypeToUUID(row.ID),
		UserID:           pgtypeToUUID(row.UserID),
		DeviceID:         row.DeviceID,
		RefreshTokenHash: row.RefreshTokenHash,
		FcmToken:         pgtypeToString(row.FcmToken),
		IPAddress:        pgtypeToString(row.IpAddress),
		CreatedAt:        row.CreatedAt.Time,
		Deleted:          row.Deleted,
	}
}
