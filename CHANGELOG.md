# Changelog

All notable changes to Karl are documented here.

## [v0.4.1] - 2026-02-15

Highlights:
- Fixed fail-fast behavior for canceled tasks in concurrency examples and runtime checks.

Concurrency:
- Canceled async tasks are now treated as expected control flow in fail-fast mode.
- `task.cancel()` no longer risks being surfaced as an unhandled fatal task failure due to timing races.
- Added regression coverage for canceled detached tasks and canceled blocked `recv()` tasks.

## [v0.4.0] - 2026-02-15

Highlights:
- Added an interactive REPL with both local and remote modes.
- Introduced the `karl loom` command family for REPL workflows.
- Added native string slicing with range syntax (`"text"[start..end]`).
- Refactored the interpreter into a modular architecture (no intended behavior change).

REPL & CLI:
- Added `karl loom` for local interactive sessions.
- Added `karl loom serve` and `karl loom connect <host:port>` for remote REPL sessions.
- Kept compatibility aliases for `repl`, `repl-server`, and `repl-client`.
- Added REPL docs and examples under `repl/`.

Language:
- Added string slicing support via existing slice syntax (`s[start..end]`), including open bounds and negative indices.

Interpreter:
- Split monolithic evaluator/builtins/runtime/value code into focused modules.
- Moved import resolution/loading/evaluation into dedicated files.
- Added regression tests for concurrency task-failure policy behavior.

Examples & tooling:
- Reorganized concurrency examples under `examples/features/concurrency/`.
- Removed legacy `examples/programs/*` in favor of feature-focused examples.
- Added helper scripts for example runtime validation and example diffs.

## [v0.3.5] - 2026-02-10

Highlights:
- Fixed Windows release packaging in GitHub Actions.

CI/CD:
- Release workflow now uses native PowerShell `Compress-Archive` on Windows runners.
- Unix release packaging remains `tar.gz`.

## [v0.3.4] - 2026-02-10

Highlights:
- Concurrency task-failure model updated with explicit policies.
- Default `karl run` behavior is now `fail-fast` for unobserved detached task failures.
- `defer` mode remains available for end-of-run unhandled failure reporting.

Concurrency:
- Task errors are stored on task handles and surface on `wait` / `.then(...)`.
- `wait task ? { ... }` remains recoverable for observed task failures.
- Join/race and cancellation behavior is documented with updated feature examples.

CLI:
- Added `--task-failure-policy=fail-fast|defer` to `karl run`.

Docs & examples:
- Updated concurrency docs in specs.
- Updated `examples/features/concurrency/*` to match the new semantics.

## [v0.3.2] - 2026-02-03

Highlights:
- Added buffered channels via `buffered(size)`.
- Added `channel()` as an alias for `rendezvous()`.
- CLI no longer prints `()` when the program result is `Unit`.

Docs & examples:
- Documented `channel()` alias in specs and examples.
- Added a buffered channels feature example.

Developer:
- Ignore `.vscode/` in `.gitignore`.

## [v0.3.1] - 2026-02-03

Highlights:
- Truthy/falsy semantics across conditionals and logical operators (Python-style; empty map/set are falsy).
- New truthy/falsy feature examples.
- Added a full workflow-engine example suite (contrib).

Interpreter:
- `if`, `for`, and match guards now use truthy/falsy rules.
- `!`, `&&`, and `||` operate on truthy/falsy and return `Bool`.
- Empty `map()` and `set()` are falsy; non-empty are truthy.

Tests:
- Added unit tests covering truthy/falsy evaluation, short-circuiting, loop conditions, and guards.

Examples:
- Added `examples/features/truthy_falsy.k` and `examples/features/truthy_falsy_comprehensive.k`.
- Added `examples/contrib/workflow/*` (DAG engine, pipelines, timers, file watcher, quickstart).

Docs:
- Updated `SPECS/interpreter.md` with truthy/falsy rules.
- Updated `examples/README.md` with new examples.

Developer:
- Ignore `*.karl-new` and `karl` artifacts in `.gitignore`.
