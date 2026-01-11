#!/usr/bin/env bash
# Monitor the length of the withdrawal queue and alert if it grows unexpectedly.

set -euo pipefail

QUEUE_API_URL="${QUEUE_API_URL:-http://localhost:8080/queue/length}"
THRESHOLD="${THRESHOLD:-50}"

length=$(curl -s "$QUEUE_API_URL")
timestamp=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

if [[ $length -ge $THRESHOLD ]]; then
  echo "$timestamp WARNING: Queue length is $length, exceeds threshold $THRESHOLD"
else
  echo "$timestamp OK: Queue length is $length"
fi
