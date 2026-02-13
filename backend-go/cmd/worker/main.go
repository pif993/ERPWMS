package main

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"erpwms/backend-go/internal/common/config"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/nats-io/nats.go"
)

func main() {
	cfg, _ := config.Load()
	db, _ := pgxpool.New(context.Background(), cfg.DBURL)
	nc, _ := nats.Connect(cfg.NATSURL)
	for {
		rows, _ := db.Query(context.Background(), "SELECT id, subject, payload FROM outbox_events WHERE sent_at IS NULL ORDER BY id LIMIT 50")
		for rows.Next() {
			var id int64
			var subject string
			var payload map[string]any
			_ = rows.Scan(&id, &subject, &payload)
			b, _ := json.Marshal(payload)
			if err := nc.Publish(subject, b); err == nil {
				_, _ = db.Exec(context.Background(), "UPDATE outbox_events SET sent_at=now() WHERE id=$1", id)
			}
		}
		rows.Close()
		time.Sleep(2 * time.Second)
		log.Print("worker tick")
	}
}
