-- name: CreateItemText :one
INSERT INTO item_texts (item_id, item_type, lang, description, is_source, model_version, updated_at)
VALUES ($1, $2, $3, $4, $5, $6, now())
RETURNING item_id, item_type, lang, description, is_source, model_version, updated_at;

-- name: GetItemText :one
SELECT item_id, item_type, lang, description, is_source, model_version, updated_at
FROM item_texts
WHERE item_id = $1 AND item_type = $2 AND lang = $3;

-- name: ListItemTextsByItem :many
SELECT item_id, item_type, lang, description, is_source, model_version, updated_at
FROM item_texts
WHERE item_id = $1 AND item_type = $2
ORDER BY is_source DESC, lang;

-- name: UpdateItemText :one
UPDATE item_texts
SET description = sqlc.arg(description),
    model_version = sqlc.arg(model_version),
    updated_at = now()
WHERE item_id = sqlc.arg(item_id) AND item_type = sqlc.arg(item_type) AND lang = sqlc.arg(lang)
RETURNING item_id, item_type, lang, description, is_source, model_version, updated_at;

-- name: DeleteItemText :exec
DELETE FROM item_texts WHERE item_id = $1 AND item_type = $2 AND lang = $3;

-- name: DeleteItemTextsByItem :exec
DELETE FROM item_texts WHERE item_id = $1 AND item_type = $2;

-- name: DeleteItemTextsByItemID :exec
DELETE FROM item_texts WHERE item_id = $1;
