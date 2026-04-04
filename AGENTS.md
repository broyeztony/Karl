# AGENTS.md

Agents working in this repository must follow the design documents and maintain
semantic consistency across the language runtime.

## Product Context

Karl is a modern language for building command-line tools, pipelines, and
infrastructure automation.

Primary use cases:
- CLI tools
- pipelines
- infrastructure scripting
- process orchestration
- DevOps automation

Karl occupies the space between Bash and Go.

## Source of Truth

Use these sources in order:
1. `SPECS/` (all specs are relevant design input)
2. `tests/` (behavioral contract)
3. implementation (`parser/`, `interpreter/`, `lexer/`, `token/`, CLI)

Notes:
- `SPECS/stream.md` and `SPECS/process.md` are scoped specs, not global supersets.
- If semantics change, update the relevant spec files in the same change set.

## Core Karl Concepts

Karl has four primary abstractions:
- `Stream<T>`
- `Task<T>`
- `Channel<T>`
- `Sink<T>`

Conceptual roles:
- Tasks compute
- Channels communicate
- Streams flow
- Sinks terminate pipelines

Preserve this separation. Do not collapse these abstractions or blur semantics.

Error handling model:
- recoverable/runtime errors with explicit recovery via `? { ... }`
- structural fail-fast by default in stream/process execution paths

## Stream Model Guardrails

Streams are:
- lazy
- pull-driven internally
- composable through operators
- executed only when consumed by sinks

Pipelines must obey `SPECS/stream.md`.
Do not introduce push-based execution or implicit buffering that violates
backpressure semantics.

## Implementation Rules

When implementing features:
1. Read the relevant spec sections.
2. Identify the minimal required change.
3. Preserve existing architecture.
4. Prefer simple implementations over clever ones.
5. Add tests covering specified behavior.

Avoid speculative features or architectural redesigns.

Additional non-negotiables:
- No implicit lifecycle/runtime magic without explicit approval.
- No aliases by default unless explicitly requested.
- Keep behavior predictable across CLI, REPL, Playground, and tests.
- Keep repo skills synchronized with Codex global skills (`$CODEX_HOME/skills`,
  default `~/.codex/skills`) whenever skills are added or updated.
  - Install hooks once: `make install-skill-hooks`
  - Manual sync: `make sync-skills`
- For API demonstrations (docs, landing snippets, bench examples), prefer one
  `log(...)` per API call/result. Avoid dumping large aggregated objects (for
  example `log(toJson({ ...many fields... }))`) when readability suffers.
- Runtime deadlock probing must not misclassify external-event waits as
  deadlocks (for example, top-level `signalWatch(...).recv()` must wait).
- Prefer single-instance module import shorthand:
  - `let x = (import "path/module.k")()`
- Use two-step factory imports only when multiple independent instances are needed:
  - `let makeX = import "path/module.k"; let a = makeX(); let b = makeX()`

## Development Workflow

Before implementing:
- summarize relevant spec rules
- identify affected files
- state assumptions

After implementing:
- explain what changed
- explain why it matches the spec
- list unresolved ambiguities (if any)

For non-obvious concepts, include concise examples in specs/docs/examples.

## Testing

All behavioral changes must include tests.

Tests should validate, when applicable:
- stream semantics
- error propagation
- cancellation behavior
- concurrency correctness
- memory safety characteristics
- deadlock probe correctness (true deadlocks vs valid external waits)

Suggested command flow:
- targeted tests first
- then `go test ./...`
- for debugger/process/streams, also run the relevant `make` targets

Useful commands:
- `make test`
- `make test-nocache`
- `make test-debugger`
- `make bench-copy`
- `make examples`
- `make workflow`

## Design Philosophy

Karl favors:
- small conceptual surface
- composable primitives
- predictable runtime behavior
- explicit error handling
- strong streaming guarantees

Agents must preserve these principles.

## Repo Map

- `main.go` - CLI entrypoint
- `lexer/`, `parser/`, `ast/`, `token/` - language front-end
- `interpreter/` - evaluator, runtime, built-ins
- `tests/` - language/runtime behavior tests
- `debugger/` - DAP/debugger
- `playground/`, `assets/playground/` - bench + wasm assets
- `plugins/` - editor plugins (VS Code, Sublime)
- `examples/` - feature and workflow examples
- `SPECS/` - design specifications
