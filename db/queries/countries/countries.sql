-- name: CreateCountry :one
INSERT INTO countries (id, name, short_code, created_at)
VALUES ($1, $2, $3, now())
RETURNING id, name, short_code, created_at;

-- name: GetCountryByID :one
SELECT id, name, short_code, created_at
FROM countries
WHERE id = $1;

-- name: GetCountryByShortCode :one
SELECT id, name, short_code, created_at
FROM countries
WHERE short_code = $1;

-- name: ListCountries :many
SELECT id, name, short_code, created_at
FROM countries
ORDER BY name;

-- name: CountCountries :one
SELECT count(*) FROM countries;

-- name: UpdateCountry :one
UPDATE countries
SET name = COALESCE(NULLIF(sqlc.arg(name), ''), name),
    short_code = COALESCE(NULLIF(sqlc.arg(short_code), ''), short_code)
WHERE id = sqlc.arg(id)
RETURNING id, name, short_code, created_at;

-- name: DeleteCountry :exec
DELETE FROM countries WHERE id = $1;
