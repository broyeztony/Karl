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

fail() {
  msg="$1"
  output_file="${2:-}"
  echo "FAIL: $msg"
  if [ -n "$output_file" ] && [ -f "$output_file" ]; then
    cat "$output_file"
  fi
  exit 1
}

assert_grep() {
  pattern="$1"
  file="$2"
  msg="$3"
  if ! grep -Eq "$pattern" "$file"; then
    fail "$msg" "$file"
  fi
}

assert_not_grep() {
  pattern="$1"
  file="$2"
  msg="$3"
  if grep -Eq "$pattern" "$file"; then
    fail "$msg" "$file"
  fi
}

run_trace() {
  program_file="$1"
  input_file="$2"
  output_file="$3"
  if ! "$KARL_BIN" trace "$program_file" < "$input_file" > "$output_file" 2>&1; then
    fail "debugger command exited non-zero ($program_file)" "$output_file"
  fi
}

# Scenario 1: basic breakpoint / print / next flow.
program_1="$tmp_dir/scenario1_basic.k"
input_1="$tmp_dir/scenario1_basic.in"
output_1="$tmp_dir/scenario1_basic.out"
cat > "$program_1" <<'EOF'
let add = (a, b) -> a + b
let x = 1
let y = add(x, 2)
y
EOF
cat > "$input_1" <<'EOF'
break 3
continue
stack
locals
print x
next
continue
EOF
run_trace "$program_1" "$input_1" "$output_1"
assert_grep "Karl Debugger" "$output_1" "missing debugger banner in scenario 1"
assert_grep "paused at .*:1:" "$output_1" "expected initial pause in scenario 1"
assert_grep "breakpoint #[0-9]+ set at line 3" "$output_1" "expected breakpoint setup output in scenario 1"
assert_grep "paused at .*:3:" "$output_1" "expected pause at line 3 in scenario 1"
assert_grep "x = 1" "$output_1" "expected locals output for x in scenario 1"
assert_grep "paused at .*:4:" "$output_1" "expected next-stop at line 4 in scenario 1"
assert_grep "result: 3" "$output_1" "expected final result in scenario 1"
assert_not_grep "runtime error:" "$output_1" "unexpected runtime error in scenario 1"

# Scenario 2: imported + async + wait stop order.
module_2="$tmp_dir/lib_import_async.k"
program_2="$tmp_dir/main_import_async.k"
input_2="$tmp_dir/main_import_async.in"
output_2="$tmp_dir/main_import_async.out"
cat > "$module_2" <<'EOF'
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
cat > "$program_2" <<'EOF'
let makeLib = import "./lib_import_async.k"
let lib = makeLib()
let task = lib.asyncTwice(40)
let out = wait task
out
EOF
cat > "$input_2" <<EOF
break 4
break $module_2:5
break 5
continue
continue
continue
continue
EOF
run_trace "$program_2" "$input_2" "$output_2"
assert_grep "paused at .*main_import_async\\.k:4:" "$output_2" "expected stop at main wait line in scenario 2"
assert_grep "paused at .*lib_import_async\\.k:5:" "$output_2" "expected stop in imported async worker in scenario 2"
assert_grep "paused at .*main_import_async\\.k:5:" "$output_2" "expected stop at main post-wait line in scenario 2"
assert_grep "result: 42" "$output_2" "expected final result in scenario 2"
assert_not_grep "runtime error:" "$output_2" "unexpected runtime error in scenario 2"

# Scenario 3: step into imported function maps to imported source path.
module_3="$tmp_dir/lib_step_import.k"
program_3="$tmp_dir/main_step_import.k"
input_3="$tmp_dir/main_step_import.in"
output_3="$tmp_dir/main_step_import.out"
cat > "$module_3" <<'EOF'
let inc = n -> {
  let y = n + 1
  y
}
let use = n -> inc(n)
{ use: use }
EOF
cat > "$program_3" <<'EOF'
let makeLib = import "./lib_step_import.k"
let lib = makeLib()
let out = lib.use(41)
out
EOF
cat > "$input_3" <<'EOF'
break 3
continue
step
stack
continue
EOF
run_trace "$program_3" "$input_3" "$output_3"
assert_grep "paused at .*main_step_import\\.k:3:" "$output_3" "expected stop at call site in scenario 3"
assert_grep "paused at .*lib_step_import\\.k:[0-9]+:" "$output_3" "expected step into imported file in scenario 3"
assert_grep "\\*#0 use at .*lib_step_import\\.k" "$output_3" "expected top stack frame in imported file in scenario 3"
assert_grep "result: 42" "$output_3" "expected final result in scenario 3"
assert_not_grep "runtime error:" "$output_3" "unexpected runtime error in scenario 3"

