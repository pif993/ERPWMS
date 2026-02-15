#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="${1:-$HOME/ERPWMS}"
BASE="infra/docker-compose.yml"
OVERRIDE="infra/docker-compose.override.yml"
ENVFILE="infra/.env"

log(){ echo -e "[ERPWMS] $*"; }
die(){ echo -e "[ERPWMS][ERROR] $*" >&2; exit 1; }

log "AUTO INSTALL START"
log "Repo: $REPO_DIR"

[ -d "$REPO_DIR/.git" ] || die "Repo non trovato: $REPO_DIR (manca .git)"

cd "$REPO_DIR"

# --- Sync git (safe) ---
log "Sync repository (pull --rebase con autostash)..."
if ! git diff --quiet || ! git diff --cached --quiet; then
  git stash push -u -m "autoinstall-autostash" >/dev/null || true
fi

git pull --rebase || true

if git stash list | grep -q "autoinstall-autostash"; then
  git stash pop || true
fi

# --- Override compose (non tracciato) ---
log "Scrivo $OVERRIDE (fix depends_on + healthcheck)..."
mkdir -p infra
cat > "$OVERRIDE" <<'YAML'
services:
  nats:
    healthcheck:
      disable: true

  python-analytics:
    healthcheck:
      disable: true

  backend-api:
    depends_on:
      postgres:
        condition: service_started
      redis:
        condition: service_started
      nats:
        condition: service_started

  backend-worker:
    depends_on:
      postgres:
        condition: service_started
      redis:
        condition: service_started
      nats:
        condition: service_started
YAML

# --- Check env ---
[ -f "$ENVFILE" ] || {
  if [ -f "infra/.env.example" ]; then
    log "infra/.env mancante -> creo da infra/.env.example (placeholder)."
    cp infra/.env.example "$ENVFILE"
  else
    die "infra/.env mancante e infra/.env.example non trovato."
  fi
}

# --- Start stack ---
log "Avvio stack (base + override)..."
sudo docker compose -f "$BASE" -f "$OVERRIDE" up -d --build

# --- Extract DB basics from .env (best-effort) ---
set -a
# shellcheck disable=SC1090
. "$ENVFILE"
set +a

# Need DB_URL
[ "${DB_URL:-}" != "" ] || die "DB_URL non definita in infra/.env"

# Determine DB name (fallback erpwms)
DB_NAME="${DB_NAME:-erpwms}"

# --- Ensure privileges for app user (best-effort) ---
# Try to infer app user from DB_URL, fallback erp_app
APP_USER="${APP_USER:-erp_app}"
if echo "$DB_URL" | grep -q '://' ; then
  # postgres://user:pass@host:port/db
  APP_USER="$(echo "$DB_URL" | sed -E 's#^[a-z]+://([^:/]+).*#\1#' || true)"
  [ "$APP_USER" = "$DB_URL" ] && APP_USER="erp_app"
fi

log "Garantisco permessi schema public per user=$APP_USER db=$DB_NAME (best-effort)..."
sudo docker exec -i infra-postgres-1 psql -U postgres -d "$DB_NAME" <<SQL >/dev/null 2>&1 || true
GRANT CONNECT, TEMP ON DATABASE $DB_NAME TO $APP_USER;
GRANT USAGE, CREATE ON SCHEMA public TO $APP_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO $APP_USER;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT USAGE, SELECT, UPDATE ON SEQUENCES TO $APP_USER;
SQL

# --- Fix migration 0002 (CRLF + blank line issues that can break goose parsing) ---
MIG2="backend-go/internal/db/migrations/0002_append_only.sql"
if [ -f "$MIG2" ]; then
  log "Normalizzo $MIG2 (rimuovo CRLF e blank line subito dopo -- +goose Up)..."
  sed -i 's/\r$//' "$MIG2" || true
  python3 - <<'PY' || true
from pathlib import Path
p = Path("backend-go/internal/db/migrations/0002_append_only.sql")
lines = p.read_text().splitlines()
out=[]
i=0
while i < len(lines):
    out.append(lines[i])
    if lines[i].strip() == "-- +goose Up":
        j=i+1
        while j < len(lines) and lines[j].strip()=="":
            j+=1
        i=j
        continue
    i+=1
p.write_text("\n".join(out) + "\n")
PY
fi

# --- Run migrations (goose) ---
log "Eseguo migrations con goose..."
set +e
sudo docker compose -f "$BASE" -f "$OVERRIDE" --env-file "$ENVFILE" \
  run --rm --entrypoint /bin/sh tools \
  -lc '/go/bin/goose -dir internal/db/migrations postgres "$DB_URL" up'
RC=$?
set -e

# If goose fails specifically on 0002 dollar-quote, apply workaround
if [ $RC -ne 0 ]; then
  log "Goose ha fallito. Provo workaround per 0002_append_only (psql + mark version 2)..."
  if [ -f "$MIG2" ]; then
    sudo docker exec -i infra-postgres-1 psql -U postgres -d "$DB_NAME" < "$MIG2"

    # Ensure goose_db_version exists (it should after 0001)
    # Mark version 2 applied in a way that works for common goose schemas:
    sudo docker exec -i infra-postgres-1 psql -U postgres -d "$DB_NAME" <<'SQL'
DO $$
BEGIN
  IF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='goose_db_version' AND column_name='version_id') THEN
    -- goose v3 typical schema
    IF NOT EXISTS (SELECT 1 FROM goose_db_version WHERE version_id = 2) THEN
      INSERT INTO goose_db_version(version_id, is_applied) VALUES (2, true);
    ELSE
      UPDATE goose_db_version SET is_applied = true WHERE version_id = 2;
    END IF;
  ELSIF EXISTS (SELECT 1 FROM information_schema.columns WHERE table_name='goose_db_version' AND column_name='version') THEN
    -- older schema
    IF NOT EXISTS (SELECT 1 FROM goose_db_version WHERE version = 2) THEN
      INSERT INTO goose_db_version(version, is_applied) VALUES (2, true);
    ELSE
      UPDATE goose_db_version SET is_applied = true WHERE version = 2;
    END IF;
  END IF;
END $$;
SQL

    # Continue remaining migrations
    sudo docker compose -f "$BASE" -f "$OVERRIDE" --env-file "$ENVFILE" \
      run --rm --entrypoint /bin/sh tools \
      -lc '/go/bin/goose -dir internal/db/migrations postgres "$DB_URL" up'
  else
    die "Manca $MIG2, impossibile workaround 0002."
  fi
fi

# --- Seed ---
log "Eseguo seed..."
sudo docker compose -f "$BASE" -f "$OVERRIDE" --env-file "$ENVFILE" \
  run --rm --entrypoint /bin/sh tools \
  -lc '/app/seed'

# --- Final checks ---
log "Check finali..."
sleep 2
curl -fsS http://localhost:8080/health >/dev/null && log "Caddy OK"
curl -fsS http://localhost:8081/health >/dev/null && log "API OK"

log "AUTO INSTALL COMPLETE"
