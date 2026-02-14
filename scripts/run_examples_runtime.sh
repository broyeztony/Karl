#!/usr/bin/env sh
set -eu

ROOT_DIR=$(git rev-parse --show-toplevel)
KARL_BIN=${KARL_BIN:-karl}
TIMEOUT_SECONDS=${KARL_EXAMPLE_TIMEOUT:-60}
INCLUDE_NETWORK=${KARL_INCLUDE_NETWORK_EXAMPLES:-0}

run_with_timeout() {
  perl -e 'alarm shift; exec @ARGV' "$@"
}

if ! command -v "$KARL_BIN" >/dev/null 2>&1; then
  if [ -x "$ROOT_DIR/karl" ]; then
    KARL_BIN="$ROOT_DIR/karl"
  else
    echo "error: karl binary not found (set KARL_BIN or build ./karl)" >&2
    exit 1
  fi
fi

total=0
passed=0
failed=0
skipped=0
tmp_log=$(mktemp)
tmp_list=$(mktemp)
trap 'rm -f "$tmp_log" "$tmp_list"' EXIT

find "$ROOT_DIR/examples" -type f -name '*.k' | sort > "$tmp_list"

while IFS= read -r file; do
  [ -z "$file" ] && continue
  total=$((total + 1))

  if [ "$INCLUDE_NETWORK" != "1" ] && grep -q 'http(' "$file"; then
    echo "[SKIP][network] ${file#$ROOT_DIR/}"
    skipped=$((skipped + 1))
    continue
  fi

  short_file=${file#$ROOT_DIR/}
  echo "[RUN] $short_file"

  set +e
  run_with_timeout "$TIMEOUT_SECONDS" "$KARL_BIN" run "$file" >"$tmp_log" 2>&1
  code=$?
  set -e

  if [ "$code" -ne 0 ]; then
    echo "[FAIL][$code] $short_file"
    tail -20 "$tmp_log" | sed 's/^/  /'
    failed=$((failed + 1))
    continue
  fi

  if grep -q "runtime error:" "$tmp_log"; then
    echo "[FAIL][runtime-error-output] $short_file"
    tail -20 "$tmp_log" | sed 's/^/  /'
    failed=$((failed + 1))
    continue
  fi

  passed=$((passed + 1))
done < "$tmp_list"

echo ""
echo "Examples runtime summary: total=$total passed=$passed skipped=$skipped failed=$failed"

if [ "$failed" -ne 0 ]; then
  exit 1
fi