# Scenario 4: finish steps out to caller.
program_4="$tmp_dir/scenario4_finish.k"
input_4="$tmp_dir/scenario4_finish.in"
output_4="$tmp_dir/scenario4_finish.out"
cat > "$program_4" <<'EOF'
let inc = n -> {
  let x = n + 1
  x
}
let wrap = n -> inc(n)
let out = wrap(41)
out
EOF
cat > "$input_4" <<'EOF'
break 6
continue
step
stack
finish
stack
continue
EOF
run_trace "$program_4" "$input_4" "$output_4"
assert_grep "paused at .*scenario4_finish\\.k:6:" "$output_4" "expected stop at call line in scenario 4"
assert_grep "paused at .*scenario4_finish\\.k:5:" "$output_4" "expected step into wrap in scenario 4"
assert_grep "\\*#0 wrap at .*scenario4_finish\\.k:5:" "$output_4" "expected wrap on stack in scenario 4"
assert_grep "paused at .*scenario4_finish\\.k:7:" "$output_4" "expected finish to stop at caller in scenario 4"
assert_grep "\\*#0 <top-level> at .*scenario4_finish\\.k:7:" "$output_4" "expected top-level stack after finish in scenario 4"
assert_grep "result: 42" "$output_4" "expected final result in scenario 4"
assert_not_grep "runtime error:" "$output_4" "unexpected runtime error in scenario 4"

# Scenario 5: breakpoint lifecycle commands (list/delete/clear).
program_5="$tmp_dir/scenario5_breakpoint_lifecycle.k"
input_5="$tmp_dir/scenario5_breakpoint_lifecycle.in"
output_5="$tmp_dir/scenario5_breakpoint_lifecycle.out"
cat > "$program_5" <<'EOF'
let x = 1
let y = x + 1
let z = y + 1
z
EOF
cat > "$input_5" <<'EOF'
break 2
breakpoints
delete 1
breakpoints
break 3
clear
breakpoints
continue
EOF
run_trace "$program_5" "$input_5" "$output_5"
assert_grep "breakpoint #1 set at line 2" "$output_5" "expected first breakpoint set in scenario 5"
assert_grep "#1 .*scenario5_breakpoint_lifecycle\\.k:2" "$output_5" "expected listed breakpoint in scenario 5"
assert_grep "removed breakpoint #1" "$output_5" "expected delete output in scenario 5"
assert_grep "breakpoint #2 set at line 3" "$output_5" "expected second breakpoint set in scenario 5"
assert_grep "cleared 1 breakpoint\\(s\\)" "$output_5" "expected clear output in scenario 5"
assert_grep "\\(no breakpoints\\)" "$output_5" "expected empty breakpoints output in scenario 5"
assert_grep "result: 3" "$output_5" "expected final result in scenario 5"
assert_not_grep "runtime error:" "$output_5" "unexpected runtime error in scenario 5"

# Scenario 6: watch lifecycle and watch value rendering.
program_6="$tmp_dir/scenario6_watch_flow.k"
input_6="$tmp_dir/scenario6_watch_flow.in"
output_6="$tmp_dir/scenario6_watch_flow.out"
cat > "$program_6" <<'EOF'
let add = (a, b) -> {
  let sum = a + b
  sum
}
let x = 10
let y = add(x, 2)
y
EOF
cat > "$input_6" <<'EOF'
break 2
continue
locals
watch a
watch a + b
watches
next
unwatch 1
clearwatches
continue
EOF
run_trace "$program_6" "$input_6" "$output_6"
assert_grep "paused at .*scenario6_watch_flow\\.k:2:" "$output_6" "expected function-body pause in scenario 6"
assert_grep "a = 10" "$output_6" "expected local a in scenario 6"
assert_grep "b = 2" "$output_6" "expected local b in scenario 6"
assert_grep "watch #1: a" "$output_6" "expected first watch add in scenario 6"
assert_grep "watch #2: a \\+ b" "$output_6" "expected second watch add in scenario 6"
assert_grep "#1 a" "$output_6" "expected watch listing #1 in scenario 6"
assert_grep "#2 a \\+ b" "$output_6" "expected watch listing #2 in scenario 6"
assert_grep "watch:" "$output_6" "expected watch values block in scenario 6"
assert_grep "#2 a \\+ b = 12" "$output_6" "expected evaluated watch value in scenario 6"
assert_grep "removed watch #1" "$output_6" "expected unwatch output in scenario 6"
assert_grep "cleared 1 watch\\(es\\)" "$output_6" "expected clearwatches output in scenario 6"
assert_grep "result: 12" "$output_6" "expected final result in scenario 6"
assert_not_grep "runtime error:" "$output_6" "unexpected runtime error in scenario 6"

