#!/usr/bin/env sh
set -eu

ROOT="$(git rev-parse --show-toplevel)"
HOOKS_DIR="$(git rev-parse --git-path hooks)"
SYNC_SCRIPT="$ROOT/scripts/sync_skills_to_codex.sh"

if [ ! -x "$SYNC_SCRIPT" ]; then
  chmod +x "$SYNC_SCRIPT"
fi

install_hook() {
  hook_name="$1"
  hook_file="$HOOKS_DIR/$hook_name"
  cat >"$hook_file" <<'EOF'
#!/usr/bin/env sh
set -eu
ROOT="$(git rev-parse --show-toplevel)"
"$ROOT/scripts/sync_skills_to_codex.sh" --quiet || true
EOF
  chmod +x "$hook_file"
}

install_hook post-commit
install_hook post-merge
install_hook post-checkout

echo "Installed skill sync hooks in $HOOKS_DIR (post-commit, post-merge, post-checkout)"
