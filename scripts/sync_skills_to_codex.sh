#!/usr/bin/env sh
set -eu

QUIET=0
if [ "${1:-}" = "--quiet" ]; then
  QUIET=1
fi

ROOT="$(git rev-parse --show-toplevel)"
SRC="$ROOT/skills"
CODEX_HOME_DIR="${CODEX_HOME:-$HOME/.codex}"
DST="$CODEX_HOME_DIR/skills"

if [ ! -d "$SRC" ]; then
  [ "$QUIET" -eq 1 ] || echo "No skills directory at $SRC"
  exit 0
fi

mkdir -p "$DST"

synced=0
for skill_dir in "$SRC"/*; do
  [ -d "$skill_dir" ] || continue
  skill_name="$(basename "$skill_dir")"
  target_dir="$DST/$skill_name"
  mkdir -p "$target_dir"

  if command -v rsync >/dev/null 2>&1; then
    rsync -a "$skill_dir"/ "$target_dir"/
  else
    rm -rf "$target_dir"
    cp -R "$skill_dir" "$target_dir"
  fi
  synced=$((synced + 1))
done

if [ "$QUIET" -ne 1 ]; then
  echo "Synced $synced skill(s) from $SRC to $DST"
fi
