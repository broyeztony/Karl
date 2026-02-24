#!/usr/bin/env sh
set -eu

ROOT_DIR=$(git rev-parse --show-toplevel)
KARL_BIN=${KARL_BIN:-karl}

if ! command -v "$KARL_BIN" >/dev/null 2>&1; then
  if [ -x "$ROOT_DIR/karl" ]; then
    KARL_BIN="$ROOT_DIR/karl"
  else
    echo "error: karl binary not found (set KARL_BIN or build ./karl)" >&2
    exit 1
  fi
fi

tmp_dir=$(mktemp -d)
trap 'rm -rf "$tmp_dir"' EXIT

program_file="$tmp_dir/debuggee.k"
input_file="$tmp_dir/input.txt"
output_file="$tmp_dir/output.log"

cat > "$program_file" <<'EOF'
let add = (a, b) -> a + b
let x = 1
let y = add(x, 2)
y
EOF

cat > "$input_file" <<'EOF'
break 3
continue
stack
locals
print x
next
continue
EOF

if ! "$KARL_BIN" trace "$program_file" < "$input_file" > "$output_file" 2>&1; then
  echo "FAIL: debugger command exited non-zero"
  cat "$output_file"
  exit 1
fi

if ! grep -q "Karl Debugger" "$output_file"; then
  echo "FAIL: missing debugger banner"
  cat "$output_file"
  exit 1
fi

if ! grep -Eq "paused at .*:1:" "$output_file"; then
  echo "FAIL: expected initial pause on line 1"
  cat "$output_file"
  exit 1
fi

if ! grep -Eq "breakpoint #[0-9]+ set at line 3" "$output_file"; then
  echo "FAIL: expected breakpoint setup output"
  cat "$output_file"
  exit 1
fi

if ! grep -Eq "paused at .*:3:" "$output_file"; then
  echo "FAIL: expected pause at breakpoint line 3"
  cat "$output_file"
  exit 1
fi

if ! grep -q "x = 1" "$output_file"; then
  echo "FAIL: expected locals output for x"
  cat "$output_file"
  exit 1
fi

if ! grep -Eq "paused at .*:4:" "$output_file"; then
  echo "FAIL: expected step-over pause at line 4"
  cat "$output_file"
  exit 1
fi

if ! grep -q "result: 3" "$output_file"; then
  echo "FAIL: expected final result output"
  cat "$output_file"
  exit 1
fi

if grep -q "runtime error:" "$output_file"; then
  echo "FAIL: unexpected runtime error output"
  cat "$output_file"
  exit 1
fi

module_file_2="$tmp_dir/lib_import_async.k"
program_file_2="$tmp_dir/main_import_async.k"
input_file_2="$tmp_dir/input_import_async.txt"
output_file_2="$tmp_dir/output_import_async.log"

cat > "$module_file_2" <<'EOF'
let addOne = n -> {
  let result = n + 1
  result
}
let twice = n -> addOne(addOne(n))
let asyncTwice = n -> & (() -> {
  sleep(80)
  twice(n)
})()
{ asyncTwice: asyncTwice }
EOF

cat > "$program_file_2" <<'EOF'
let makeLib = import "./lib_import_async.k"
let lib = makeLib()
let task = lib.asyncTwice(40)
let out = wait task
out
EOF

cat > "$input_file_2" <<EOF
break 4
break $module_file_2:5
break 5
continue
continue
continue
continue
EOF

if ! "$KARL_BIN" trace "$program_file_2" < "$input_file_2" > "$output_file_2" 2>&1; then
  echo "FAIL: debugger imported+async scenario exited non-zero"
  cat "$output_file_2"
  exit 1
fi

if ! grep -Eq "paused at .*main_import_async\\.k:4:" "$output_file_2"; then
  echo "FAIL: expected stop at main wait line"
  cat "$output_file_2"
  exit 1
fi

if ! grep -Eq "paused at .*lib_import_async\\.k:5:" "$output_file_2"; then
  echo "FAIL: expected stop in imported async worker line"
  cat "$output_file_2"
  exit 1
fi

if ! grep -Eq "paused at .*main_import_async\\.k:5:" "$output_file_2"; then
  echo "FAIL: expected stop at main post-wait line"
  cat "$output_file_2"
  exit 1
fi

if ! grep -q "result: 42" "$output_file_2"; then
  echo "FAIL: expected imported+async final result"
  cat "$output_file_2"
  exit 1
fi

if grep -q "runtime error:" "$output_file_2"; then
  echo "FAIL: unexpected runtime error in imported+async scenario"
  cat "$output_file_2"
  exit 1
fi

echo "PASS: debugger CLI e2e"
