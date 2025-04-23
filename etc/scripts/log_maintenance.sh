#!/bin/bash

#log storage location
LOG_DIR="./logs"

#compress .log files older than 1 day (except those that have been compressed).
find "$LOG_DIR" -type f -name "*.log" -mtime +1 ! -name "*.gz" -exec gzip {} \;

#delete .log.gz older than 30 days.
find "$LOG_DIR" -type f -name "*.log.gz" -mtime +30 -exec rm -f {} \;