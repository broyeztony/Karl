<img src="assets/karl.png">

### The Karl programming language

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

Watch the YouTube video: [**Karl Playground, Loom, Sheets & Jupyter Lab integration**](https://www.youtube.com/watch?v=DKqPl7-Rjg8)

Try Karl today in your browser: [karl-lang.org](https://karl-lang.org)

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

macOS/Linux (`amd64` + `arm64`) one-liner:
```bash
os="$(uname -s | tr '[:upper:]' '[:lower:]')"; arch="$(uname -m)"; case "$arch" in x86_64) arch=amd64 ;; arm64|aarch64) arch=arm64 ;; *) echo "unsupported arch: $arch" && exit 1 ;; esac; base="karl-${os}-${arch}"; curl -fsSL "https://github.com/broyeztony/Karl/releases/latest/download/${base}.tar.gz" | tar -xz && mv "$base" karl && chmod +x karl
```

Then move `karl` into your `PATH` (for example `/usr/local/bin`).
All releases: [GitHub Releases](https://github.com/broyeztony/Karl/releases).

Minimal CLI usage:
```bash
karl run file.k
karl parse file.k
karl loom
```

## Project Map

- `examples/`  
  Feature-focused Karl programs, from basics to concurrency and workflow demos. Start with [examples/README.md](examples/README.md).

- Notebook + Jupyter integration (`karl notebook`, `kernel/`)  
  Run `.knb` notebooks from CLI and use Karl inside Jupyter Lab/Notebook via the Karl kernel. See [notebook/README.md](notebook/README.md) and [kernel/README.md](kernel/README.md).

- Karl Sheets (`karl spreadsheet`)  
  Reactive spreadsheet runtime where cells evaluate Karl expressions, served at `http://localhost:8080` by default.

- `Makefile`  
  Common developer workflow commands: `make build`, `make build-wasm`, `make build-all`, `make test`, `make examples`, `make workflow`, `make ci`.

## Specs

- `SPECS/language.md` — syntax + semantics
- `SPECS/interpreter.md` — runtime model and evaluator notes
- `SPECS/todolist.md` — short, current priorities for contributors
