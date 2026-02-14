#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${DB_URL:-}" ]]; then
  echo "DB_URL is required"
  exit 1
fi
if [[ -z "${ADMIN_EMAIL:-}" || -z "${ADMIN_PASSWORD:-}" ]]; then
  echo "ADMIN_EMAIL and ADMIN_PASSWORD are required"
  exit 1
fi

echo "[seeder] running seed..."
/app/seed
echo "[seeder] done."
