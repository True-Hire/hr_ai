-- name: CreateCompanyText :one
INSERT INTO company_texts (
    company_id, lang, name, activity_type, company_type, about, market,
    is_source, model_version, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, now()
)
RETURNING company_id, lang, name, activity_type, company_type, about, market,
    is_source, model_version, updated_at;

-- name: GetCompanyText :one
SELECT company_id, lang, name, activity_type, company_type, about, market,
    is_source, model_version, updated_at
FROM company_texts
WHERE company_id = $1 AND lang = $2;

-- name: ListCompanyTextsByCompany :many
SELECT company_id, lang, name, activity_type, company_type, about, market,
    is_source, model_version, updated_at
FROM company_texts
WHERE company_id = $1;

-- name: UpdateCompanyText :one
UPDATE company_texts
SET name = COALESCE(NULLIF(sqlc.arg(name), ''), name),
    activity_type = COALESCE(NULLIF(sqlc.arg(activity_type), ''), activity_type),
    company_type = COALESCE(NULLIF(sqlc.arg(company_type), ''), company_type),
    about = COALESCE(NULLIF(sqlc.arg(about), ''), about),
    market = COALESCE(NULLIF(sqlc.arg(market), ''), market),
    is_source = sqlc.arg(is_source),
    model_version = sqlc.arg(model_version),
    updated_at = now()
WHERE company_id = sqlc.arg(company_id) AND lang = sqlc.arg(lang)
RETURNING company_id, lang, name, activity_type, company_type, about, market,
    is_source, model_version, updated_at;

-- name: DeleteCompanyText :exec
DELETE FROM company_texts
WHERE company_id = $1 AND lang = $2;

-- name: DeleteCompanyTextsByCompany :exec
DELETE FROM company_texts
WHERE company_id = $1;
