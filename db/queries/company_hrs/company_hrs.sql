-- name: CreateCompanyHR :one
INSERT INTO company_hrs (
    id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, company_id, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, now()
)
RETURNING id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, created_at, company_id;

-- name: GetCompanyHRByID :one
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, created_at, company_id
FROM company_hrs
WHERE id = $1;

-- name: ListCompanyHRs :many
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, created_at, company_id
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
    company_id = CASE WHEN sqlc.narg(company_id)::UUID IS NOT NULL THEN sqlc.narg(company_id) ELSE company_id END
WHERE id = sqlc.arg(id)
RETURNING id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, created_at, company_id;

-- name: DeleteCompanyHR :exec
DELETE FROM company_hrs
WHERE id = $1;

-- name: GetCompanyHRByPhone :one
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, created_at, company_id
FROM company_hrs
WHERE phone = $1;

-- name: GetCompanyHRByEmail :one
SELECT id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    position, status, password_hash, created_at, company_id
FROM company_hrs
WHERE email = $1;

-- name: SetCompanyHRPassword :exec
UPDATE company_hrs SET password_hash = $2 WHERE id = $1;
