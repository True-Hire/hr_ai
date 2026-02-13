-- name: CreateSession :one
INSERT INTO user_sessions (id, user_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted)
VALUES ($1, $2, $3, $4, $5, $6, now(), false)
RETURNING id, user_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted;

-- name: GetSessionByID :one
SELECT id, user_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted
FROM user_sessions
WHERE id = $1 AND deleted = false;

-- name: GetSessionByDeviceID :one
SELECT id, user_id, device_id, refresh_token_hash, fcm_token, ip_address, created_at, deleted
FROM user_sessions
WHERE user_id = $1 AND device_id = $2 AND deleted = false;

-- name: SoftDeleteSession :exec
UPDATE user_sessions SET deleted = true WHERE id = $1;

-- name: SoftDeleteUserSessions :exec
UPDATE user_sessions SET deleted = true WHERE user_id = $1;

-- name: UpdateSessionRefreshToken :exec
UPDATE user_sessions SET refresh_token_hash = $2 WHERE id = $1;
