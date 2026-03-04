# Karl Process + Stream M1 Contract

Status: implementation contract for first stream-native process/runtime pass.

## Goals

- Keep existing process model (`cmd` / `proc` / `run` / `wait <process>`).
- Replace process stdio exposure with streams (not channels).
- Add minimal file stream and transfer built-ins.
- Keep API explicit and short enough for shell-oriented usage.

## Core types

- `<process>`
- `<stream-reader>`
- `<stream-writer>`

## Built-ins

- `cmd({ command, args?, cwd?, env?, inheritEnv? }) -> <cmd>`
- `proc(cmdOrPipeline, opts?) -> <process>`
- `run(cmdOrPipeline, opts?) -> RunStatus`
- `reader(path, opts?) -> <stream-reader>`
- `writer(path, opts?) -> <stream-writer>`
- `pipe(srcReader, dstWriter, opts?) -> { bytes, chunks }`

## Process members

- `p.pid -> Int`
- `p.running -> Bool`
- `p.stdin -> <stream-writer>` (when `stdIn == "pipe"`)
- `p.stdout -> <stream-reader>` (when `stdOut == "pipe"`)
- `p.stderr -> <stream-reader>` (when `stdErr == "pipe"`)
- `p.abort() -> Unit`
- `p.kill() -> Unit`
- `p.signal(name) -> Unit`

## Stream members

- `stream.read(size?) -> [chunk, eof]`
  - `size` optional integer > 0
  - on EOF: `[null, true]`
- `stream.write(chunk) -> Int`
  - returns written byte count
- `stream.close() -> Unit`

## Constants

- `PIPE = "pipe"`
- `INHERIT = "inherit"`
- `NULL = "null"`
- `TEXT = "text"`
- `BYTES = "bytes"`

## Options

### `proc(..., opts)`

- `stdIn`, `stdOut`, `stdErr` in `"pipe" | "inherit" | "null"`
- `stdinType` / `stdInType` in `"text" | "bytes"` (default `"text"`)
- `stdoutType` / `stdOutType` in `"text" | "bytes"` (default `"text"`)
- `stderrType` / `stdErrType` in `"text" | "bytes"` (default `"text"`)
- `timeoutMs`

### `run(..., opts)`

- `stdin` (string)
- `timeoutMs`
- `maxOutputBytes`
- `overflow` in `"truncate" | "error"`

### `reader(path, opts)`

- `type` in `"bytes" | "text"` (default `"bytes"`)

### `writer(path, opts)`

- `type` in `"bytes" | "text"` (default `"bytes"`)
- `append` bool (default `false`)

### `pipe(src, dst, opts)`

- `bufferSize` integer > 0 (default `32768`)

## Errors (recoverable kinds)

- `process_spawn`
- `process_state`
- `process_io`
- `process_output_limit`
- `stream_open`
- `stream_read`
- `stream_write`
- `stream_close`
- `stream_state`

## M1 notes

- Stream chunks are represented as Karl strings in this phase.
- `BYTES`/`TEXT` mode is accepted at API level; true dedicated bytes value semantics can be expanded in M2.
- `readLine()` remains stdin-terminal helper and is not stream-member sugar.

