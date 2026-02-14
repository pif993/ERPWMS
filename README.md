# ERPWMS

Monorepo **ERP + WMS** con baseline “Milestone 1” funzionante:

- **Backend Go** (Gin API + worker outbox, modular monolith)
- **Portale SSR Go** (templates: `/login`, `/stock`)
- **Analytics Python** (FastAPI, endpoint protetti da token di servizio)
- **Infra** (PostgreSQL, Redis, NATS, Caddy, Docker Compose)

## Architettura (overview)

**backend-go**
- API: `backend-go/cmd/api`
- Worker: `backend-go/cmd/worker`
- DB schema/migrations: `backend-go/internal/db/migrations`
- SQL access: `sqlc` (`backend-go/internal/db/sqlcgen`)
- Security: password **Argon2id**, session refresh **hash-only**, RBAC da DB, idempotency keys, outbox pattern

**python-analytics**
- FastAPI: `python-analytics/app/main.py`
- Autorizzazione tramite header `X-Service-Token`

**infra**
- `docker-compose.yml`: Postgres, Redis, NATS, backend-api/worker, python-analytics, Caddy
- Caddy reverse proxy:
  - `/analytics/*` -> python-analytics
  - tutto il resto -> backend-api

## Prerequisiti

- Docker + Docker Compose plugin
- (dev) Go installato solo se vuoi eseguire tool localmente (sqlc/goose), altrimenti usa i container tools

## Quick start (locale)

1) Crea env:

```bash
cp infra/.env.example infra/.env
```

2) Avvia stack:

```bash
make dev
```

3) Migrazioni + seed (opzione A: tools container consigliata):

```bash
docker compose -f infra/docker-compose.yml run --rm migrator
docker compose -f infra/docker-compose.yml run --rm seeder
```

Oppure (opzione B: tool Go sul tuo host):

```bash
make migrate-up
make seed
```

4) Apri:

- Portale (Caddy): `http://localhost:8080`
- API diretta: `http://localhost:8081/health`

## Autotest (GUI Web)

L’API espone una pagina di autotest (abilitabile via env):

- `GET /autotest` (GUI)
- `POST /api/autotest/run` (esecuzione suite)

Config in `infra/.env`:

```env
AUTOTEST_ENABLED=true
AUTOTEST_TOKEN=... # token forte
```

Apri `http://localhost:8081/autotest`, inserisci il token e premi **Run**.

> Nota: l’autotest esegue una suite E2E “in-process” (routing/middleware/handlers) e verifica health/login/balances.

## Autoinstall Linux server (completamente automatico)

Script:

```bash
sudo bash scripts/autoinstall_linux.sh
```

Variabili opzionali:

```bash
sudo REPO_URL=https://github.com/pif993/ERPWMS.git \
  INSTALL_DIR=/opt/erpwms \
  ADMIN_EMAIL=admin@example.com \
  ADMIN_PASSWORD='StrongPassw0rd!' \
  bash scripts/autoinstall_linux.sh
```

Lo script:
- installa Docker + compose plugin
- clona il repo
- genera `infra/.env` con segreti forti
- avvia compose
- esegue `migrator` e `seeder`

## Endpoint principali

- `GET /health`
- `POST /api/auth/login`
- `POST /api/auth/refresh`
- `POST /api/auth/logout`
- `GET /api/stock/balances`
- `POST /api/stock/moves` (header `Idempotency-Key` obbligatorio)

## Sicurezza (baseline)

- CORS allowlist (niente wildcard in prod)
- Security headers + request id
- Password hashing: **Argon2id**
- Field encryption: **AES-256-GCM** con key id (rotazione supportata)
- Refresh session: token non salvato in chiaro (hash-only)
- RBAC reale da DB (roles/permissions)
- Append-only per `audit_log` e `stock_ledger` (trigger DB)
- Outbox immutabile salvo stato invio/retry

## Sviluppo (comandi utili)

```bash
make dev          # up -d --build
make down         # stop
make logs         # logs -f
make gen-sqlc     # genera sqlc (richiede Go sul host)
make migrate-up   # goose up (host)
make seed         # seed admin (host)
make test-go      # go test ./...
make fmt          # gofmt + ruff format
```

## Troubleshooting

**1) Migrazioni falliscono**
- verifica `DB_URL` in `infra/.env`
- verifica che Postgres sia healthy: `docker compose -f infra/docker-compose.yml ps`

**2) API non in ascolto**
- controlla logs: `make logs`
- controlla health: `curl -sS http://localhost:8081/health`

**3) Autotest 403**
- controlla `AUTOTEST_ENABLED=true`
- usa header `X-Autotest-Token` corretto (lo stesso di `AUTOTEST_TOKEN`)

## License
MIT — vedi `LICENSE`.
