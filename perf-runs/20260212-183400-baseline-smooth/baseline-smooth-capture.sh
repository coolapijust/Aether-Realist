#!/usr/bin/env bash
set -euo pipefail
sleep 5
timeout 60 docker compose -f "/root/deploy/docker-compose.yml" logs --since "2026-02-12T10:34:05Z" -f aether-gateway-core 2>&1 | grep -E --line-buffered "\[PERF(-GW)?\]" > "/root/perf-runs/20260212-183400-baseline-smooth/baseline-smooth-perf.log" || true

{
  echo "preset=baseline-smooth"
  echo "since_utc=2026-02-12T10:34:05Z"
  echo "window_sec=60"
  if [[ -s "/root/perf-runs/20260212-183400-baseline-smooth/baseline-smooth-perf.log" ]]; then
    awk '
      match($0, /down\{mbps=([0-9.]+).*read_us=([0-9.]+)/, m) {
        n_core++
        mbps = m[1] + 0
        r = m[2] + 0
        core_sum += mbps
        if (n_core == 1 || mbps < core_min) core_min = mbps
        if (mbps > core_max) core_max = mbps
        core_rsum += r
      }
      match($0, /dl\{mbps=([0-9.]+).*write_us=([0-9.]+)/, g) {
        n_gw++
        mbps = g[1] + 0
        w = g[2] + 0
        gw_sum += mbps
        if (mbps > 0.10) {
          gw_nz++
          gw_nzsum += mbps
          gw_streak++
          if (gw_streak > gw_max_streak) gw_max_streak = gw_streak
        } else {
          gw_streak = 0
        }
        if (n_gw == 1 || mbps < gw_min) gw_min = mbps
        if (mbps > gw_max) gw_max = mbps
        gw_wsum += w
      }
      END {
        printf "points_total=%d\n", n_core + n_gw
        if (n_core > 0) {
          printf "core_down_points=%d\n", n_core
          printf "core_down_avg_mbps=%.3f\n", core_sum / n_core
          printf "core_down_max_mbps=%.3f\n", core_max
          printf "core_down_min_mbps=%.3f\n", core_min
          printf "core_down_avg_read_us=%.1f\n", core_rsum / n_core
        }
        if (n_gw > 0) {
          printf "gw_dl_points=%d\n", n_gw
          printf "gw_dl_avg_mbps=%.3f\n", gw_sum / n_gw
          printf "gw_dl_max_mbps=%.3f\n", gw_max
          printf "gw_dl_min_mbps=%.3f\n", gw_min
          printf "gw_dl_nonzero_points=%d\n", gw_nz
          if (gw_nz > 0) printf "gw_dl_nonzero_avg_mbps=%.3f\n", gw_nzsum / gw_nz
          printf "gw_dl_avg_write_us=%.1f\n", gw_wsum / n_gw
          printf "gw_dl_max_nonzero_streak=%d\n", gw_max_streak
          if (gw_max_streak >= 6) print "gw_dl_is_continuous6=yes"; else print "gw_dl_is_continuous6=no"
        }
        if (n_core == 0 && n_gw == 0) {
          print "points=0"
        }
      }
    ' "/root/perf-runs/20260212-183400-baseline-smooth/baseline-smooth-perf.log"
    echo "--- tail ---"
    tail -n 5 "/root/perf-runs/20260212-183400-baseline-smooth/baseline-smooth-perf.log" || true
  else
    echo "points=0"
    echo "note=perf file is empty; check runner log"
  fi
} > "/root/perf-runs/20260212-183400-baseline-smooth/baseline-smooth-summary.log" 2>&1
