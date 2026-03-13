package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	hrsessionsdb "github.com/ruziba3vich/hr-ai/db/sqlc/hr_sessions"
	"github.com/ruziba3vich/hr-ai/internal/domain"
)

type HRSessionRepository struct {
	q    *hrsessionsdb.Queries
	pool *pgxpool.Pool
}

func NewHRSessionRepository(pool *pgxpool.Pool) *HRSessionRepository {
	return &HRSessionRepository{
		q:    hrsessionsdb.New(pool),
		pool: pool,
	}
}

func (r *HRSessionRepository) Create(ctx context.Context, session *domain.HRSession) (*domain.HRSession, error) {
	row, err := r.q.CreateHRSession(ctx, hrsessionsdb.CreateHRSessionParams{
		ID:               uuidToPgtype(session.ID),
		HrID:             uuidToPgtype(session.HRID),
		DeviceID:         session.DeviceID,
		RefreshTokenHash: session.RefreshTokenHash,
		FcmToken:         textToPgtype(session.FcmToken),
		IpAddress:        textToPgtype(session.IPAddress),
	})
	if err != nil {
		return nil, fmt.Errorf("create hr session: %w", err)
	}
	return hrSessionToDomain(row), nil
}

func (r *HRSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.HRSession, error) {
	row, err := r.q.GetHRSessionByID(ctx, uuidToPgtype(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHRSessionNotFound
		}
		return nil, fmt.Errorf("get hr session by id: %w", err)
	}
	return hrSessionToDomain(row), nil
}

func (r *HRSessionRepository) GetByDeviceID(ctx context.Context, hrID uuid.UUID, deviceID string) (*domain.HRSession, error) {
	row, err := r.q.GetHRSessionByDeviceID(ctx, hrsessionsdb.GetHRSessionByDeviceIDParams{
		HrID:     uuidToPgtype(hrID),
		DeviceID: deviceID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrHRSessionNotFound
		}
		return nil, fmt.Errorf("get hr session by device id: %w", err)
	}
	return hrSessionToDomain(row), nil
}

func (r *HRSessionRepository) SoftDelete(ctx context.Context, id uuid.UUID) error {
	if err := r.q.SoftDeleteHRSession(ctx, uuidToPgtype(id)); err != nil {
		return fmt.Errorf("soft delete hr session: %w", err)
	}
	return nil
}

func (r *HRSessionRepository) SoftDeleteByHR(ctx context.Context, hrID uuid.UUID) error {
	if err := r.q.SoftDeleteHRSessions(ctx, uuidToPgtype(hrID)); err != nil {
		return fmt.Errorf("soft delete hr sessions: %w", err)
	}
	return nil
}

func (r *HRSessionRepository) HardDeleteByHR(ctx context.Context, hrID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, "DELETE FROM hr_sessions WHERE hr_id = $1", uuidToPgtype(hrID))
	if err != nil {
		return fmt.Errorf("hard delete hr sessions: %w", err)
	}
	return nil
}

func (r *HRSessionRepository) UpdateRefreshToken(ctx context.Context, id uuid.UUID, hash string) error {
	if err := r.q.UpdateHRSessionRefreshToken(ctx, hrsessionsdb.UpdateHRSessionRefreshTokenParams{
		ID:               uuidToPgtype(id),
		RefreshTokenHash: hash,
	}); err != nil {
		return fmt.Errorf("update hr session refresh token: %w", err)
	}
	return nil
}

func hrSessionToDomain(row hrsessionsdb.HrSession) *domain.HRSession {
	return &domain.HRSession{
		ID:               pgtypeToUUID(row.ID),
		HRID:             pgtypeToUUID(row.HrID),
		DeviceID:         row.DeviceID,
		RefreshTokenHash: row.RefreshTokenHash,
		FcmToken:         pgtypeToString(row.FcmToken),
		IPAddress:        pgtypeToString(row.IpAddress),
		CreatedAt:        row.CreatedAt.Time,
		Deleted:          row.Deleted,
	}
}
