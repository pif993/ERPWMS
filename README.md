# ERPWMS

Milestone 1 baseline funzionante (API Go + Worker + Analytics Python) con focus su sicurezza e affidabilità:
- Auth reale (users DB, Argon2id, refresh session hash-only)
- RBAC reale da DB (permissions/roles)
- Stock move transazionale con idempotency key + outbox
- Worker outbox robusto (`FOR UPDATE SKIP LOCKED`, retry/backoff)
- Portale SSR minimo (`/login`, `/stock`)

## Avvio rapido
1. Copia `infra/.env.example` in `infra/.env` e imposta valori sicuri.
2. Avvia stack: `docker compose -f infra/docker-compose.yml up -d --build`
3. Migrazioni: `make migrate-up`
4. Genera sqlc: `make gen-sqlc`
5. Seed admin: `make seed`

## Endpoint principali
- `GET /health`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `GET /api/stock/balances`
- `POST /api/stock/moves` (header `Idempotency-Key` obbligatorio)

## Sicurezza
- CORS allowlist (niente wildcard in prod)
- secure headers + request_id
- redaction header sensibili
- append-only per `audit_log` e `stock_ledger`
- outbox con colonne immutabili eccetto stato invio/retry
