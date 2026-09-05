#!/bin/bash
# ReportSummary entrypoint invoked by Fluid's CacheRuntime. It samples the
# Mooncake master's metrics endpoint and prints the JSON that Fluid writes into
# the Dataset's status.cacheStates field.
set -euo pipefail

RAW=$(curl -s http://localhost:9003/metrics/summary)

if [[ -z "$RAW" ]]; then
  echo "Error: empty response from metrics endpoint" >&2
  exit 1
fi

# Extract the "Mem Storage: 0 B / 2.00 GB (0.0%)" segment.
MEM_LINE=$(echo "$RAW" | grep -oE 'Mem Storage: [^|]+' || true)

CACHED_RAW=$(echo "$MEM_LINE" | sed -E 's/Mem Storage: ([^/]+) \/.*/\1/' | xargs)
CAPACITY_RAW=$(echo "$MEM_LINE" | sed -E 's/.*\/ ([^(]+) \(.*/\1/' | xargs)
PERCENT_RAW=$(echo "$MEM_LINE" | grep -oE '\([0-9.]+%\)' | tr -d '()%')

# Normalize units ("2.00 GB" -> "2.00GiB").
normalize_unit() {
  echo "$1" | sed -E 's/ ?GB$/GiB/; s/ ?MB$/MiB/; s/ ?B$/B/' | tr -d ' '
}
CACHED=$(normalize_unit "$CACHED_RAW")
CACHE_CAPACITY=$(normalize_unit "$CAPACITY_RAW")

# Report the number of keys as fileNum.
FILE_NUM=$(echo "$RAW" | grep -oE 'Keys: [0-9]+' | grep -oE '[0-9]+' || echo "0")

# Approximate the hit ratio with the success rate of Get requests.
GET_STATS=$(echo "$RAW" | grep -oE 'Get=[0-9.]+/[0-9.]+' || echo "Get=0.00/0.00")
GET_SUCCESS=$(echo "$GET_STATS" | cut -d= -f2 | cut -d/ -f1)
GET_TOTAL=$(echo "$GET_STATS" | cut -d/ -f2)
HIT_RATIO=$(awk -v s="$GET_SUCCESS" -v t="$GET_TOTAL" \
  'BEGIN{ if (t>0) printf "%.0f", (s/t*100); else print "0" }')

# Mooncake has no underlying storage, so report the cache capacity as ufsTotal.
UFS_TOTAL="$CACHE_CAPACITY"

cat <<JSON
{
  "cached": "$CACHED",
  "cachedPercentage": "$PERCENT_RAW",
  "cacheCapacity": "$CACHE_CAPACITY",
  "cacheHitRatio": "$HIT_RATIO",
  "fileNum": "$FILE_NUM",
  "ufsTotal": "$UFS_TOTAL"
}
JSON
