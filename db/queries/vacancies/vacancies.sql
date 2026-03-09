-- name: CreateVacancy :one
INSERT INTO vacancies (
    id, hr_id, company_data, country_id, salary_min, salary_max, salary_currency,
    experience_min, experience_max, format, schedule,
    phone, telegram, email, address, status, source_lang, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11,
    $12, $13, $14, $15, $16, $17, now()
)
RETURNING id, hr_id, company_data, country_id, salary_min, salary_max, salary_currency,
    experience_min, experience_max, format, schedule,
    phone, telegram, email, address, status, source_lang, created_at;

-- name: GetVacancyByID :one
SELECT id, hr_id, company_data, country_id, salary_min, salary_max, salary_currency,
    experience_min, experience_max, format, schedule,
    phone, telegram, email, address, status, source_lang, created_at
FROM vacancies
WHERE id = $1;

-- name: ListVacancies :many
SELECT id, hr_id, company_data, country_id, salary_min, salary_max, salary_currency,
    experience_min, experience_max, format, schedule,
    phone, telegram, email, address, status, source_lang, created_at
FROM vacancies
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListVacanciesByHR :many
SELECT id, hr_id, company_data, country_id, salary_min, salary_max, salary_currency,
    experience_min, experience_max, format, schedule,
    phone, telegram, email, address, status, source_lang, created_at
FROM vacancies
WHERE hr_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVacancies :one
SELECT count(*) FROM vacancies;

-- name: CountVacanciesByHR :one
SELECT count(*) FROM vacancies WHERE hr_id = $1;

-- name: UpdateVacancy :one
UPDATE vacancies
SET salary_min = CASE WHEN sqlc.arg(salary_min)::INT = 0 THEN salary_min ELSE sqlc.arg(salary_min) END,
    salary_max = CASE WHEN sqlc.arg(salary_max)::INT = 0 THEN salary_max ELSE sqlc.arg(salary_max) END,
    salary_currency = COALESCE(NULLIF(sqlc.arg(salary_currency), ''), salary_currency),
    experience_min = CASE WHEN sqlc.arg(experience_min)::INT = 0 THEN experience_min ELSE sqlc.arg(experience_min) END,
    experience_max = CASE WHEN sqlc.arg(experience_max)::INT = 0 THEN experience_max ELSE sqlc.arg(experience_max) END,
    format = COALESCE(NULLIF(sqlc.arg(format), ''), format),
    schedule = COALESCE(NULLIF(sqlc.arg(schedule), ''), schedule),
    phone = COALESCE(NULLIF(sqlc.arg(phone), ''), phone),
    telegram = COALESCE(NULLIF(sqlc.arg(telegram), ''), telegram),
    email = COALESCE(NULLIF(sqlc.arg(email), ''), email),
    address = COALESCE(NULLIF(sqlc.arg(address), ''), address),
    status = COALESCE(NULLIF(sqlc.arg(status), ''), status),
    country_id = COALESCE(sqlc.arg(country_id), country_id)
WHERE id = sqlc.arg(id)
RETURNING id, hr_id, company_data, country_id, salary_min, salary_max, salary_currency,
    experience_min, experience_max, format, schedule,
    phone, telegram, email, address, status, source_lang, created_at;

-- name: SearchVacancies :many
SELECT DISTINCT v.id, v.hr_id, v.company_data, v.country_id, v.salary_min, v.salary_max, v.salary_currency,
    v.experience_min, v.experience_max, v.format, v.schedule,
    v.phone, v.telegram, v.email, v.address, v.status, v.source_lang, v.created_at
FROM vacancies v
JOIN vacancy_texts vt ON vt.vacancy_id = v.id
WHERE vt.lang = sqlc.arg(lang)
  AND (vt.title ILIKE '%' || sqlc.arg(query) || '%' OR vt.description ILIKE '%' || sqlc.arg(query) || '%')
ORDER BY v.created_at DESC
LIMIT sqlc.arg(lim) OFFSET sqlc.arg(off);

-- name: CountSearchVacancies :one
SELECT COUNT(DISTINCT v.id) FROM vacancies v
JOIN vacancy_texts vt ON vt.vacancy_id = v.id
WHERE vt.lang = sqlc.arg(lang)
  AND (vt.title ILIKE '%' || sqlc.arg(query) || '%' OR vt.description ILIKE '%' || sqlc.arg(query) || '%');

-- name: DeleteVacancy :exec
DELETE FROM vacancies WHERE id = $1;

-- name: NullifyVacancyHRID :exec
UPDATE vacancies SET hr_id = NULL WHERE hr_id = $1;

-- name: ListVacancyIDsByHR :many
SELECT id FROM vacancies WHERE hr_id = $1;
