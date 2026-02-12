-- name: CreateEducationItem :one
INSERT INTO education_items (id, user_id, institution, degree, field_of_study, start_date, end_date, location, item_order, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
RETURNING id, user_id, institution, degree, field_of_study, start_date, end_date, location, item_order, updated_at;

-- name: GetEducationItemByID :one
SELECT id, user_id, institution, degree, field_of_study, start_date, end_date, location, item_order, updated_at
FROM education_items
WHERE id = $1;

-- name: ListEducationItemsByUser :many
SELECT id, user_id, institution, degree, field_of_study, start_date, end_date, location, item_order, updated_at
FROM education_items
WHERE user_id = $1
ORDER BY item_order;

-- name: UpdateEducationItem :one
UPDATE education_items
SET institution = COALESCE(NULLIF(sqlc.arg(institution), ''), institution),
    degree = COALESCE(NULLIF(sqlc.arg(degree), ''), degree),
    field_of_study = COALESCE(NULLIF(sqlc.arg(field_of_study), ''), field_of_study),
    start_date = COALESCE(NULLIF(sqlc.arg(start_date), ''), start_date),
    end_date = COALESCE(NULLIF(sqlc.arg(end_date), ''), end_date),
    location = COALESCE(NULLIF(sqlc.arg(location), ''), location),
    item_order = sqlc.arg(item_order),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, user_id, institution, degree, field_of_study, start_date, end_date, location, item_order, updated_at;

-- name: DeleteEducationItem :exec
DELETE FROM education_items WHERE id = $1;

-- name: DeleteEducationItemsByUser :exec
DELETE FROM education_items WHERE user_id = $1;
