# Karl Process and Pipeline Specification

This document is the normative specification for Karl process execution, process pipelines, and process control.

## Scope

This specification defines:
- process value types (`<cmd>`, `<pipeline>`, `<process>`)
- process built-ins (`cmd`, `proc`, `run`, `stdIn`, `stdOut`, `stdErr`)
- pipeline composition via `|`
- process lifecycle and waiting semantics
- process status objects and recoverable error kinds

## Runtime support

- Native runtime (non-WASM): fully supported.
- WASM/browser runtime (Playground/Bench): process API is unavailable and returns recoverable errors.

Recoverable errors in unsupported runtimes use:
- `kind = "process_spawn"` for `cmd` / `proc` / `run`
- `kind = "process_state"` for `stdIn` / `stdOut` / `stdErr` / `wait <process>`

## Value types

- `<cmd>`: one process stage definition (not started).
- `<pipeline>`: ordered composition of process stages (not started).
- `<process>`: running process handle returned by `proc(...)`.

## Built-ins

### `cmd`

Signatures:
- `cmd(command, args?) -> <cmd>`
- `cmd({ command, args }) -> <cmd>`

Rules:
- `command` must be non-empty string.
- `args` must be array of strings when provided.

### `|`

Signature:
- `<cmd>|<pipeline> | <cmd>|<pipeline> -> <pipeline>`

Rules:
- `|` composes stages only; it does not start processes.
- Both operands must be `<cmd>` or `<pipeline>`, else runtime error.
- Bare prefix `| ...` is a parse error.

### `proc`

Signature:
- `proc(specOrPlan) -> <process>`

Accepted input:
- `<cmd>`
- `<pipeline>`
- object spec:
  - single-stage form: `{ command, args?, ... }`
  - plan form: `{ plan, ... }` where `plan` is `<cmd>` or `<pipeline>`

Process starts immediately and returns a waitable/abortable handle.

### `run`

Signature:
- `run(specOrPlan) -> RunStatus`

Accepted input:
- same forms as `proc` (`<cmd>`, `<pipeline>`, or object with `command/...` or `plan`).

`run` is blocking convenience API: equivalent behavior to launching and waiting, with captured output/error.

### `stdIn` / `stdOut` / `stdErr`

Signatures:
- `stdIn(process) -> Channel<String>`
- `stdOut(process) -> Channel<String>`
- `stdErr(process) -> Channel<String>`

Rules:
- Argument must be `<process>`.
- Channel accessor is valid only when corresponding mode is `"pipe"`.
- Otherwise recoverable `process_state` error.

### `wait`

`wait` accepts both `<task>` and `<process>`.

For process values:
- `wait process -> ProcessStatus`
- waiting a completed process returns cached status immediately.

## Process members

For `p: <process>`, supported members are:
- `p.pid -> Int`
- `p.running -> Bool`
- `p.abort() -> Unit`
- `p.kill() -> Unit`
- `p.signal(name) -> Unit`

Rules:
- `abort/kill/signal` on non-running process raise recoverable `process_state`.
- `signal(name)` requires string signal name; unknown name raises recoverable `process_state`.

## Object spec fields

### Shared (`proc` and `run`)

- `command: String`
- `args: [String]` (optional)
- `plan: <cmd>|<pipeline>` (optional; mutually exclusive with `command` in intent)
- `cwd: String` (optional)
- `env: Object|Map|ModuleObject` values must be strings (optional)
- `inheritEnv: Bool` (optional, default `true`)
- `timeoutMs: Int` (optional, default `0` = no timeout)

If both `plan` and `command` are present, `plan` is used.

### `proc`-specific stdio fields

- `stdIn: "pipe" | "inherit" | "null"` (default `"inherit"`)
- `stdOut: "pipe" | "inherit" | "null"` (default `"inherit"`)
- `stdErr: "pipe" | "inherit" | "null"` (default `"inherit"`)

### `run`-specific fields

- `stdin: String` (optional; when provided it is sent to process stdin)
- `maxOutputBytes: Int` (optional, default `1048576`)
- `overflow: "truncate" | "error"` (optional, default `"truncate"`)

`run` effective stdio defaults:
- stdin mode: `"null"` unless `stdin` field is provided
- stdout mode: `"pipe"`
- stderr mode: `"pipe"`

## Pipeline I/O semantics

For `stage1 | stage2 | ... | stageN`:
- stdout of stage `i` is connected to stdin of stage `i+1`.
- exposed `stdOut(process)` corresponds to stage `N` stdout.
- exposed `stdErr(process)` merges stderr from all stages.
- exposed `stdIn(process)` writes to stage `1` stdin.

## Channels and data shape

- `stdIn(process)` expects string chunks sent through channel.
- `stdOut/stdErr` yield string chunks.
- Standard channel receive shape applies: `[value, is_closed]`.
- Closing stdin channel closes underlying process stdin pipe.

## Status objects

### `ProcessStatus` (from `wait process`)

```
{
  ok: Bool,
  code: Int,          // exit code (or -1 when unavailable)
  signal: String|null,
  timedOut: Bool,
  aborted: Bool,
  durationMs: Int
}
```

`ok` is true iff:
- exit code is 0
- no terminating signal
- not timed out
- not aborted

### `RunStatus` (from `run`)

`RunStatus = ProcessStatus +`:

```
{
  output: String,
  error: String,
  outputTruncated: Bool,
  errorTruncated: Bool
}
```

Notes:
- non-zero exits return `RunStatus` (`ok=false`), they do not raise by themselves.
- overflow behavior:
  - `"truncate"`: output/error truncated to max bytes and `*Truncated=true`
  - `"error"`: recoverable `process_output_limit`

## Recoverable errors

Kinds used by process APIs:
- `process_spawn`: spawn/start/build failures (including unsupported runtime)
- `process_state`: invalid lifecycle/state or accessor usage
- `process_io`: stream forwarding or I/O failures after spawn
- `process_output_limit`: `run(... overflow: "error")` exceeded capture limit
- `canceled`: waiting task was canceled while awaiting process completion

## Parser constraints

- `spawn`/`&` require call expression target or group of call expressions.
- `proc(...)` is rejected as spawn target:
  - `& proc(...)` parse error
  - `spawn proc(...)` parse error
  - `spawn { proc(...), ... }` parse error

Rationale: `proc(...)` already yields `<process>` and starts execution immediately.

## Examples

### Blocking run with captured output

```karl
let st = run({ command: "echo", args: ["hello"], })
if !st.ok { exit("command failed: " + str(st.code)) }
log(st.output)
```

### Managed process with pipes

```karl
let p = proc({ command: "cat", stdIn: "pipe", stdOut: "pipe", stdErr: "pipe", })
let inCh = stdIn(p)
inCh.send("hello\n")
inCh.done()
let [line, _closed] = stdOut(p).recv()
log(line)
let st = wait p
```

### Pipeline

```karl
let plan = cmd("printf", ["alpha\nbeta\n"]) | cmd("wc", ["-l"])
let p = proc({ plan: plan, stdOut: "pipe", stdErr: "pipe", stdIn: "null", })
let [count, _] = stdOut(p).recv()
log(count)
wait p
```
