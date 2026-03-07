# Karl Stream Specification

This document is the normative specification for Karl stream values, stream
pipelines, stream built-ins, and stream runtime behavior.

## Scope

This specification defines:
- stream pipeline value kinds (`<stream-source>`, `<stream-stage>`, `<stream-sink>`, `<stream-plan>`)
- low-level stream handles (`<stream-reader>`, `<stream-writer>`)
- pipeline operator semantics (`|`)
- stream execution model and backpressure behavior
- stream built-ins and stream member methods
- recoverable stream error kinds

This specification does not define process lifecycle APIs; see `SPECS/process.md`.

## Runtime support

- Native runtime (non-WASM): stream APIs are supported.
- WASM/browser runtime: low-level stream/file APIs may be unavailable and return
  recoverable stream errors.

## Value kinds

- `<stream-source>`: pipeline source descriptor.
- `<stream-stage>`: pipeline transform descriptor.
- `<stream-sink>`: terminal consumer descriptor.
- `<stream-plan>`: source + zero or more stages, not yet executed.
- `<stream-reader>`: imperative read handle.
- `<stream-writer>`: imperative write handle.

## Pipeline operator `|`

`|` is reserved for stream pipelines.

Evaluation rules:
- left operand must be `<stream-source>`, `<stream-plan>`, or compatible stream source handle
- right operand must be `<stream-stage>` or `<stream-sink>`
- `source | stage` returns `<stream-plan>`
- `source | ... | sink` executes immediately in current task

Runtime validation errors:
- `operator '|' expects stream source or plan on the left`
- `operator '|' expects stream stage or sink on the right`

## Execution model

- Pipelines execute at sink application time.
- Sink execution is blocking in the current task.
- `source | stage` builds a lazy `<stream-plan>` and does not execute.
- To run a pipeline in background, wrap expression with `&`/`spawn`:

```karl
let t = & (read("events.log") | lines() | stdout())
wait t
```

Output behavior:
- `stdout()` writes to runtime console/output.
- concurrent pipelines can interleave console lines nondeterministically.

## Stream model invariants

Streams are:
- lazy
- pull-driven internally
- composable through operators
- executed only when consumed by sinks

Runtime implementations must not introduce push-based execution or implicit
buffering that violates backpressure guarantees.

## Built-ins

### Pipeline source

- `read(path, opts?) -> <stream-source>`
  - `path: String`
  - `opts.type: "bytes" | "text"` (optional)

### Pipeline stage

- `lines() -> <stream-stage>`
  - converts byte/text chunks into line-oriented text values

### Pipeline sinks

- `stdout() -> <stream-sink>` (execution result: `Unit`)
- `write(path, opts?) -> <stream-sink>` (execution result: `Unit`)
- `collect() -> <stream-sink>` (execution result: `Array`)

`write(path, opts?)` options:
- `type: "bytes" | "text"` (optional)
- `append: Bool` (optional, default `false`)

### Low-level stream/file built-ins

- `reader(path, opts?) -> <stream-reader>`
- `writer(path, opts?) -> <stream-writer>`
- `copy(srcReader, dstWriter, opts?) -> { bytes, chunks }`

`reader`/`writer` options:
- `type: "bytes" | "text"` (optional, default `"bytes"`)
- `writer.append: Bool` (optional)

`copy` options:
- `bufferSize: Int` (optional, default `32768`)

## Stream members

For `r: <stream-reader>`:
- `r.read(size?) -> [chunk, eof]`
- `r.close() -> Unit`

Rules:
- `size` optional integer > 0
- when `eof == true`, `chunk == null`
- in BYTES mode, `chunk` is `Bytes`
- in TEXT mode, `chunk` is `String`

For `w: <stream-writer>`:
- `w.write(chunk) -> Int`
- `w.close() -> Unit`

Rules:
- in BYTES mode, `chunk` must be `Bytes`
- in TEXT mode, `chunk` must be `String`
- return value is number of written bytes

## Constants

Prebound string constants:
- `PIPE = "pipe"`
- `INHERIT = "inherit"`
- `NULL = "null"`
- `TEXT = "text"`
- `BYTES = "bytes"`

## Process integration points

Process handles expose stream-compatible properties:
- `p.stdin`
- `p.stdout`
- `p.stderr`

These are defined in `SPECS/process.md` and can participate in stream usage
when process stdio mode is `"pipe"`.

## Error model

Streams use structural fail-fast behavior by default:
- stage/source/sink failures terminate current pipeline execution
- errors propagate as recoverable/runtime errors

Recoverable stream kinds:
- `stream_open`
- `stream_read`
- `stream_write`
- `stream_close`
- `stream_state`

## Backpressure

Karl stream pipelines are pull-driven from sinks:
- slower sink slows upstream stages and source
- faster source cannot outrun downstream consumers indefinitely
- memory usage remains bounded by stage and buffer strategy

## Examples

Blocking:

```karl
read("events.log") | lines() | stdout()
```

Collect:

```karl
let rows = read("events.log") | lines() | collect()
log(rows.length)
```

File copy:

```karl
let src = reader("in.bin", { type: BYTES, })
let dst = writer("out.bin", { type: BYTES, })
let st = copy(src, dst, { bufferSize: 65536, })
src.close()
dst.close()
log(st)
```
