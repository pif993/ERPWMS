-- Runs at first container init (docker-entrypoint-initdb.d)
-- Requires POSTGRES_SUPER_USER to be superuser inside container.

DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_roles WHERE rolname = 'erp_app') THEN
    CREATE ROLE erp_app LOGIN PASSWORD 'change-me-app';
  END IF;
END $$;

-- Ensure database exists (usually already created via POSTGRES_DB)
DO $$
BEGIN
  IF NOT EXISTS (SELECT 1 FROM pg_database WHERE datname = 'erpwms') THEN
    CREATE DATABASE erpwms;
  END IF;
END $$;

\connect erpwms

GRANT CONNECT, TEMP ON DATABASE erpwms TO erp_app;
GRANT USAGE, CREATE ON SCHEMA public TO erp_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO erp_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
  GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO erp_app;
