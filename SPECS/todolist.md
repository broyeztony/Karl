# Karl TODO

## Current priorities

- runbook blocker: select-like concurrency primitive (wait on channel/task/timer in one construct)
- runbook blocker: structured errors (kind/code/data) + consistent propagation/recovery patterns
- runbook blocker: guaranteed cleanup primitive (`defer`/`finally`)
- runbook blocker: first-class time/duration ergonomics (durations, deadlines, timeout composition)
- runbook blocker: safer optional access ergonomics (avoid missing-property footguns in workflows)
- runbook blocker: durable state/checkpoint API for crash-safe resume
- runbook blocker: scheduler/trigger runtime (cron, interval, event/webhook)
- runbook blocker: process execution primitive (I/O now covers env/argv/stdin scan, fs, and http client/server)
- runbook blocker: secrets/config boundary with redaction-safe error/log behavior
- runbook blocker: observability API (structured logs, step events, metrics, correlation IDs)
- phase-1 runtime I/O primitives (`argv`, `programPath`, `environ`, `env`, `readLine`) ✅
- add a httpServer built-in ✅
- add runtime foundations for Vigil port (sql*, signalWatch, uuid/time helpers, sha256) ✅
- build a debugger (breakpoints, step in/over/out, stack/locals inspection; CLI + DAP) ✅
- add debugger regression coverage in CI (DAP tests + CLI e2e scenarios) ✅
- add VSCode/Cursor GUI e2e debugger checks (editor navigation/highlight assertions across platforms)
- build a notebook system for Karl like Jupyter using the repl server
- add binary data support
- change divide-by-zero semantics: raise runtime error instead of returning Inf/NaN ✅
- Keep tests green as syntax/runtime changes land (`gotest`).
- Parser: consider treating newlines as statement boundaries to reduce adjacency ambiguity.
- Extend test coverage when new syntax is added (parser + interpreter + examples).
- Brainstorm objects versus maps versus mutability versus shapes
- recover block that runs for runtime throw situations (`expr ? { ... }`) ✅
- string interpolation
- make a <task> cancelable ✅


## Known review points

- Disambiguation rules (block vs object) must stay consistent with examples and tests ✅
- Import/factory behavior and live exports should remain explicit in specs.
