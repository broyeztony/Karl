# Changelog

All notable changes to Karl are documented here.

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
