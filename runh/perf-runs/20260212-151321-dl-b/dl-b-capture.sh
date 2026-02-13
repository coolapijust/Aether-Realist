#!/usr/bin/env bash
set -euo pipefail
sleep 5
timeout 120 docker compose -f "/root/deploy/docker-compose.yml" logs -f aether-gateway-core 2>&1 | grep --line-buffered "\[PERF\]" > "/root/perf-runs/20260212-151321-dl-b/dl-b-perf.log" || true
