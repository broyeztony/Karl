
<img src="assets/karl.png">

### The Karl programming language

[![CI](https://github.com/broyeztony/Karl/actions/workflows/ci.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/ci.yml)
[![Workflow Tests](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml)

<figure>
  <img src="assets/vscode.png" alt="Visual Studio Code plugin for Karl">
  <figcaption>Visual Studio Code plugin for Karl</figcaption>
</figure>

#### Tour of Karl

```
// Closures as first-class expressions.
let addN = n -> x -> n + x
let add5 = addN(5)
add5(10)
```

```
// Match + guards (expression-based branching).
let tempo = 160
let feel = match tempo {
    case _ if tempo >= 180 -> "🔥"
    case 120..179 -> "🎶"
    case _ -> "chill"
}
feel
```

```
// Blocks are expressions and return their last value.
let computed = {
    let x = 1
    let y = 3
    x + y
}
computed
```

```
// Destructuring + object literals (trailing comma disambiguates from blocks).
let track = { title: "Neon Steps", bpm: 160, }
let { title, bpm, } = track
title + " @ " + str(bpm)
```

```
// Truthy/falsy in conditionals.
let items = []
let label = if items { "non-empty" } else { "empty" }
label
```

```
// Object indexing for external keys.
let headers = decodeJson("{\"User-Agent\":\"Karl\"}")
headers["User-Agent"]
```

```
// `for` is an expression; `then` is a default when the loop doesn't break.
let nums = [3, 5, 8, 9]
let firstEven = for i < nums.length with i = 0 {
    if nums[i] % 2 == 0 { break nums[i] }
    i++
} then "none"
firstEven
```

```
// Recoverable errors with `? { ... }`.
let raw = "{ \"bpm\": 120 }"
let data = decodeJson(raw) ? { bpm: 90, }
data.bpm
```

```
// Async tasks (`&`) and channels.
let ch = channel()
let producer = & { ch.send("ready"); ch.done() }
let consumer = & ch.recv()
wait producer
wait consumer
```

Explore more examples in the `examples/` folder: [Karl Examples](examples/README.md)

#### Workflow examples (contrib by [Nico](https://github.com/hellonico))

The `examples/contrib/workflow/` folder is a small workflow engine and a set of demos built on top of it:

- `engine.k` — core workflow runner (sequential, parallel, DAG)
- `quickstart.k` — smallest end‑to‑end example
- `examples.k` — multiple workflow variants in one file
- `dag_pipeline.k` — large, multi‑stage data pipeline
- `subdag_demo.k` — composing workflows with sub‑DAGs
- `csv_pipeline.k` — data pipeline with validation + stats
- `file_watcher.k` — event‑driven workflow on file changes
- `timer_tasks.k` — scheduled/recurring task demos
- `test_simple_dag.k` — minimal DAG sanity check

**Running Workflow Tests:**

```bash
cd examples/contrib/workflow
./run_all_tests.sh
```

This runs all 13 tests including unit tests, integration tests, demos, and pipeline examples.

### Get Karl

Grab a prebuilt binary from GitHub Releases:
[Releases](https://github.com/broyeztony/Karl/releases)

Make it executable (`chmod +x <binary>`), then add it to your `PATH` or drop it in your `~/go/bin` directory.

Or build from source:

```
go build -o karl .
```

### CLI

```
karl parse <file.k> [--format=pretty|json]
karl run <file.k>
cat <file.k> | karl run -
```

### Tests

```
go test ./...
```

### Specs

- `SPECS/language.md` — syntax + semantics
- `SPECS/interpreter.md` — runtime model and evaluator notes
- `SPECS/todolist.md` — short, current priorities for contributors
