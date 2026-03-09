-- name: SaveUser :one
INSERT INTO hr_saved_users (hr_id, user_id, note, created_at)
VALUES ($1, $2, $3, now())
ON CONFLICT (hr_id, user_id) DO UPDATE SET note = EXCLUDED.note
RETURNING hr_id, user_id, note, created_at;

-- name: UnsaveUser :exec
DELETE FROM hr_saved_users WHERE hr_id = $1 AND user_id = $2;

-- name: IsSaved :one
SELECT count(*) FROM hr_saved_users WHERE hr_id = $1 AND user_id = $2;

-- name: CountSavedByHR :one
SELECT count(*) FROM hr_saved_users WHERE hr_id = $1;

-- name: ListSavedByHR :many
SELECT hr_id, user_id, note, created_at
FROM hr_saved_users
WHERE hr_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;
