
<img src="assets/karl.png">

### The Karl programming language

[![CI](https://github.com/broyeztony/Karl/actions/workflows/ci.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/ci.yml)
[![Workflow Tests](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml)

![Karl REPL](./assets/repl.png)

Watch the YouTube video: [**Karl: First Impressions (Live Language Demo)**](https://www.youtube.com/watch?v=pes-ZOvM0s0)

#### Tour of Karl

```
// Compose behavior with first-class functions.
let run = (value, fns) -> for i < fns.length with i = 0, out = value {
    out = fns[i](out)
    i++
} then out
let double = x -> x * 2
let inc = x -> x + 1
run(10, [double, inc, double]) // 42
```

```
// Match + guards (expression-based branching).
let tempo = 160
match tempo {
    case _ if tempo >= 180 -> "sprint"
    case 120..179 -> "groove"
    case _ -> "chill"
}
```

```
// Blocks and destructuring are expression-friendly.
let track = {
    let title = "Neon Steps"
    let bpm = 160
    { title: title, bpm: bpm, }
}
let { title, bpm, } = track
title + " @ " + str(bpm)
```

```
// `for` is an expression; `then` is a default when the loop doesn't break.
let nums = [3, 5, 8, 9]
for i < nums.length with i = 0 {
    if nums[i] % 2 == 0 { break nums[i] }
    i++
} then "none"
```

```
// Optional field access with object indexing + recover.
let getOr = (obj, key, fallback) -> obj[key] ? fallback
let req = decodeJson("{\"headers\":{\"User-Agent\":\"Karl\"}}")
let ua = getOr(req.headers, "User-Agent", "unknown")
ua
```

```
// Recover with either a block or a direct value.
let raw = "{ \"bpm\": 120 }"
let parsed = decodeJson(raw) ? { bpm: 90, }
let trace = parsed["X-Amzn-Trace-Id"] ? "<missing>"
{ bpm: parsed.bpm, trace: trace, }
```

```
// Async task continuation.
let delayed = () -> {
    sleep(50)
    "{ \"ok\": true }"
}
let t1 = & delayed()
let t2 = t1.then(body -> decodeJson(body))
wait t2
```

```
// Channels for task coordination.
let ch = channel()
let reader = & ch.recv()
let writer = & { ch.send("ping"); ch.done() }
wait writer
let [msg, done] = wait reader
msg
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

### Jupyter Kernel and Notebook Command

Use Karl notebooks from the CLI with:
```
karl notebook notebook/examples/01-quickstart.knb
karl notebook convert in.ipynb out.knb
```

To run Karl inside Jupyter Notebook/Lab, install the kernel with:
```
./kernel/install.sh
```
Then select **Karl** as the notebook kernel in Jupyter. More details: [kernel/README.md](kernel/README.md) and [notebook/README.md](notebook/README.md).

### CLI

```
karl parse <file.k> [--format=pretty|json]
karl run <file.k>
karl loom
cat <file.k> | karl run -
```

**REPL**: Start an interactive session with `karl loom`. See [repl/README.md](repl/README.md) for details.

**Playground**: Launch the web-based playground with `karl playground`. See [playground/README.md](playground/README.md) for details.

### VS Code

Karl ships with a VS Code extension in `karl-vscode/` for syntax highlighting and editor support.
See setup and usage details in [karl-vscode/README.md](karl-vscode/README.md).

### Tests

```
go test ./...
```

### Specs

- `SPECS/language.md` — syntax + semantics
- `SPECS/interpreter.md` — runtime model and evaluator notes
- `SPECS/todolist.md` — short, current priorities for contributors
