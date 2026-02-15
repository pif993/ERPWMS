-- name: AuthGetUserByEmail :one
SELECT u.id,u.status,uc.email,uc.password_hash,uc.totp_enabled,uc.totp_secret
FROM user_credentials uc JOIN users u ON u.id=uc.user_id
WHERE uc.email=$1 LIMIT 1;

-- name: AuthUpsertCredentials :exec
INSERT INTO user_credentials(user_id,email,password_hash,updated_at)
VALUES($1,$2,$3,now())
ON CONFLICT(email) DO UPDATE SET user_id=EXCLUDED.user_id,password_hash=EXCLUDED.password_hash,updated_at=now();
