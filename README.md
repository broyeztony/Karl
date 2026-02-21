
<img src="assets/karl.png">

### The Karl programming language

[![CI](https://github.com/broyeztony/Karl/actions/workflows/ci.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/ci.yml)
[![Workflow Tests](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml)

Karl is a functional-first, expression-based programming language built on top on Go.
It is co-designed with AI.

Notably, it features: 
 - Functions as first-class entities
 - Encourages composability
 - Pattern matching with guards
 - Properties destructuring
 - Language constructs as expressions: if/match/for-loop contructs returns a value
 - A recover operator
 - A concurrency model inspired by Go

Watch the YouTube video: [**Karl Playground, Loom, Sheets & Jupyter Lab integration**](https://www.youtube.com/watch?v=DKqPl7-Rjg8)

Try Karl today in your browser: [karl-lang.org](https://karl-lang.org)

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

### Karl Playground

Use Karl directly in the browser at [karl-lang.org](https://karl-lang.org).
You can also run it locally with `karl playground` (default `http://localhost:8081`).
More details: [playground/README.md](playground/README.md).

### Karl Sheets

Karl includes a reactive spreadsheet runtime where cell formulas are Karl expressions.
Start it with `karl spreadsheet` (default `http://localhost:8080`).
The web app lives in `assets/spreadsheet/`.

### Karl Notebook and Jupyter Kernel

Use Karl notebooks from the CLI:
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
karl run <file.k> [--task-failure-policy=fail-fast|defer] [-- <program args...>]
karl loom
cat <file.k> | karl run -
```

**REPL**: Start an interactive session with `karl loom`. See [repl/README.md](repl/README.md) for details.

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