# Scenario 7: frame switching + print from selected frame.
program_7="$tmp_dir/scenario7_frame_selection.k"
input_7="$tmp_dir/scenario7_frame_selection.in"
output_7="$tmp_dir/scenario7_frame_selection.out"
cat > "$program_7" <<'EOF'
let inc = n -> {
  let x = n + 1
  x
}
let wrap = m -> inc(m)
let out = wrap(41)
out
EOF
cat > "$input_7" <<'EOF'
break 2
continue
stack
frame 1
print m
print n
frame 0
print n
print m
continue
EOF
run_trace "$program_7" "$input_7" "$output_7"
assert_grep "\\*#0 inc at .*scenario7_frame_selection\\.k:1:" "$output_7" "expected top frame inc in scenario 7"
assert_grep "#1 wrap at .*scenario7_frame_selection\\.k:5:" "$output_7" "expected second frame wrap in scenario 7"
assert_grep "selected frame #1" "$output_7" "expected frame #1 selection in scenario 7"
assert_grep "selected frame #0" "$output_7" "expected frame #0 selection in scenario 7"
assert_grep "undefined identifier: n" "$output_7" "expected frame-scoped lookup failure for n in scenario 7"
assert_grep "undefined identifier: m" "$output_7" "expected frame-scoped lookup failure for m in scenario 7"
assert_grep "result: 42" "$output_7" "expected final result in scenario 7"
assert_not_grep "runtime error:" "$output_7" "unexpected runtime error in scenario 7"

# Scenario 8: breakpoint inside spawned task.
program_8="$tmp_dir/scenario8_spawned_task.k"
input_8="$tmp_dir/scenario8_spawned_task.in"
output_8="$tmp_dir/scenario8_spawned_task.out"
cat > "$program_8" <<'EOF'
let c = channel()
let worker = (ch) -> {
  ch.send("ok")
  ch.done()
}
let t = & worker(c)
let pair = c.recv()
wait t
pair
EOF
cat > "$input_8" <<EOF
break $program_8:3
continue
stack
locals
continue
EOF
run_trace "$program_8" "$input_8" "$output_8"
assert_grep "breakpoint #[0-9]+ set at .*scenario8_spawned_task\\.k:3" "$output_8" "expected spawned-task breakpoint set in scenario 8"
assert_grep "paused at .*scenario8_spawned_task\\.k:3:" "$output_8" "expected spawned-task breakpoint hit in scenario 8"
assert_grep "\\*#0 worker at .*scenario8_spawned_task\\.k:2:" "$output_8" "expected worker frame on stack in scenario 8"
assert_grep "ch = <channel>" "$output_8" "expected worker locals in scenario 8"
assert_grep "result: \\[\"ok\", false\\]" "$output_8" "expected final channel pair in scenario 8"
assert_not_grep "runtime error:" "$output_8" "unexpected runtime error in scenario 8"

# Scenario 9: command validation / error messaging.
program_9="$tmp_dir/scenario9_invalid_commands.k"
input_9="$tmp_dir/scenario9_invalid_commands.in"
output_9="$tmp_dir/scenario9_invalid_commands.out"
cat > "$program_9" <<'EOF'
1
EOF
cat > "$input_9" <<'EOF'
break
break abc
break :3
delete x
frame abc
frame 99
locals nope
print
watch
unwatch foo
clearwatches
unknowncmd
continue
EOF
run_trace "$program_9" "$input_9" "$output_9"
assert_grep "usage: break <line\\|file:line>" "$output_9" "expected break usage in scenario 9"
assert_grep "invalid line number: abc" "$output_9" "expected invalid line message in scenario 9"
assert_grep "missing file in breakpoint spec: :3" "$output_9" "expected missing file message in scenario 9"
assert_grep "invalid breakpoint id: x" "$output_9" "expected invalid breakpoint id message in scenario 9"
assert_grep "invalid frame index: abc" "$output_9" "expected invalid frame index message in scenario 9"
assert_grep "frame #99 out of range" "$output_9" "expected frame out-of-range message in scenario 9"
assert_grep "usage: locals \\[all\\]" "$output_9" "expected locals usage in scenario 9"
assert_grep "usage: print <expr>" "$output_9" "expected print usage in scenario 9"
assert_grep "usage: watch <expr>" "$output_9" "expected watch usage in scenario 9"
assert_grep "invalid watch id: foo" "$output_9" "expected invalid watch id message in scenario 9"
assert_grep "\\(no watches\\)" "$output_9" "expected no-watches message in scenario 9"
assert_grep "unknown command: unknowncmd" "$output_9" "expected unknown-command message in scenario 9"
assert_grep "result: 1" "$output_9" "expected final result in scenario 9"
assert_not_grep "runtime error:" "$output_9" "unexpected runtime error in scenario 9"

# Scenario 10: quit terminates session cleanly.
program_10="$tmp_dir/scenario10_quit.k"
input_10="$tmp_dir/scenario10_quit.in"
output_10="$tmp_dir/scenario10_quit.out"
cat > "$program_10" <<'EOF'
let x = 1
x
EOF
cat > "$input_10" <<'EOF'
quit
EOF
run_trace "$program_10" "$input_10" "$output_10"
assert_not_grep "runtime error:" "$output_10" "unexpected runtime error in scenario 10"
assert_not_grep "^result:" "$output_10" "quit should terminate without final result in scenario 10"

echo "PASS: debugger CLI e2e"
