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
		tx, err := db.Begin(context.Background())
		if err != nil {
			time.Sleep(backoff)
			continue
		}
		qt := q.WithTx(tx)
		events, err := qt.ListPendingOutboxEventsForUpdate(context.Background(), 20)
		if err != nil {
			_ = tx.Rollback(context.Background())
			time.Sleep(backoff)
			continue
		}
		if len(events) == 0 {
			_ = tx.Commit(context.Background())
			time.Sleep(2 * time.Second)
			continue
		}
		for _, e := range events {
			if err := nc.Publish(e.Topic, e.Payload); err != nil {
				_ = qt.BumpOutboxAttempt(context.Background(), sqlc.BumpOutboxAttemptParams{ID: e.ID, LastError: pgtype.Text{String: err.Error(), Valid: true}})
				logger.Error("publish failed", "id", e.ID.String(), "err", err)
				backoff = min(backoff*2, 30*time.Second)
				continue
			}
			_ = qt.MarkOutboxSent(context.Background(), e.ID)
			backoff = time.Second
		}
		_ = tx.Commit(context.Background())
	}
}

func min(a, b time.Duration) time.Duration {
	if a < b {
		return a
	}
	return b
}
