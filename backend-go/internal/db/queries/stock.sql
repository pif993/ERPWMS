-- name: ListStockBalances :many
SELECT sb.item_id, sb.location_id, sb.qty_on_hand, sb.qty_allocated, sb.updated_at,
       i.sku, i.name item_name, l.code location_code, w.code warehouse_code
FROM stock_balance sb
JOIN items i ON i.id = sb.item_id
JOIN locations l ON l.id = sb.location_id
JOIN warehouses w ON w.id = l.warehouse_id
WHERE ($1::text = '' OR i.sku ILIKE '%' || $1 || '%' OR i.name ILIKE '%' || $1 || '%')
  AND ($2::text = '' OR w.code = $2)
  AND ($3::text = '' OR l.code = $3)
ORDER BY i.sku, l.code
LIMIT $4 OFFSET $5;

-- name: InsertStockLedgerMove :one
INSERT INTO stock_ledger (
  item_id, qty, from_location_id, to_location_id, reason_code, ref_type, ref_id, actor_user_id, request_id
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
RETURNING *;

-- name: UpsertStockBalanceDelta :exec
INSERT INTO stock_balance (item_id, location_id, qty_on_hand, qty_allocated)
VALUES ($1, $2, $3, $4)
ON CONFLICT (item_id, location_id)
DO UPDATE SET
  qty_on_hand = stock_balance.qty_on_hand + EXCLUDED.qty_on_hand,
  qty_allocated = stock_balance.qty_allocated + EXCLUDED.qty_allocated,
  updated_at = now();
