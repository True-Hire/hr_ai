-- name: CreateHRSession :one
INSERT INTO hr_sessions (id, hr_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted)
VALUES ($1, $2, $3, $4, $5, $6, now(), false)
RETURNING id, hr_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted;

-- name: GetHRSessionByID :one
SELECT id, hr_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted
FROM hr_sessions
WHERE id = $1 AND deleted = false;

-- name: GetHRSessionByDeviceID :one
SELECT id, hr_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted
FROM hr_sessions
WHERE hr_id = $1 AND device_id = $2 AND deleted = false;

-- name: SoftDeleteHRSession :exec
UPDATE hr_sessions SET deleted = true WHERE id = $1;

-- name: SoftDeleteHRSessions :exec
UPDATE hr_sessions SET deleted = true WHERE hr_id = $1;

-- name: UpdateHRSessionRefreshToken :exec
UPDATE hr_sessions SET refresh_token_hash = $2 WHERE id = $1;
