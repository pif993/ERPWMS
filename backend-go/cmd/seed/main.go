package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"os"

	"erpwms/backend-go/internal/common/config"
	"erpwms/backend-go/internal/common/crypto"
	"erpwms/backend-go/internal/common/security"
	sqlc "erpwms/backend-go/internal/db/sqlcgen"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()

	db, err := pgxpool.New(ctx, cfg.DBURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	q := sqlc.New(db)

	// Baseline RBAC seed (idempotente)
	_, _ = db.Exec(ctx, `
		INSERT INTO permissions(name)
		VALUES ('wms.stock.read'),('wms.stock.move'),('admin.users.read')
		ON CONFLICT DO NOTHING
	`)
	_, _ = db.Exec(ctx, `INSERT INTO roles(name) VALUES ('Admin') ON CONFLICT DO NOTHING`)
	_, _ = db.Exec(ctx, `
		INSERT INTO role_permissions(role_id, permission_id)
		SELECT r.id,p.id FROM roles r, permissions p
		WHERE r.name='Admin'
		ON CONFLICT DO NOTHING
	`)

	email := os.Getenv("ADMIN_EMAIL")
	pwd := os.Getenv("ADMIN_PASSWORD")
	if email == "" || pwd == "" {
		log.Fatal("ADMIN_EMAIL and ADMIN_PASSWORD required")
	}

	// Field encryption (prod: keys obbligatorie; dev: fallback)
	fe := crypto.FieldEncryption{
		CurrentID:  cfg.FieldEncCurrentKeyID,
		PreviousID: cfg.FieldEncPrevKeyID,
	}
	if cfg.FieldEncCurrentB64 != "" {
		fe.CurrentKey, _ = base64.StdEncoding.DecodeString(cfg.FieldEncCurrentB64)
	} else {
		// DEV-only fallback. In prod config.Load() dovrebbe imporre chiavi valide.
		fe.CurrentKey = []byte("12345678901234567890123456789012")
	}
	if cfg.FieldEncPreviousB64 != "" {
		fe.PreviousKey, _ = base64.StdEncoding.DecodeString(cfg.FieldEncPreviousB64)
	}

	enc, err := fe.EncryptString(email, "users:new:email")
	if err != nil {
		log.Fatal(err)
	}

	passHash, err := crypto.HashPassword(pwd, crypto.DefaultArgon2Params())
	if err != nil {
		log.Fatal(err)
	}

	emailHash := security.EmailHash(email, cfg.SearchPepper)

	// 1) Prova a creare
	u, err := q.CreateUser(ctx, sqlc.CreateUserParams{
		EmailHash:     emailHash,
		EmailEnc:      enc.Ciphertext,
		EmailNonce:    enc.Nonce,
		EmailKeyID:    enc.KeyID,
		PasswordHash:  passHash,
		Status:        "active",
	})

	// 2) Se esiste già → aggiorna credenziali + rileggi
	if err != nil {
		if isUniqueViolationUsersEmailHash(err) {
			_, uerr := db.Exec(ctx, `
				UPDATE users
				SET email_enc=$2, email_nonce=$3, email_key_id=$4,
				    password_hash=$5, status='active', updated_at=now()
				WHERE email_hash=$1
			`, emailHash, enc.Ciphertext, enc.Nonce, enc.KeyID, passHash)
			if uerr != nil {
				log.Fatal(uerr)
			}

			u, err = q.GetUserByEmailHash(ctx, emailHash)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Printf("[seed] admin already existed, updated credentials. user_id=%s\n", u.ID.String())
		} else {
			log.Fatal(err)
		}
	} else {
		fmt.Printf("[seed] created admin user. user_id=%s\n", u.ID.String())
	}

	// 3) Assicura ruolo Admin (idempotente)
	_, _ = db.Exec(ctx, `
		INSERT INTO user_roles(user_id, role_id)
		SELECT $1, id FROM roles WHERE name='Admin'
		ON CONFLICT DO NOTHING
	`, u.ID)

	fmt.Println("[seed] done.")
}

func isUniqueViolationUsersEmailHash(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505" && pgErr.ConstraintName == "users_email_hash_key"
	}
	return false
}
