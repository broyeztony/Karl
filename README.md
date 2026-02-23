<p align="center">
  <img src="assets/karl.png" alt="Karl" />
</p>

[![CI](https://github.com/broyeztony/Karl/actions/workflows/ci.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/ci.yml)
[![Workflow Tests](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml)

Karl is a functional-first, expression-based programming language built on top of Go.
It is co-designed with AI.

Notably, it features:
- Functions as first-class entities
- Composable expression-first style
- Pattern matching with guards
- Property destructuring
- Expression-based control flow (`if`/`match`/`for` all return values)
- Recover operator (`?`)
- Concurrency model inspired by Go

Try Karl today in your browser: [karl-lang.org](https://karl-lang.org)

Watch the YouTube video: [**Karl Playground, Loom, Sheets & Jupyter Lab integration**](https://www.youtube.com/watch?v=DKqPl7-Rjg8)

## Start Here in 2 Commands

Install Karl from the latest release binaries:  
[https://github.com/broyeztony/Karl/releases/latest](https://github.com/broyeztony/Karl/releases/latest)

```bash
karl version
karl loom
```

## Start Here

### `bench` (Karl Playground)
`bench` is Karl's browser-first playground experience: run Karl instantly at [karl-lang.org](https://karl-lang.org), no install needed.  
Run it locally with `karl playground` (default `http://localhost:8081`). See [playground/README.md](playground/README.md).

### `loom` (Karl REPL)
`loom` is the interactive Karl REPL/runtime entrypoint for fast experimentation.  
Start it with:
```bash
karl loom
```
See [repl/README.md](repl/README.md) for local/remote modes.

### VS Code Plugin
Karl ships with a VS Code extension in `karl-vscode/` for syntax highlighting and editor support.  
Setup and usage: [karl-vscode/README.md](karl-vscode/README.md).

## Install Karl CLI (Latest Release)

Grab the latest release binaries here:
[https://github.com/broyeztony/Karl/releases/latest](https://github.com/broyeztony/Karl/releases/latest)

Download the archive for your OS/architecture, extract the binary, make it executable, and move it to your `PATH`.
Then run:
```bash
karl version
```

Minimal CLI usage:
```bash
karl version
karl run file.k
karl parse file.k
karl loom
```

## Project Map

- `examples/`  
  Feature-focused Karl programs, from basics to concurrency and workflow demos. Start with [examples/README.md](examples/README.md).

- `vigil_in_karl`  
  Full migration of the Vigil app to Karl (mock provider API, discovery worker, dedupe + PostgreSQL, Docker runbook): [broyeztony/vigil_in_karl](https://github.com/broyeztony/vigil_in_karl).

- Notebook + Jupyter integration (`karl notebook`, `kernel/`)  
  Run `.knb` notebooks from CLI and use Karl inside Jupyter Lab/Notebook via the Karl kernel. See [notebook/README.md](notebook/README.md) and [kernel/README.md](kernel/README.md).

- Karl Sheets (`karl spreadsheet`)  
  Reactive spreadsheet runtime where cells evaluate Karl expressions, served at `http://localhost:8080` by default.

- `Makefile`  
  Common developer workflow commands: `make build`, `make build-wasm`, `make build-all`, `make test`, `make examples`, `make workflow`, `make ci`.
