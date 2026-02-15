-- name: CreateVacancyApplication :one
INSERT INTO vacancy_applications (
    id, user_id, vacancy_id, status, cover_letter, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, now(), now()
)
RETURNING id, user_id, vacancy_id, status, cover_letter, created_at, updated_at;

-- name: GetVacancyApplicationByID :one
SELECT id, user_id, vacancy_id, status, cover_letter, created_at, updated_at
FROM vacancy_applications
WHERE id = $1;

-- name: GetVacancyApplicationByUserAndVacancy :one
SELECT id, user_id, vacancy_id, status, cover_letter, created_at, updated_at
FROM vacancy_applications
WHERE user_id = $1 AND vacancy_id = $2;

-- name: ListVacancyApplicationsByUser :many
SELECT id, user_id, vacancy_id, status, cover_letter, created_at, updated_at
FROM vacancy_applications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListVacancyApplicationsByVacancy :many
SELECT id, user_id, vacancy_id, status, cover_letter, created_at, updated_at
FROM vacancy_applications
WHERE vacancy_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountVacancyApplicationsByUser :one
SELECT count(*) FROM vacancy_applications WHERE user_id = $1;

-- name: CountVacancyApplicationsByVacancy :one
SELECT count(*) FROM vacancy_applications WHERE vacancy_id = $1;

-- name: UpdateVacancyApplicationStatus :one
UPDATE vacancy_applications
SET status = $2, updated_at = now()
WHERE id = $1
RETURNING id, user_id, vacancy_id, status, cover_letter, created_at, updated_at;

-- name: DeleteVacancyApplication :exec
DELETE FROM vacancy_applications WHERE id = $1;
