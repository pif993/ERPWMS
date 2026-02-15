#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${DB_URL:-}" ]]; then
  echo "DB_URL is required"
  exit 1
fi

echo "[migrator] running goose up..."
/go/bin/goose -dir internal/db/migrations postgres "$DB_URL" up
echo "[migrator] done."
