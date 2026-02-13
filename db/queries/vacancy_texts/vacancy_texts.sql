-- name: CreateVacancyText :one
INSERT INTO vacancy_texts (
    vacancy_id, lang, title, description, responsibilities,
    requirements, benefits, is_source, model_version, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, now()
)
RETURNING vacancy_id, lang, title, description, responsibilities,
    requirements, benefits, is_source, model_version, updated_at;

-- name: GetVacancyText :one
SELECT vacancy_id, lang, title, description, responsibilities,
    requirements, benefits, is_source, model_version, updated_at
FROM vacancy_texts
WHERE vacancy_id = $1 AND lang = $2;

-- name: ListVacancyTextsByVacancy :many
SELECT vacancy_id, lang, title, description, responsibilities,
    requirements, benefits, is_source, model_version, updated_at
FROM vacancy_texts
WHERE vacancy_id = $1;

-- name: UpdateVacancyText :one
UPDATE vacancy_texts
SET title = COALESCE(NULLIF(sqlc.arg(title), ''), title),
    description = COALESCE(NULLIF(sqlc.arg(description), ''), description),
    responsibilities = COALESCE(NULLIF(sqlc.arg(responsibilities), ''), responsibilities),
    requirements = COALESCE(NULLIF(sqlc.arg(requirements), ''), requirements),
    benefits = COALESCE(NULLIF(sqlc.arg(benefits), ''), benefits),
    is_source = sqlc.arg(is_source),
    model_version = sqlc.arg(model_version),
    updated_at = now()
WHERE vacancy_id = sqlc.arg(vacancy_id) AND lang = sqlc.arg(lang)
RETURNING vacancy_id, lang, title, description, responsibilities,
    requirements, benefits, is_source, model_version, updated_at;

-- name: DeleteVacancyText :exec
DELETE FROM vacancy_texts WHERE vacancy_id = $1 AND lang = $2;

-- name: DeleteVacancyTextsByVacancy :exec
DELETE FROM vacancy_texts WHERE vacancy_id = $1;
