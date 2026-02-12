-- name: CreateProfileField :one
INSERT INTO profile_fields (id, user_id, field_name, source_lang, updated_at)
VALUES ($1, $2, $3, $4, now())
RETURNING id, user_id, field_name, source_lang, updated_at;

-- name: GetProfileFieldByID :one
SELECT id, user_id, field_name, source_lang, updated_at
FROM profile_fields
WHERE id = $1;

-- name: GetProfileFieldByUserAndName :one
SELECT id, user_id, field_name, source_lang, updated_at
FROM profile_fields
WHERE user_id = $1 AND field_name = $2;

-- name: ListProfileFieldsByUser :many
SELECT id, user_id, field_name, source_lang, updated_at
FROM profile_fields
WHERE user_id = $1
ORDER BY updated_at DESC;

-- name: UpdateProfileField :one
UPDATE profile_fields
SET field_name = COALESCE(NULLIF(sqlc.arg(field_name), ''), field_name),
    source_lang = COALESCE(NULLIF(sqlc.arg(source_lang), ''), source_lang),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, user_id, field_name, source_lang, updated_at;

-- name: DeleteProfileField :exec
DELETE FROM profile_fields
WHERE id = $1;

-- name: DeleteProfileFieldsByUser :exec
DELETE FROM profile_fields
WHERE user_id = $1;
