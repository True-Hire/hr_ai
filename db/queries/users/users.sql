-- name: CreateUser :one
INSERT INTO users (
    id, first_name, last_name, patronymic, phone, telegram, telegram_id, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, language, profile_score, 
    main_category_id, sub_category_id, created_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8,
    $9, $10, $11, $12, $13,
    $14, $15, $16, $17, $18, $19, $20,
    $21, $22, now()
)
RETURNING id, first_name, last_name, patronymic, phone, telegram, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, created_at, password_hash, telegram_id, language, profile_score, estimated_salary_min, estimated_salary_max, estimated_salary_currency,
    main_category_id, sub_category_id;

-- name: GetUserByID :one
SELECT id, first_name, last_name, patronymic, phone, telegram, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, created_at, password_hash, telegram_id, language, profile_score, estimated_salary_min, estimated_salary_max, estimated_salary_currency,
    main_category_id, sub_category_id
FROM users
WHERE id = $1;

-- name: ListUsers :many
SELECT id, first_name, last_name, patronymic, phone, telegram, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, created_at, password_hash, telegram_id, language, profile_score, estimated_salary_min, estimated_salary_max, estimated_salary_currency,
    main_category_id, sub_category_id
FROM users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountUsers :one
SELECT count(*) FROM users;

-- name: UpdateUser :one
UPDATE users
SET first_name = COALESCE(NULLIF(sqlc.arg(first_name), ''), first_name),
    last_name = COALESCE(NULLIF(sqlc.arg(last_name), ''), last_name),
    patronymic = COALESCE(NULLIF(sqlc.arg(patronymic), ''), patronymic),
    phone = COALESCE(NULLIF(sqlc.arg(phone), ''), phone),
    telegram = COALESCE(NULLIF(sqlc.arg(telegram), ''), telegram),
    telegram_id = COALESCE(NULLIF(sqlc.arg(telegram_id), ''), telegram_id),
    email = COALESCE(NULLIF(sqlc.arg(email), ''), email),
    gender = COALESCE(NULLIF(sqlc.arg(gender), ''), gender),
    country = COALESCE(NULLIF(sqlc.arg(country), ''), country),
    region = COALESCE(NULLIF(sqlc.arg(region), ''), region),
    nationality = COALESCE(NULLIF(sqlc.arg(nationality), ''), nationality),
    profile_pic_url = COALESCE(NULLIF(sqlc.arg(profile_pic_url), ''), profile_pic_url),
    status = COALESCE(NULLIF(sqlc.arg(status), ''), status),
    tariff_type = COALESCE(NULLIF(sqlc.arg(tariff_type), ''), tariff_type),
    job_status = COALESCE(NULLIF(sqlc.arg(job_status), ''), job_status),
    activity_type = COALESCE(NULLIF(sqlc.arg(activity_type), ''), activity_type),
    specializations = CASE WHEN sqlc.arg(specializations)::TEXT[] = '{}' THEN specializations ELSE sqlc.arg(specializations) END,
    language = COALESCE(NULLIF(sqlc.arg(language), ''), language),
    main_category_id = COALESCE(sqlc.arg(main_category_id), main_category_id),
    sub_category_id = COALESCE(sqlc.arg(sub_category_id), sub_category_id)
WHERE id = sqlc.arg(id)
RETURNING id, first_name, last_name, patronymic, phone, telegram, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, created_at, password_hash, telegram_id, language, profile_score, estimated_salary_min, estimated_salary_max, estimated_salary_currency,
    main_category_id, sub_category_id;

-- name: DeleteUser :exec
DELETE FROM users
WHERE id = $1;

-- name: GetUserByPhone :one
SELECT id, first_name, last_name, patronymic, phone, telegram, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, created_at, password_hash, telegram_id, language, profile_score, estimated_salary_min, estimated_salary_max, estimated_salary_currency,
    main_category_id, sub_category_id
FROM users
WHERE phone = $1;

-- name: GetUserByEmail :one
SELECT id, first_name, last_name, patronymic, phone, telegram, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, created_at, password_hash, telegram_id, language, profile_score, estimated_salary_min, estimated_salary_max, estimated_salary_currency,
    main_category_id, sub_category_id
FROM users
WHERE email = $1;

-- name: GetUserByTelegramID :one
SELECT id, first_name, last_name, patronymic, phone, telegram, email,
    gender, country, region, nationality, profile_pic_url,
    status, tariff_type, job_status, activity_type, specializations, created_at, password_hash, telegram_id, language, profile_score, estimated_salary_min, estimated_salary_max, estimated_salary_currency,
    main_category_id, sub_category_id
FROM users
WHERE telegram_id = $1;

-- name: SetUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: SetProfileScore :exec
UPDATE users SET profile_score = $2 WHERE id = $1;

-- name: SetEstimatedSalary :exec
UPDATE users SET estimated_salary_min = $2, estimated_salary_max = $3, estimated_salary_currency = $4 WHERE id = $1;

-- name: CountMatchingUsers :one
SELECT count(*) FROM users
WHERE main_category_id = $1 AND sub_category_id = $2;
