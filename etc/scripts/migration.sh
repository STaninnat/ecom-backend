#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
ENV_FILE="$REPO_ROOT/.env.development"
[ -f "$ENV_FILE" ] && . "$ENV_FILE"

: "${DATABASE_URL:?DATABASE_URL is not set - check .env.development}"

goose -dir "$REPO_ROOT/sql/schema" postgres "$DATABASE_URL" "$@"
