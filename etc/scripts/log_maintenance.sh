#!/usr/bin/env bash

# -e: Exit immediately if any command returns a non-zero exit status.
#     This will stop the script as soon as a command fails, ensuring that errors don't go unnoticed.
# -u: Treat unset variables as an error and exit immediately.
#     This will stop the script if any variable is used without being initialized.
# -x: Print each command and its arguments as they are executed.
#     Useful for debugging during development, but should be disabled in production for cleaner logs and security.
set -eux

ENV_FILE="/home/humblestuff/workspace/github.com/STaninnat/ecom-backend/.env.development"
[ -f "$ENV_FILE" ] && . "$ENV_FILE"

: "${LOG_DIR:?LOG_DIR is not set - check .env}"

#compress .log files older than 1 day (except those that have been compressed).
find "$LOG_DIR" -type f -name "*.log" -mtime +1 ! -name "*.gz" -exec gzip {} \;

#delete .log.gz older than 30 days.
find "$LOG_DIR" -type f -name "*.log.gz" -mtime +30 -exec rm -f {} \;