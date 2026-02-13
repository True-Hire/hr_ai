-- name: CreateCompany :one
INSERT INTO companies (
    id, employee_count, country, address, phone, telegram, telegram_channel,
    email, logo_url, web_site, instagram, source_lang, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12, now()
)
RETURNING id, employee_count, country, address, phone, telegram, telegram_channel,
    email, logo_url, web_site, instagram, source_lang, created_at;

-- name: GetCompanyByID :one
SELECT id, employee_count, country, address, phone, telegram, telegram_channel,
    email, logo_url, web_site, instagram, source_lang, created_at
FROM companies
WHERE id = $1;

-- name: ListCompanies :many
SELECT id, employee_count, country, address, phone, telegram, telegram_channel,
    email, logo_url, web_site, instagram, source_lang, created_at
FROM companies
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCompanies :one
SELECT count(*) FROM companies;

-- name: UpdateCompany :one
UPDATE companies
SET employee_count = CASE WHEN sqlc.arg(employee_count)::INT = 0 THEN employee_count ELSE sqlc.arg(employee_count) END,
    country = COALESCE(NULLIF(sqlc.arg(country), ''), country),
    address = COALESCE(NULLIF(sqlc.arg(address), ''), address),
    phone = COALESCE(NULLIF(sqlc.arg(phone), ''), phone),
    telegram = COALESCE(NULLIF(sqlc.arg(telegram), ''), telegram),
    telegram_channel = COALESCE(NULLIF(sqlc.arg(telegram_channel), ''), telegram_channel),
    email = COALESCE(NULLIF(sqlc.arg(email), ''), email),
    logo_url = COALESCE(NULLIF(sqlc.arg(logo_url), ''), logo_url),
    web_site = COALESCE(NULLIF(sqlc.arg(web_site), ''), web_site),
    instagram = COALESCE(NULLIF(sqlc.arg(instagram), ''), instagram)
WHERE id = sqlc.arg(id)
RETURNING id, employee_count, country, address, phone, telegram, telegram_channel,
    email, logo_url, web_site, instagram, source_lang, created_at;

-- name: DeleteCompany :exec
DELETE FROM companies
WHERE id = $1;
