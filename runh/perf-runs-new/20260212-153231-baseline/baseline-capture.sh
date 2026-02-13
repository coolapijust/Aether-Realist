#!/usr/bin/env bash
set -euo pipefail
sleep 5
timeout 120 docker compose -f "/root/deploy/docker-compose.yml" logs --since "2026-02-12T07:33:00Z" -f aether-gateway-core 2>&1 | grep --line-buffered "\[PERF\]" > "/root/perf-runs/20260212-153231-baseline/baseline-perf.log" || true

{
  echo "preset=baseline"
  echo "since_utc=2026-02-12T07:33:00Z"
  echo "window_sec=120"
  if [[ -s "/root/perf-runs/20260212-153231-baseline/baseline-perf.log" ]]; then
    awk '
      match($0, /down\{mbps=([0-9.]+).*read_us=([0-9.]+)/, m) {
        n++
        mbps = m[1] + 0
        r = m[2] + 0
        sum += mbps
        if (mbps > 0.05) { nz++; nzsum += mbps }
        if (n == 1 || mbps < min) min = mbps
        if (mbps > max) max = mbps
        rsum += r
      }
      END {
        if (n == 0) {
          print "points=0"
        } else {
          printf "points=%d\n", n
          printf "down_avg_mbps=%.3f\n", sum / n
          printf "down_max_mbps=%.3f\n", max
          printf "down_min_mbps=%.3f\n", min
          printf "down_nonzero_points=%d\n", nz
          if (nz > 0) printf "down_nonzero_avg_mbps=%.3f\n", nzsum / nz
          printf "down_avg_read_us=%.1f\n", rsum / n
        }
      }
    ' "/root/perf-runs/20260212-153231-baseline/baseline-perf.log"
    echo "--- tail ---"
    tail -n 5 "/root/perf-runs/20260212-153231-baseline/baseline-perf.log" || true
  else
    echo "points=0"
    echo "note=perf file is empty; check runner log"
  fi
} > "/root/perf-runs/20260212-153231-baseline/baseline-summary.log" 2>&1
