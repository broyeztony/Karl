# Karl Process Specification

This document is the normative specification for Karl process creation,
process lifecycle control, process status reporting, and process-specific
recoverable errors.

## Scope

This specification defines:
- process value type (`<process>`)
- process built-ins (`proc`, `run`)
- process spec object fields
- process options and defaults
- process member access and control methods
- process status objects (`ProcessStatus`, `RunStatus`)
- process-related recoverable error kinds

This specification does not define generic stream or pipeline semantics; see
`SPECS/stream.md`.

## Runtime support

- Native runtime (non-WASM): fully supported.
- WASM/browser runtime: process APIs are unavailable and return recoverable errors.

Unsupported runtime recoverable kind:
- `process_spawn`

## Built-ins

### Constants

Prebound string constants:
- `PIPE = "pipe"`
- `INHERIT = "inherit"`
- `NULL = "null"`
- `TEXT = "text"`
- `BYTES = "bytes"`

### `proc`

Signature:
- `proc(spec, opts?) -> <process>`

`spec` must be a process spec object (see "Process spec object").

Behavior:
- starts process execution immediately
- returns a waitable and abortable handle

### `run`

Signature:
- `run(spec, opts?) -> RunStatus`

Behavior:
- blocking convenience API
- equivalent to start + wait with captured stdout/stderr

### `wait`

`wait` accepts `<task>` and `<process>`.

For process handles:
- `wait p -> ProcessStatus`
- waiting a completed process returns cached status

## Process spec object

`spec` shape:

```karl
{
  command: "sh",                 // required, non-empty String
  args: ["-c", "echo hi"],       // optional [String]
  cwd: "/tmp",                   // optional String
  env: { KEY: "value", },        // optional Object/Map/ModuleObject; values must be String
  inheritEnv: true,              // optional Bool, default true
}
```

Validation rules:
- `command` required and non-empty
- `args` must contain only strings when provided
- `cwd` must be string when provided
- `env` keys/values must be strings after conversion
- `inheritEnv` must be bool when provided

## Options

### `proc(..., opts?)`

Supported fields:
- `stdIn` / `stdin`: `"pipe" | "inherit" | "null"` (default `"inherit"`)
- `stdOut` / `stdout`: `"pipe" | "inherit" | "null"` (default `"inherit"`)
- `stdErr` / `stderr`: `"pipe" | "inherit" | "null"` (default `"inherit"`)
- `stdinType` / `stdInType`: `"text" | "bytes"` (default `"bytes"`)
- `stdoutType` / `stdOutType`: `"text" | "bytes"` (default `"bytes"`)
- `stderrType` / `stdErrType`: `"text" | "bytes"` (default `"bytes"`)
- `timeoutMs: Int` (default `0`, no timeout)

### `run(..., opts?)`

Supported fields:
- `stdin: String` (optional)
- `timeoutMs: Int` (default `0`)
- `maxOutputBytes: Int` (default `1048576`)
- `overflow: "truncate" | "error"` (default `"truncate"`)

`run` effective stdio defaults:
- stdin: `"null"` unless `stdin` provided
- stdout: `"pipe"`
- stderr: `"pipe"`

## Process members

For `p: <process>`:
- `p.pid -> Int`
- `p.running -> Bool`
- `p.stdin -> <stream-writer>` (only when stdin mode is `"pipe"`)
- `p.stdout -> <stream-reader>` (only when stdout mode is `"pipe"`)
- `p.stderr -> <stream-reader>` (only when stderr mode is `"pipe"`)
- `p.abort() -> Unit`
- `p.kill() -> Unit`
- `p.signal(name) -> Unit`

Rules:
- accessing unavailable stdio properties raises recoverable `process_state`
- `abort`, `kill`, `signal` on non-running process raise recoverable `process_state`
- `signal(name)` requires valid string signal name

Stream handle method semantics are defined in `SPECS/stream.md`.

## Status objects

### `ProcessStatus` (`wait p`)

```karl
{
  ok: Bool,
  code: Int,          // exit code, -1 when unavailable
  signal: String|null,
  timedOut: Bool,
  aborted: Bool,
  durationMs: Int,
}
```

`ok` is true iff all are true:
- exit code is `0`
- `signal == null`
- `timedOut == false`
- `aborted == false`

### `RunStatus` (`run(...)`)

`RunStatus` extends `ProcessStatus` with:

```karl
{
  output: String,
  error: String,
  outputTruncated: Bool,
  errorTruncated: Bool,
}
```

Notes:
- non-zero exit is returned as `ok: false` status; it is not an implicit throw
- `overflow: "truncate"` truncates captured output and sets truncation flags
- `overflow: "error"` raises recoverable `process_output_limit` when cap exceeded

## Recoverable errors

Process-related recoverable kinds:
- `process_spawn`
- `process_state`
- `process_io`
- `process_output_limit`
- `canceled` (waiting task canceled during process wait)

Stream kinds used by process stream handles are specified in `SPECS/stream.md`.

## Parser constraints

- `proc(...)` returns a process handle and starts immediately.
- `& proc(...)` and `spawn proc(...)` are parse-rejected.
- For async process usage, call `proc(...)` directly and then `wait` the handle.

## Examples

### Blocking run

```karl
let st = run({ command: "echo", args: ["hello"], })
if !st.ok { exit("command failed: " + str(st.code)) }
log(st.output)
```

### Managed process with piped stdio

```karl
let p = proc(
  { command: "cat", },
  { stdIn: PIPE, stdOut: PIPE, stdErr: PIPE, stdinType: BYTES, stdoutType: BYTES, stderrType: BYTES, }
)

p.stdin.write(encodeUtf8("hello\n"))
p.stdin.close()
let [chunk, eof] = p.stdout.read()
if !eof { log(decodeUtf8(chunk)) }
let st = wait p
log(st)
```
