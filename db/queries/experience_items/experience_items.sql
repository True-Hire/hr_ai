-- name: CreateExperienceItem :one
INSERT INTO experience_items (id, user_id, company, position, start_date, end_date, projects, web_site, item_order, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, now())
RETURNING id, user_id, company, position, start_date, end_date, projects, web_site, item_order, updated_at;

-- name: GetExperienceItemByID :one
SELECT id, user_id, company, position, start_date, end_date, projects, web_site, item_order, updated_at
FROM experience_items
WHERE id = $1;

-- name: ListExperienceItemsByUser :many
SELECT id, user_id, company, position, start_date, end_date, projects, web_site, item_order, updated_at
FROM experience_items
WHERE user_id = $1
ORDER BY item_order;

-- name: UpdateExperienceItem :one
UPDATE experience_items
SET company = COALESCE(NULLIF(sqlc.arg(company), ''), company),
    position = COALESCE(NULLIF(sqlc.arg(position), ''), position),
    start_date = COALESCE(NULLIF(sqlc.arg(start_date), ''), start_date),
    end_date = COALESCE(NULLIF(sqlc.arg(end_date), ''), end_date),
    projects = COALESCE(NULLIF(sqlc.arg(projects), ''), projects),
    web_site = COALESCE(NULLIF(sqlc.arg(web_site), ''), web_site),
    item_order = sqlc.arg(item_order),
    updated_at = now()
WHERE id = sqlc.arg(id)
RETURNING id, user_id, company, position, start_date, end_date, projects, web_site, item_order, updated_at;

-- name: DeleteExperienceItem :exec
DELETE FROM experience_items WHERE id = $1;

-- name: DeleteExperienceItemsByUser :exec
DELETE FROM experience_items WHERE user_id = $1;
