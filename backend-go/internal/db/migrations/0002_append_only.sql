-- +goose Up
CREATE OR REPLACE FUNCTION forbid_update_delete() RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'table % is append-only', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_stock_ledger_append_only
  BEFORE UPDATE OR DELETE ON stock_ledger
  FOR EACH ROW EXECUTE FUNCTION forbid_update_delete();

CREATE TRIGGER tr_audit_log_append_only
  BEFORE UPDATE OR DELETE ON audit_log
  FOR EACH ROW EXECUTE FUNCTION forbid_update_delete();

CREATE OR REPLACE FUNCTION outbox_update_guard() RETURNS trigger AS $$
BEGIN
  IF OLD.id <> NEW.id OR OLD.topic <> NEW.topic OR OLD.payload <> NEW.payload OR OLD.created_at <> NEW.created_at THEN
    RAISE EXCEPTION 'outbox_events immutable columns cannot be updated';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER tr_outbox_guard
  BEFORE UPDATE ON outbox_events
  FOR EACH ROW EXECUTE FUNCTION outbox_update_guard();

-- +goose Down
DROP TRIGGER IF EXISTS tr_outbox_guard ON outbox_events;
DROP FUNCTION IF EXISTS outbox_update_guard();
DROP TRIGGER IF EXISTS tr_stock_ledger_append_only ON stock_ledger;
DROP TRIGGER IF EXISTS tr_audit_log_append_only ON audit_log;
DROP FUNCTION IF EXISTS forbid_update_delete();
