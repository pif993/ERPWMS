-- name: GetUserByEmailHash :one
SELECT * FROM users WHERE email_hash = $1;

-- name: CreateUser :one
INSERT INTO users (email_hash, email_enc, email_nonce, email_key_id, password_hash, status)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1;
