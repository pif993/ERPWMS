-- +goose Up
CREATE OR REPLACE FUNCTION forbid_update_delete()
RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'append-only';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_audit_log_append_only ON audit_log;
CREATE TRIGGER trg_audit_log_append_only
BEFORE UPDATE OR DELETE ON audit_log
FOR EACH ROW
EXECUTE FUNCTION forbid_update_delete();

DROP TRIGGER IF EXISTS trg_stock_ledger_append_only ON stock_ledger;
CREATE TRIGGER trg_stock_ledger_append_only
BEFORE UPDATE OR DELETE ON stock_ledger
FOR EACH ROW
EXECUTE FUNCTION forbid_update_delete();

CREATE OR REPLACE FUNCTION outbox_events_guard_immutable()
RETURNS trigger AS $$
BEGIN
  IF NEW.id IS DISTINCT FROM OLD.id
    OR NEW.topic IS DISTINCT FROM OLD.topic
    OR NEW.payload IS DISTINCT FROM OLD.payload
    OR NEW.created_at IS DISTINCT FROM OLD.created_at THEN
    RAISE EXCEPTION 'outbox immutable fields cannot be modified';
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_outbox_events_guard_immutable ON outbox_events;
CREATE TRIGGER trg_outbox_events_guard_immutable
BEFORE UPDATE ON outbox_events
FOR EACH ROW
EXECUTE FUNCTION outbox_events_guard_immutable();

-- +goose Down
DROP TRIGGER IF EXISTS trg_outbox_events_guard_immutable ON outbox_events;
DROP FUNCTION IF EXISTS outbox_events_guard_immutable();

DROP TRIGGER IF EXISTS trg_stock_ledger_append_only ON stock_ledger;
DROP TRIGGER IF EXISTS trg_audit_log_append_only ON audit_log;
DROP FUNCTION IF EXISTS forbid_update_delete();
