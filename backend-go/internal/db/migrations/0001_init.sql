-- +goose Up

CREATE EXTENSION IF NOT EXISTS pgcrypto;
CREATE EXTENSION IF NOT EXISTS citext;

CREATE TABLE users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email_hash TEXT UNIQUE NOT NULL,
  email_enc  TEXT NOT NULL,
  email_nonce TEXT NOT NULL,
  email_key_id TEXT NOT NULL,
  password_hash TEXT NOT NULL,
  status TEXT NOT NULL DEFAULT 'active',
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE roles (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE permissions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  name TEXT UNIQUE NOT NULL
);

CREATE TABLE role_permissions (
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  permission_id UUID NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
  PRIMARY KEY (role_id, permission_id)
);

CREATE TABLE user_roles (
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
  PRIMARY KEY (user_id, role_id)
);

CREATE TABLE refresh_sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  refresh_hash TEXT UNIQUE NOT NULL,
  ua_hash TEXT,
  ip_hash TEXT,
  expires_at TIMESTAMPTZ NOT NULL,
  revoked_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE audit_log (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ts TIMESTAMPTZ NOT NULL DEFAULT now(),
  actor_user_id UUID REFERENCES users(id),
  actor_type TEXT NOT NULL,
  action TEXT NOT NULL,
  resource TEXT NOT NULL,
  resource_id TEXT,
  status TEXT NOT NULL,
  ip_hash TEXT,
  ua_hash TEXT,
  request_id TEXT,
  trace_id TEXT,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE warehouses (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  code TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL
);

CREATE TABLE locations (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  warehouse_id UUID NOT NULL REFERENCES warehouses(id),
  code TEXT NOT NULL,
  type TEXT NOT NULL,
  path TEXT,
  UNIQUE (warehouse_id, code)
);

CREATE TABLE items (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  sku TEXT UNIQUE NOT NULL,
  name TEXT NOT NULL,
  barcode TEXT,
  uom TEXT NOT NULL
);

CREATE TABLE stock_ledger (
  move_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  ts TIMESTAMPTZ NOT NULL DEFAULT now(),
  item_id UUID NOT NULL REFERENCES items(id),
  qty NUMERIC NOT NULL,
  from_location_id UUID REFERENCES locations(id),
  to_location_id UUID REFERENCES locations(id),
  reason_code TEXT NOT NULL,
  ref_type TEXT,
  ref_id TEXT,
  actor_user_id UUID REFERENCES users(id),
  request_id TEXT
);

CREATE TABLE stock_balance (
  item_id UUID NOT NULL REFERENCES items(id),
  location_id UUID NOT NULL REFERENCES locations(id),
  qty_on_hand NUMERIC NOT NULL DEFAULT 0,
  qty_allocated NUMERIC NOT NULL DEFAULT 0,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (item_id, location_id)
);

CREATE TABLE outbox_events (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  topic TEXT NOT NULL,
  payload JSONB NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  sent_at TIMESTAMPTZ,
  attempts INT NOT NULL DEFAULT 0,
  last_error TEXT
);

CREATE TABLE idempotency_keys (
  key TEXT NOT NULL,
  endpoint TEXT NOT NULL,
  actor_user_id UUID REFERENCES users(id),
  request_hash TEXT NOT NULL,
  response_json JSONB,
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  PRIMARY KEY (key, endpoint)
);

CREATE INDEX idx_stock_ledger_item_ts ON stock_ledger(item_id, ts DESC);
CREATE INDEX idx_audit_ts ON audit_log(ts DESC);
CREATE INDEX idx_audit_actor_ts ON audit_log(actor_user_id, ts DESC);
CREATE INDEX idx_outbox_pending ON outbox_events(sent_at, created_at);
CREATE INDEX idx_refresh_user_exp ON refresh_sessions(user_id, expires_at);

-- +goose Down
DROP TABLE IF EXISTS idempotency_keys, outbox_events, stock_balance, stock_ledger, items, locations, warehouses,
  audit_log, refresh_sessions, user_roles, role_permissions, permissions, roles, users CASCADE;
