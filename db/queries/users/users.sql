-- name: CreateUser :one
INSERT INTO users (id, phone, email, profile_pic_url, created_at)
VALUES ($1, $2, $3, $4, now())
RETURNING id, phone, email, profile_pic_url, created_at;

-- name: GetUserByID :one
SELECT id, phone, email, profile_pic_url, created_at
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, phone, email, profile_pic_url, created_at
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: UpdateUser :one
UPDATE users
SET phone = COALESCE(NULLIF(sqlc.arg(phone), ''), phone),
    email = COALESCE(NULLIF(sqlc.arg(email), ''), email),
    profile_pic_url = COALESCE(NULLIF(sqlc.arg(profile_pic_url), ''), profile_pic_url)
WHERE id = sqlc.arg(id)
RETURNING id, phone, email, profile_pic_url, created_at;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;
