-- name: CreateCountryText :one
INSERT INTO country_texts (
    country_id, lang, name, is_source, model_version, updated_at
) VALUES (
    $1, $2, $3, $4, $5, now()
)
RETURNING country_id, lang, name, is_source, model_version, updated_at;

-- name: GetCountryText :one
SELECT country_id, lang, name, is_source, model_version, updated_at
FROM country_texts
WHERE country_id = $1 AND lang = $2;

-- name: ListCountryTextsByCountry :many
SELECT country_id, lang, name, is_source, model_version, updated_at
FROM country_texts
WHERE country_id = $1;

-- name: UpdateCountryText :one
UPDATE country_texts
SET name = COALESCE(NULLIF(sqlc.arg(name), ''), name),
    is_source = sqlc.arg(is_source),
    model_version = sqlc.arg(model_version),
    updated_at = now()
WHERE country_id = sqlc.arg(country_id) AND lang = sqlc.arg(lang)
RETURNING country_id, lang, name, is_source, model_version, updated_at;

-- name: DeleteCountryText :exec
DELETE FROM country_texts
WHERE country_id = $1 AND lang = $2;

-- name: DeleteCountryTextsByCountry :exec
DELETE FROM country_texts
WHERE country_id = $1;
