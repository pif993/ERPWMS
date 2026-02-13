# Events

## NATS subjects
- `stock.moved`
- `orders.created`
- `orders.allocated`

Events are inserted in `outbox_events` in the same DB transaction, then published by worker.
