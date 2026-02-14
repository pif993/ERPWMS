-- name: InsertOutboxEvent :one
INSERT INTO outbox_events (topic, payload)
VALUES ($1, $2)
RETURNING *;

-- name: ListPendingOutboxEventsForUpdate :many
SELECT * FROM outbox_events
WHERE sent_at IS NULL
ORDER BY created_at
FOR UPDATE SKIP LOCKED
LIMIT $1;

-- name: MarkOutboxSent :exec
UPDATE outbox_events SET sent_at = now() WHERE id = $1;

-- name: BumpOutboxAttempt :exec
UPDATE outbox_events
SET attempts = attempts + 1, last_error = $2
WHERE id = $1;
