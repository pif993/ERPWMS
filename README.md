# ERPWMS Monorepo

Production-grade ERP+WMS monorepo scaffold with:
- **Go backend** (Gin API + worker, modular monolith ready for extraction)
- **Go SSR portal** (templates + HTMX)
- **Python analytics** (FastAPI, read-only DB access)
- **Infra** (PostgreSQL, Redis, NATS, Caddy, Docker Compose, K8s/Helm scaffolding)

## Quick start
1. Copy `infra/.env.example` to `infra/.env` and set secure values.
2. Run `make dev`.
3. Run `make gen-sqlc migrate-up seed`.
4. Open `http://localhost:8080`.

## Security posture
- Deny-by-default permissions and restrictive CORS.
- Argon2id password hashing and envelope field encryption.
- JWT key rotation support (`CURRENT` + `PREVIOUS`).
- Append-only audit and stock ledgers with DB triggers.
- Structured JSON logging with PII redaction.

See `docs/` for architecture, runbooks, and security details.
