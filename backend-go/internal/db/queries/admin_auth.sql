-- name: AdminListUsers :many
SELECT u.id, u.email_hash, u.status, u.created_at
FROM users u
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminListRoles :many
SELECT id, name FROM roles ORDER BY name;

-- name: AdminClearUserRoles :exec
DELETE FROM user_roles WHERE user_id=$1;

-- name: AdminSetUserRole :exec
INSERT INTO user_roles(user_id, role_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING;
