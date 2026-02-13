-- name: UpsertSkill :one
INSERT INTO skills (id, name)
VALUES ($1, $2)
ON CONFLICT (name) DO UPDATE SET name = EXCLUDED.name
RETURNING id, name, created_at;

-- name: GetSkillByName :one
SELECT id, name, created_at FROM skills WHERE name = $1;

-- name: GetSkillByID :one
SELECT id, name, created_at FROM skills WHERE id = $1;

-- name: ListSkills :many
SELECT id, name, created_at FROM skills ORDER BY name;

-- name: SearchSkills :many
SELECT id, name, created_at FROM skills WHERE name LIKE $1 ORDER BY name LIMIT 20;

-- name: AddUserSkill :exec
INSERT INTO user_skills (user_id, skill_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: RemoveUserSkills :exec
DELETE FROM user_skills WHERE user_id = $1;

-- name: ListUserSkills :many
SELECT s.id, s.name, s.created_at
FROM skills s
JOIN user_skills us ON us.skill_id = s.id
WHERE us.user_id = $1
ORDER BY s.name;
