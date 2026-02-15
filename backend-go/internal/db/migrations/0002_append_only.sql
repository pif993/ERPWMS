-- +goose Up

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION forbid_update_delete()
RETURNS trigger AS $$
BEGIN
  RAISE EXCEPTION 'append-only table: %', TG_TABLE_NAME;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

-- audit_log immutable
DROP TRIGGER IF EXISTS trg_audit_log_no_update ON audit_log;
CREATE TRIGGER trg_audit_log_no_update
BEFORE UPDATE OR DELETE ON audit_log
FOR EACH ROW EXECUTE FUNCTION forbid_update_delete();

-- stock_ledger immutable
DROP TRIGGER IF EXISTS trg_stock_ledger_no_update ON stock_ledger;
CREATE TRIGGER trg_stock_ledger_no_update
BEFORE UPDATE OR DELETE ON stock_ledger
FOR EACH ROW EXECUTE FUNCTION forbid_update_delete();

-- +goose StatementBegin
CREATE OR REPLACE FUNCTION outbox_events_only_state_mutable()
RETURNS trigger AS $$
BEGIN
  IF NEW.topic IS DISTINCT FROM OLD.topic THEN
    RAISE EXCEPTION 'outbox_events.topic is immutable';
  END IF;
  IF NEW.payload IS DISTINCT FROM OLD.payload THEN
    RAISE EXCEPTION 'outbox_events.payload is immutable';
  END IF;
  IF NEW.created_at IS DISTINCT FROM OLD.created_at THEN
    RAISE EXCEPTION 'outbox_events.created_at is immutable';
  END IF;
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;
-- +goose StatementEnd

DROP TRIGGER IF EXISTS trg_outbox_events_state_only ON outbox_events;
CREATE TRIGGER trg_outbox_events_state_only
BEFORE UPDATE ON outbox_events
FOR EACH ROW EXECUTE FUNCTION outbox_events_only_state_mutable();

-- +goose Down
DROP TRIGGER IF EXISTS trg_outbox_events_state_only ON outbox_events;
DROP FUNCTION IF EXISTS outbox_events_only_state_mutable();
DROP TRIGGER IF EXISTS trg_stock_ledger_no_update ON stock_ledger;
DROP TRIGGER IF EXISTS trg_audit_log_no_update ON audit_log;
DROP FUNCTION IF EXISTS forbid_update_delete();
