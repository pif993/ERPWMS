-- name: GetIdempotency :one
SELECT * FROM idempotency_keys WHERE key = $1 AND endpoint = $2;

-- name: InsertIdempotency :exec
INSERT INTO idempotency_keys (key, endpoint, actor_user_id, request_hash, response_json)
VALUES ($1, $2, $3, $4, $5);
