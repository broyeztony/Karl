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
- left operand must be `<stream-source>`, `<stream-plan>`, or compatible stream source handle (`<stream-reader>`, `<process>` with piped stdout)
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

## Graph model (DAG)

Every stream pipeline expression corresponds to a directed acyclic graph (DAG):
- source nodes produce values/chunks
- stage nodes transform values/chunks
- sink nodes consume values/chunks and terminate execution
- edges carry stream items downstream

Current runtime execution model:
- each `source | ... | sink` expression builds a linear DAG path
- a `<stream-plan>` is immutable: appending a stage returns a new plan
- execution starts only when a sink is attached
- execution is pull-driven from the sink (downstream demand drives upstream reads)

This gives deterministic linear dataflow semantics while keeping room for richer
DAG operators in later milestones.

## DAG composition today

Current stream primitives support linear pipelines directly. More complex DAG
topologies are composed with `Task` and `Channel` primitives around stream plans.

Pattern: run multiple pipelines concurrently:

```karl
let t = & {
  (() -> read("a.log") | lines() | stdout())(),
  (() -> read("b.log") | lines() | stdout())(),
}
wait t
```

Pattern: stream consumer emits to a channel:

```karl
let c = channel()
let p = proc({ command: "kubectl", args: ["logs", "-f", "api"], }, { stdout: PIPE, stderr: NULL, })

let producer = & (() -> {
  for true with r = p.stdout.read() {
    let [chunk, eof] = r
    if eof { break () }
    c.send(decodeUtf8(chunk))
    r = p.stdout.read()
  } then ()
})()
```

Notes:
- true branching/forking of one upstream stream into multiple downstream sinks
  (tee/broadcast semantics) is not yet a first-class stream operator.
- first-class fan-in operators are available (`merge`, `zip`, `join`).

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
- `stdin(opts?) -> <stream-source>`
  - `opts.type: "bytes" | "text"` (optional)
- `http(url, opts?) -> <stream-source>`
  - `url: String`
  - `opts.method: String` (optional, default `"GET"`)
  - `opts.headers: Object|Map` (optional)
  - `opts.body: String` (optional)
  - `opts.type: "bytes" | "text"` (optional, default `"bytes"`)
- `fromChannel(ch) -> <stream-source>`
  - `ch: <channel>`
- `merge(sourceOrPlanA, sourceOrPlanB, ... ) -> <stream-source>`
  - fan-in source combinator
- `zip(sourceOrPlanA, sourceOrPlanB) -> <stream-source>`
  - paired source combinator, stops at first EOF
- `join(leftSourceOrPlan, rightSourceOrPlan, leftKeyFn, rightKeyFn) -> <stream-source>`
  - keyed fan-in source combinator

### Pipeline stage

- `lines() -> <stream-stage>`
  - converts byte/text chunks into line-oriented text values
- `json() -> <stream-stage>`
  - decodes each upstream text/bytes item as one JSON value
- `map(fn) -> <stream-stage>`
- `filter(pred) -> <stream-stage>`
- `flatMap(fn) -> <stream-stage>` (`fn` must return `Array`)
- implicit lambda shorthand is supported in call arguments:
  - `filter(_ != "")`
  - `map("pod:" + _)`
- `distinct() -> <stream-stage>` (item must be hashable key type)
- `sort(cmp?) -> <stream-stage>`
  - default sort supports homogeneous `Int|Float|String|Char`
  - comparator form expects numeric return like array `sort`
- `take(n) -> <stream-stage>` (`n >= 0`)
- `drop(n) -> <stream-stage>` (`n >= 0`)
- `chunk(size) -> <stream-stage>` (`size > 0`, emits `Array` chunks)
- `window(size, step) -> <stream-stage>` (`size > 0`, `step > 0`, full windows only)
- `throttle(ms) -> <stream-stage>` (`ms >= 0`, delays between emitted items)
- `debounce(ms) -> <stream-stage>` (`ms >= 0`, keeps latest burst value)
- `tee(sink) -> <stream-stage>` (duplicates items to side sink and forwards downstream)
- `spill(path, opts?) -> <stream-stage>` (writes passing stream items and forwards downstream)

