# Karl Debugger (CLI First) — Minimal Spec

## Scope

Build a Karl debugger that works first in the terminal, then expose the same engine to DAP (VS Code, etc).

Phase-1 user features:
- breakpoints (`file:line`)
- continue / pause
- step in / step over / step out
- stack frames
- locals inspection
- expression watch/eval in current frame

Non-goals for phase 1:
- time-travel debugging
- reverse execution
- full graphical debugger

## User-facing CLI

New command:
- `karl debug <file.k|-> [--task-failure-policy=...] -- [program args...]`

Suggested REPL-like debug prompt:
- `break <line>` / `break <file>:<line>`
- `delete <id>`
- `continue` (`c`)
- `pause`
- `next` (`n`)        // step over
- `step` (`s`)        // step in
- `finish` (`f`)      // step out
- `stack`
- `frame <idx>`
- `locals`
- `print <expr>`      // evaluate in selected frame
- `watch <expr>`
- `unwatch <id>`
- `quit`

## Runtime model

### 1) Execution hooks in evaluator
Add lightweight debug hooks at statement/expression boundaries:
- `beforeNode(node, env, frame)`
- optional `afterNode(...)` for richer stepping and tracing

Hooks must carry:
- source location (file, line, column)
- node kind
- current frame id/depth

No debugger attached => near-zero overhead branch and return.

### 2) Frame tracking
Track call frames explicitly:
- function call pushes frame
- return pops frame
- frame stores function label, location, and lexical environment handle

This is required for:
- `stack`
- `frame`
- `locals`
- `print <expr>` in selected frame

### 3) Breakpoint and step engine
Core controller state:
- breakpoint table (file:line)
- pause state
- step mode (`in`, `over`, `out`)
- step target depth (for over/out)

At each hook:
- if location matches breakpoint OR step condition is satisfied => pause
- pause blocks execution until next command from debugger shell

## Concurrency behavior (phase 1)

Karl has spawned tasks, so debugger behavior must be explicit.

Phase-1 policy:
- Debugger controls the main evaluation task only.
- Breakpoints in child tasks are not guaranteed.
- Child-task failures still follow existing runtime semantics.

Reason:
- keeps implementation small and predictable for first release.
- avoids global stop-the-world complexity in the first cut.

Phase 2 can add global task-aware debugging.

## Expression evaluation in debugger

`print <expr>` and `watch <expr>` evaluate in selected frame environment.

Phase-1 rule:
- allow normal expression evaluation
- document that expressions with side effects are user responsibility

Follow-up hardening (later):
- optional "safe-eval" subset for watches.

## Error semantics

- Runtime errors while debugging still surface with Karl source spans.
- Debug commands that fail (bad frame index, parse error in `print`) return debugger errors, not interpreter crashes.
- `exit()` remains a hard process stop (as currently designed), even in debug mode.

## Integration plan

### Milestone 1 — Debug engine foundation
- add evaluator hooks
- add frame tracking
- add breakpoint table + step state
- add unit tests for hook/step transitions

### Milestone 2 — CLI debugger command
- implement `karl debug ...`
- implement interactive debug prompt commands
- add integration tests (breakpoint hit, step over/in/out, locals, stack)

### Milestone 3 — DAP bridge
- map debugger engine to DAP requests/events
- support VS Code extension integration

## Test matrix (minimum)

1. Breakpoint hit at file+line in top-level code.
2. Breakpoint hit inside called function.
3. `step` enters function call.
4. `next` does not enter function call.
5. `finish` resumes until current frame returns.
6. `stack` returns expected depth/order.
7. `locals` shows bindings from selected frame.
8. `print` evaluates identifiers/member access correctly.
9. Program args after `--` still passed correctly in debug mode.
10. Existing non-debug `run` behavior unchanged.

## Open decisions (to lock before implementation)

1. Should phase-1 debugger pause all tasks or only main task? (spec currently: main task only)
2. Should watch expressions be side-effect free by default? (spec currently: no)
3. Keep command names terse (`n/s/f`) only, or both long+short aliases? (recommended: both)
