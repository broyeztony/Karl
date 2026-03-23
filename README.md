<p align="center">
  <img src="assets/karl.png" alt="Karl" />
</p>

[![CI](https://github.com/broyeztony/Karl/actions/workflows/ci.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/ci.yml)
[![Workflow Tests](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml/badge.svg)](https://github.com/broyeztony/Karl/actions/workflows/workflow-tests.yml)

A modern language for building command-line tools, pipelines, and infrastructure automation.

Karl is a programming language designed for:
- CLI tools
- pipelines
- infrastructure scripting
- process orchestration
- DevOps automation

Karl aims to occupy the space between:
- Bash
- Go

Notably, it features:
- Functions as first-class entities
- Composable expression-first style
- Pattern matching with guards
- Property destructuring
- Expression-based control flow (`if`/`match`/`for` all return values)
- Recover operator (`?`)
- Concurrency model inspired by Go
- Stream pipelines with `|` (`read(...) | lines() | stdout()`)

Website: [karl-lang.org](https://karl-lang.org)

Try Karl today in your browser (`bench`): [karl-lang.org/bench](https://karl-lang.org/bench)

Docs: [https://karl-lang.org/docs/](https://karl-lang.org/docs/)

## Start Here in 2 Commands

Install Karl from the latest release binaries:  
[https://github.com/broyeztony/Karl/releases/latest](https://github.com/broyeztony/Karl/releases/latest)

```bash
karl version
karl loom
```

## Next Steps

### `bench` (Karl Playground)
`bench` is Karl's browser-first playground experience: run Karl instantly at [karl-lang.org/bench](https://karl-lang.org/bench), no install needed.  
Run it locally with `karl playground` (default `http://localhost:8081/bench`). See [playground/README.md](playground/README.md).

### Editor Plugins (VS Code + Sublime)
Karl ships with editor plugins in `plugins/`:
- VS Code/Cursor extension: `plugins/karl-vscode/`
- Sublime Text syntax/theme: `plugins/karl-sublime/`

Install instructions: [plugins/README.md](plugins/README.md).

### Notebook + Jupyter Integration
Run `.knb` notebooks with `karl notebook`, or use Karl from Jupyter Lab/Notebook through the Karl kernel.  
See [notebook/README.md](notebook/README.md) and [kernel/README.md](kernel/README.md).

### Karl Sheets App
Use `karl spreadsheet` to run the reactive spreadsheet app where cells evaluate Karl expressions.  
Try it in your browser at [karl-lang.org/sheets](https://karl-lang.org/sheets).  
By default it serves on `http://localhost:8080`.

## CLI Commands

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
