package main

import (
	"context"
	"encoding/base64"
	"log"
	"os"

	"erpwms/backend-go/internal/common/config"
	"erpwms/backend-go/internal/common/crypto"
	"erpwms/backend-go/internal/common/security"
	sqlc "erpwms/backend-go/internal/db/sqlcgen"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	db, err := pgxpool.New(context.Background(), cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	q := sqlc.New(db)
	ctx := context.Background()

	_, _ = db.Exec(ctx, "INSERT INTO permissions(name) VALUES ('wms.stock.read'),('wms.stock.move'),('admin.users.read') ON CONFLICT DO NOTHING")
	_, _ = db.Exec(ctx, "INSERT INTO roles(name) VALUES ('Admin') ON CONFLICT DO NOTHING")
	_, _ = db.Exec(ctx, "INSERT INTO role_permissions(role_id, permission_id) SELECT r.id,p.id FROM roles r, permissions p WHERE r.name='Admin' ON CONFLICT DO NOTHING")

	email := os.Getenv("ADMIN_EMAIL")
	pwd := os.Getenv("ADMIN_PASSWORD")
	if email == "" || pwd == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD required")
	}

	fe := crypto.FieldEncryption{CurrentID: cfg.FieldEncCurrentKeyID}
	if cfg.FieldEncCurrentB64 != "" {
		fe.CurrentKey, _ = base64.StdEncoding.DecodeString(cfg.FieldEncCurrentB64)
	} else {
		fe.CurrentKey = []byte("12345678901234567890123456789012")
	}
	enc, _ := fe.EncryptString(email, "users:new:email")
	hash, _ := crypto.HashPassword(pwd, crypto.DefaultArgon2Params())
	u, err := q.CreateUser(ctx, sqlc.CreateUserParams{EmailHash: security.EmailHash(email, cfg.SearchPepper), EmailEnc: enc.Ciphertext, EmailNonce: enc.Nonce, EmailKeyID: enc.KeyID, PasswordHash: hash, Status: "active"})
	if err != nil {
		log.Fatal(err)
	}
	_, _ = db.Exec(ctx, "INSERT INTO user_roles(user_id, role_id) SELECT $1, id FROM roles WHERE name='Admin' ON CONFLICT DO NOTHING", u.ID)
	log.Printf("seeded admin user %s", u.ID)
}
