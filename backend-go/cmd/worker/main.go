package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"erpwms/backend-go/internal/common/config"
	sqlc "erpwms/backend-go/internal/db/sqlcgen"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

func main() {
	cfg, _ := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	db, _ := pgxpool.New(context.Background(), cfg.DBURL)
	q := sqlc.New(db)
	nc, _ := nats.Connect(cfg.NATSURL)
	backoff := time.Second

	for {
		txCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		tx, err := db.Begin(txCtx)
		if err != nil {
			cancel()
			time.Sleep(backoff)
			continue
		}
		qt := q.WithTx(tx)
		events, err := qt.ListPendingOutboxEventsForUpdate(txCtx, 20)
		if err != nil {
			_ = tx.Rollback(txCtx)
			cancel()
			time.Sleep(backoff)
			continue
		}
		if len(events) == 0 {
			_ = tx.Commit(txCtx)
			cancel()
			time.Sleep(2 * time.Second)
			continue
		}

		for _, e := range events {
			pubCancelCtx, pubCancel := context.WithTimeout(context.Background(), 3*time.Second)
			_ = pubCancelCtx
			err := nc.Publish(e.Topic, e.Payload)
			pubCancel()
			if err != nil {
				_ = qt.BumpOutboxAttempt(txCtx, sqlc.BumpOutboxAttemptParams{ID: e.ID, LastError: pgtype.Text{String: err.Error(), Valid: true}})
				logger.Error("publish failed", "id", e.ID.String(), "topic", e.Topic, "err", err)
				backoff = min(backoff*2, 30*time.Second)
				continue
			}
			_ = qt.MarkOutboxSent(txCtx, e.ID)
			logger.Info("published outbox event", "id", e.ID.String(), "topic", e.Topic)
			backoff = time.Second
		}
		_ = tx.Commit(txCtx)
		cancel()
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
