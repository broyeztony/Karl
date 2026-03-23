# Karl TODO

## Status

- Vigil is fully implemented and runnable in Karl (`vigil_in_karl`).
- Blockers below are now scoped to production hardening, not basic runbook viability.

## Runbook blockers (production hardening)

- structured errors (kind/code/data) + consistent propagation/recovery patterns
- durable state/checkpoint API for crash-safe resume
- secrets/config boundary with redaction-safe error/log behavior
- observability API (structured logs, step events, metrics, correlation IDs)

## Next language/runtime improvements

- select-like concurrency primitive (wait on channel/task/timer in one construct)
- guaranteed cleanup primitive (`defer`/`finally`)
- first-class time/duration ergonomics (durations, deadlines, timeout composition)
- safer optional access ergonomics (avoid missing-property footguns in workflows)
- scheduler/trigger runtime (cron, interval, event/webhook)
- add VSCode/Cursor GUI e2e debugger checks (editor navigation/highlight assertions across platforms)
- Parser: consider treating newlines as statement boundaries to reduce adjacency ambiguity
- Brainstorm objects versus maps versus mutability versus shapes
- string interpolation

## Engineering hygiene

- Keep tests green as syntax/runtime changes land (`gotest`)
- Extend test coverage when new syntax is added (parser + interpreter + examples)

## Completed foundations

- phase-1 runtime I/O primitives (`argv`, `programPath`, `environ`, `env`, `readLine`) ✅
- add a httpServer built-in ✅
- add runtime foundations for Vigil port (sql*, signalWatch, uuid/time helpers, sha256) ✅
- process execution primitive ✅
- add binary data support ✅
- build a debugger (breakpoints, step in/over/out, stack/locals inspection; CLI + DAP) ✅
- add debugger regression coverage in CI (DAP tests + CLI e2e scenarios) ✅
- recover block that runs for runtime throw situations (`expr ? { ... }`) ✅
- change divide-by-zero semantics: raise runtime error instead of returning Inf/NaN ✅
- make a `<task>` cancelable ✅


## Known review points

- Disambiguation rules (block vs object) must stay consistent with examples and tests ✅
- Import/factory behavior and live exports should remain explicit in specs.
