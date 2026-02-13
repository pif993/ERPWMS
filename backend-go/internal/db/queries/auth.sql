-- name: CreateRefreshSession :one
INSERT INTO refresh_sessions (user_id, refresh_hash, ua_hash, ip_hash, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetRefreshSessionByHash :one
SELECT * FROM refresh_sessions WHERE refresh_hash = $1;

-- name: RevokeRefreshSessionByHash :exec
UPDATE refresh_sessions SET revoked_at = now() WHERE refresh_hash = $1 AND revoked_at IS NULL;
