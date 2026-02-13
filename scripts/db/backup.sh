#!/usr/bin/env bash
set -euo pipefail
: "${POSTGRES_HOST:?}" "${POSTGRES_PORT:?}" "${POSTGRES_DB:?}" "${POSTGRES_SUPER_USER:?}" "${POSTGRES_SUPER_PASSWORD:?}"
mkdir -p backups
export PGPASSWORD="$POSTGRES_SUPER_PASSWORD"
pg_dump -h "$POSTGRES_HOST" -p "$POSTGRES_PORT" -U "$POSTGRES_SUPER_USER" "$POSTGRES_DB" > "backups/backup-$(date +%F-%H%M%S).sql"