### Pipeline sinks

- `stdout() -> <stream-sink>` (execution result: `Unit`)
- `write(path, opts?) -> <stream-sink>` (execution result: `Unit`)
- `collect() -> <stream-sink>` (execution result: `Array`)
- `send(ch) -> <stream-sink>` (execution result: `Unit`, closes `ch` on completion)
- `reduce(init, fn) -> <stream-sink>` (execution result: reduced value)
- `count() -> <stream-sink>` (execution result: `Int`)
- `group_count(keyFn?) -> <stream-sink>` (execution result: `Map<Key, Int>`)
- `reduce_by_key(keyFn, init, reducerFn) -> <stream-sink>` (execution result: `Map<Key, Value>`)
- `top(n, scoreFn?) -> <stream-sink>` (execution result: `Array`)
- `partition(selector) -> <stream-sink>`
  - bool selector form result: `{ pass: <stream-source>, fail: <stream-source> }`
- `partition(selector, branchKeys) -> <stream-sink>`
  - keyed selector form result: object with one `<stream-source>` per key in `branchKeys`
  - selector must return `String`
  - `branchKeys` is `Array<String>` and must be non-empty + unique
- `split(pred) -> <stream-sink>` (execution result: `{ left: Array, right: Array }`)
- `toChannel(ch) -> <stream-sink>` (execution result: `Unit`, closes `ch` on completion)
- `exec(specOrCommand...) -> <stream-sink>` (execution result: `RunStatus`)

`write(path, opts?)` options:
- `type: "bytes" | "text"` (optional)
- `append: Bool` (optional, default `false`)

### Low-level stream/file built-ins

- `reader(path, opts?) -> <stream-reader>`
- `writer(path, opts?) -> <stream-writer>`

`reader`/`writer` options:
- `type: "bytes" | "text"` (optional, default `"bytes"`)
- `writer.append: Bool` (optional)

## Additional notes

- `send(ch)` and `toChannel(ch)` are both supported sink adapters from stream to channel.
- `join(...)` currently builds right-side key index when opening the source, then streams left-side joins.
- `partition(...)` branch sources are single-consumer.
- unconsumed partition branches are dropped by policy (no upstream blocking for unopened branches).

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

`<process>` is also accepted directly as a stream source shorthand for stdout:

```karl
let p = proc({ command: "kubectl", args: ["get", "pods", "-A"], }, { stdout: PIPE, stderr: NULL, stdoutType: TEXT, })
let rows = p | lines() | collect()
```

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

Channel bridge:

```karl
let out = buffered(128)
let t = & (read("events.log") | lines() | send(out))
```

Fan-in:

```karl
let merged = merge(
  read("a.log") | lines(),
  read("b.log") | lines(),
) | take(20) | collect()
```

Keyed join:

```karl
let out = join(
  read("left.log") | lines(),
  read("right.log") | lines(),
  l -> l.split("|")[0],
  r -> r.split("|")[0],
) | collect()
```

File write pipeline:

```karl
read("in.bin", { type: BYTES, }) | write("out.bin", { type: BYTES, })
```

Branch routing with keyed partition:

```karl
let p = read("events.log") | lines()
    | partition(
        line -> if line.contains("ERROR") { "err" } else if line.contains("WARN") { "warn" } else { "info" },
        ["err", "warn", "info"],
    )

let errOnly = p.err | collect()
```

JSON objects + keyed aggregation:

```karl
read("events.ndjson", { type: TEXT, })
| lines()
| json()
| reduce_by_key(e -> e.kind, 0, (acc, e) -> acc + 1)
```

Process sink from stream:

```karl
read("payload.txt", { type: TEXT, })
| exec({ command: "cat", })
```
