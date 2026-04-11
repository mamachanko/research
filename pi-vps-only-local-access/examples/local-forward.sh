#!/usr/bin/env bash
set -euo pipefail

VPS_HOST="${1:-user@your-vps.example}"

exec ssh -N \
  -o ServerAliveInterval=30 \
  -o ServerAliveCountMax=3 \
  -L 9000:127.0.0.1:9000 \
  "$VPS_HOST"
