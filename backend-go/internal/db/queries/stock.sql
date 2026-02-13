-- name: ListStockBalances :many
SELECT sb.item_id, sb.location_id, sb.qty_on_hand, sb.qty_allocated, sb.updated_at
FROM stock_balance sb
ORDER BY sb.updated_at DESC
LIMIT $1 OFFSET $2;

-- name: InsertOutboxEvent :one
INSERT INTO outbox_events (subject, payload)
VALUES ($1, $2)
RETURNING id, subject, payload, created_at, sent_at;
