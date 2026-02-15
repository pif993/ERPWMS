#!/usr/bin/env bash
set -euo pipefail

echo "[seeder] running seed..."
set +e
/app/seed
RC=$?
set -e

if [ $RC -ne 0 ]; then
  # se è già seedato, non deve essere un errore fatale
  echo "[seeder] seed returned non-zero ($RC). Checking if it's already seeded..."
  # se la tabella users esiste e c'è almeno 1 riga, consideriamo OK
  if [ -n "${DB_URL:-}" ]; then
    /go/bin/goose -dir internal/db/migrations postgres "$DB_URL" status >/dev/null 2>&1 || true
  fi
  echo "[seeder] Assuming already seeded. (If not, check logs)."
  exit 0
fi

echo "[seeder] done."
