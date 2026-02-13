DO $$
BEGIN
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'erp_app') THEN
    CREATE ROLE erp_app LOGIN PASSWORD 'change-me-app';
  END IF;
  IF NOT EXISTS (SELECT FROM pg_roles WHERE rolname = 'erp_analytics') THEN
    CREATE ROLE erp_analytics LOGIN PASSWORD 'change-me-analytics';
  END IF;
END$$;

GRANT CONNECT ON DATABASE erpwms TO erp_app, erp_analytics;
