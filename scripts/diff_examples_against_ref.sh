#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
BASE_REF=${1:-origin/main}
CORPUS_FILE=${CORPUS_FILE:-"$ROOT_DIR/tests/corpus/examples_diff.txt"}
TIMEOUT_SECONDS=${KARL_DIFF_TIMEOUT:-20}
GOCACHE_DIR=${GOCACHE:-/tmp/karl-go-cache}

run_with_timeout() {
  perl -e 'alarm shift; exec @ARGV' "$@"
}

if [[ ! -f "$CORPUS_FILE" ]]; then
  echo "error: corpus not found: $CORPUS_FILE" >&2
  exit 1
fi

git -C "$ROOT_DIR" rev-parse --verify "$BASE_REF" >/dev/null

HEAD_SHA=$(git -C "$ROOT_DIR" rev-parse HEAD)
BASE_SHA=$(git -C "$ROOT_DIR" rev-parse "$BASE_REF")

HEAD_DIR=$(mktemp -d)
BASE_DIR=$(mktemp -d)
HEAD_BIN=$(mktemp)
BASE_BIN=$(mktemp)
HEAD_OUT=$(mktemp)
BASE_OUT=$(mktemp)

cleanup() {
  git -C "$ROOT_DIR" worktree remove --force "$HEAD_DIR" >/dev/null 2>&1 || true
  git -C "$ROOT_DIR" worktree remove --force "$BASE_DIR" >/dev/null 2>&1 || true
  rm -f "$HEAD_BIN" "$BASE_BIN" "$HEAD_OUT" "$BASE_OUT"
}
trap cleanup EXIT

git -C "$ROOT_DIR" worktree add --detach "$HEAD_DIR" "$HEAD_SHA" >/dev/null
git -C "$ROOT_DIR" worktree add --detach "$BASE_DIR" "$BASE_SHA" >/dev/null

(cd "$HEAD_DIR" && GOCACHE="$GOCACHE_DIR" go build -o "$HEAD_BIN" .)
(cd "$BASE_DIR" && GOCACHE="$GOCACHE_DIR" go build -o "$BASE_BIN" .)

total=0
mismatches=0

while IFS= read -r rel; do
  [[ -z "$rel" || "${rel:0:1}" == "#" ]] && continue
  total=$((total + 1))

  head_file="$HEAD_DIR/$rel"
  base_file="$BASE_DIR/$rel"

  if [[ ! -f "$head_file" || ! -f "$base_file" ]]; then
    echo "[MISMATCH][missing-file] $rel"
    mismatches=$((mismatches + 1))
    continue
  fi

  set +e
  run_with_timeout "$TIMEOUT_SECONDS" "$HEAD_BIN" run "$head_file" >"$HEAD_OUT" 2>&1
  head_code=$?
  run_with_timeout "$TIMEOUT_SECONDS" "$BASE_BIN" run "$base_file" >"$BASE_OUT" 2>&1
  base_code=$?
  set -e

  if [[ $head_code -ne $base_code ]]; then
    echo "[MISMATCH][exit-code] $rel head=$head_code base=$base_code"
    mismatches=$((mismatches + 1))
    continue
  fi

  if ! cmp -s "$HEAD_OUT" "$BASE_OUT"; then
    echo "[MISMATCH][output] $rel"
    diff -u "$BASE_OUT" "$HEAD_OUT" | head -80 | sed 's/^/  /' || true
    mismatches=$((mismatches + 1))
    continue
  fi

  echo "[OK] $rel"
done <"$CORPUS_FILE"

echo ""
echo "Diff summary: total=$total mismatches=$mismatches base=$BASE_REF head=$HEAD_SHA"

if [[ $mismatches -ne 0 ]]; then
  exit 1
fi
