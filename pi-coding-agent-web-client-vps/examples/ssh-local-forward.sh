#!/usr/bin/env bash
set -euo pipefail

VPS_HOST="${1:-user@your-vps.example}"
LOCAL_PORT="${LOCAL_PORT:-9000}"
REMOTE_PORT="${REMOTE_PORT:-9000}"

exec ssh -N \
  -o ServerAliveInterval=30 \
  -o ServerAliveCountMax=3 \
  -L "${LOCAL_PORT}:127.0.0.1:${REMOTE_PORT}" \
  "${VPS_HOST}"
