#!/bin/bash

ENV_FILE="/home/humblestuff/workspace/github.com/STaninnat/ecom-backend/.env.development"
[ -f "$ENV_FILE" ] && . "$ENV_FILE"

: "${DATABASE_URL:?DATABASE_URL is not set - check .env}"

goose postgres "$DATABASE_URL" "$@"

# first cd to schema file location -> cd sql/schema/ and run
# example...
# ../../etc/scripts/migration.sh up
# ../../etc/scripts/migration.sh up-to 00001
# ../../etc/scripts/migration.sh status