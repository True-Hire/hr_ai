-- name: CreateCompanyHR :one
INSERT INTO company_hrs (
    id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, company_name, activity_type, company_type, employee_count,
    country, market, web_site, about, logo_url, instagram, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20, now()
)
RETURNING id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, company_name, activity_type, company_type, employee_count,
    country, market, web_site, about, logo_url, instagram, created_at;

-- name: GetCompanyHRByID :one
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, company_name, activity_type, company_type, employee_count,
    country, market, web_site, about, logo_url, instagram, created_at
FROM company_hrs
WHERE id = $1;

-- name: ListCompanyHRs :many
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, company_name, activity_type, company_type, employee_count,
    country, market, web_site, about, logo_url, instagram, created_at
FROM company_hrs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountCompanyHRs :one
SELECT count(*) FROM company_hrs;

-- name: UpdateCompanyHR :one
UPDATE company_hrs
SET first_name = COALESCE(NULLIF(sqlc.arg(first_name), ''), first_name),
    last_name = COALESCE(NULLIF(sqlc.arg(last_name), ''), last_name),
    patronymic = COALESCE(NULLIF(sqlc.arg(patronymic), ''), patronymic),
    phone = COALESCE(NULLIF(sqlc.arg(phone), ''), phone),
    telegram = COALESCE(NULLIF(sqlc.arg(telegram), ''), telegram),
    telegram_id = COALESCE(NULLIF(sqlc.arg(telegram_id), ''), telegram_id),
    email = COALESCE(NULLIF(sqlc.arg(email), ''), email),
    position = COALESCE(NULLIF(sqlc.arg(position), ''), position),
    status = COALESCE(NULLIF(sqlc.arg(status), ''), status),
    company_name = COALESCE(NULLIF(sqlc.arg(company_name), ''), company_name),
    activity_type = COALESCE(NULLIF(sqlc.arg(activity_type), ''), activity_type),
    company_type = COALESCE(NULLIF(sqlc.arg(company_type), ''), company_type),
    employee_count = CASE WHEN sqlc.arg(employee_count)::INT = 0 THEN employee_count ELSE sqlc.arg(employee_count) END,
    country = COALESCE(NULLIF(sqlc.arg(country), ''), country),
    market = COALESCE(NULLIF(sqlc.arg(market), ''), market),
    web_site = COALESCE(NULLIF(sqlc.arg(web_site), ''), web_site),
    about = COALESCE(NULLIF(sqlc.arg(about), ''), about),
    logo_url = COALESCE(NULLIF(sqlc.arg(logo_url), ''), logo_url),
    instagram = COALESCE(NULLIF(sqlc.arg(instagram), ''), instagram)
WHERE id = sqlc.arg(id)
RETURNING id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, company_name, activity_type, company_type, employee_count,
    country, market, web_site, about, logo_url, instagram, created_at;

-- name: DeleteCompanyHR :exec
DELETE FROM company_hrs
WHERE id = $1;

-- name: GetCompanyHRByPhone :one
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, company_name, activity_type, company_type, employee_count,
    country, market, web_site, about, logo_url, instagram, created_at
FROM company_hrs
WHERE phone = $1;

-- name: GetCompanyHRByEmail :one
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, company_name, activity_type, company_type, employee_count,
    country, market, web_site, about, logo_url, instagram, created_at
FROM company_hrs
WHERE email = $1;

-- name: SetCompanyHRPassword :exec
UPDATE company_hrs SET password_hash = $2 WHERE id = $1;
