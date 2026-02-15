-- +goose Up
CREATE TABLE IF NOT EXISTS user_credentials(
  user_id uuid PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
  email text NOT NULL UNIQUE,
  password_hash text NOT NULL,
  totp_secret text NULL,
  totp_enabled boolean NOT NULL DEFAULT false,
  reset_token_hash text NULL,
  reset_token_expires_at timestamptz NULL,
  updated_at timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_user_credentials_email ON user_credentials(email);

-- +goose Down
DROP TABLE IF EXISTS user_credentials;
