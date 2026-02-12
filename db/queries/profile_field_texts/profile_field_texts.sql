-- name: CreateProfileFieldText :one
INSERT INTO profile_field_texts (profile_field_id, lang, content, is_source, model_version, updated_at)
VALUES ($1, $2, $3, $4, $5, now())
RETURNING profile_field_id, lang, content, is_source, model_version, updated_at;

-- name: GetProfileFieldText :one
SELECT profile_field_id, lang, content, is_source, model_version, updated_at
FROM profile_field_texts
WHERE profile_field_id = $1 AND lang = $2;

-- name: ListProfileFieldTexts :many
SELECT profile_field_id, lang, content, is_source, model_version, updated_at
FROM profile_field_texts
WHERE profile_field_id = $1
ORDER BY is_source DESC, lang;

-- name: UpdateProfileFieldText :one
UPDATE profile_field_texts
SET content = sqlc.arg(content),
    model_version = sqlc.arg(model_version),
    updated_at = now()
WHERE profile_field_id = sqlc.arg(profile_field_id) AND lang = sqlc.arg(lang)
RETURNING profile_field_id, lang, content, is_source, model_version, updated_at;

-- name: DeleteProfileFieldText :exec
DELETE FROM profile_field_texts
WHERE profile_field_id = $1 AND lang = $2;

-- name: DeleteProfileFieldTextsByField :exec
DELETE FROM profile_field_texts
WHERE profile_field_id = $1;
