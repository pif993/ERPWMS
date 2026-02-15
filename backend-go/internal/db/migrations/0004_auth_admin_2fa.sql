-- +goose Up

-- password reset tokens (hash-only)
CREATE TABLE IF NOT EXISTS password_reset_tokens (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  token_hash BYTEA NOT NULL UNIQUE,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- user mfa (TOTP)
CREATE TABLE IF NOT EXISTS user_mfa (
  user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  enabled BOOLEAN NOT NULL DEFAULT false,
  secret_enc BYTEA NOT NULL,
  secret_nonce BYTEA NOT NULL,
  secret_key_id TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- admin perms baseline
INSERT INTO permissions(name) VALUES
  ('admin.users.read'),
  ('admin.users.write'),
  ('admin.roles.read'),
  ('admin.roles.write'),
  ('admin.mfa.write')
ON CONFLICT DO NOTHING;

-- superadmin role
INSERT INTO roles(name) VALUES ('SuperAdmin') ON CONFLICT DO NOTHING;

-- SuperAdmin gets ALL permissions (current + future)
INSERT INTO role_permissions(role_id, permission_id)
SELECT r.id, p.id
FROM roles r
JOIN permissions p ON true
WHERE r.name='SuperAdmin'
ON CONFLICT DO NOTHING;

-- +goose Down
DROP TABLE IF EXISTS user_mfa;
DROP TABLE IF EXISTS password_reset_tokens;
DELETE FROM role_permissions rp USING roles r WHERE rp.role_id=r.id AND r.name='SuperAdmin';
DELETE FROM roles WHERE name='SuperAdmin';
DELETE FROM permissions WHERE name IN ('admin.users.read','admin.users.write','admin.roles.read','admin.roles.write','admin.mfa.write');
