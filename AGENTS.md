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

<!-- gitnexus:start -->
# GitNexus — Code Intelligence

This project is indexed by GitNexus as **Karl** (4293 symbols, 20874 relationships, 358 execution flows).

> Index stale? Run `node .gitnexus/run.cjs analyze --index-only` from the project root — it auto-selects an available runner. No `.gitnexus/run.cjs` yet? Bootstrap with `npx`, `bunx`, or `pnpm dlx` — e.g. `bunx gitnexus@latest analyze` (npm 11 npx crash; #1939).

## Always Do

- **MUST run impact before editing.** Use `impact({target: "symbolName", direction: "upstream"})` or `node .gitnexus/run.cjs impact "symbolName" --direction upstream --repo .`; report callers, processes, and risk. Never substitute grep for graph analysis.
- **MUST analyze graph changes before committing.** Use `detect_changes({scope: "all"})` (MCP) or `node .gitnexus/run.cjs detect-changes --scope all --repo .` (CLI fallback). `partial: true` or `truncated: true` is not a clean check — a zero means unseen, not unaffected; re-run it. For regression review: `detect_changes({scope: "compare", base_ref: "main"})` or `node .gitnexus/run.cjs detect-changes --scope compare --base-ref "main" --repo .`.
- MUST warn on HIGH/CRITICAL `risk` pre-edit; never use `riskSharedAxes` to waive a HIGH/CRITICAL `risk` warning. Compare File/symbol: MCP File omits axes; Graph-RAG expands File.
- **MUST treat `risk: UNKNOWN` as unresolved, not as low.** An empty caller set is not evidence the symbol is unused — it can also mean the callers are not resolvable by the index (plain-object property access, dynamic dispatch, cross-language calls). `impact` pairs `UNKNOWN` with a `riskNote` saying so. Confirm with a text search before treating the symbol as safe to change or delete; do not proceed on the strength of a zero.
- **MUST use `query({search_query: "concept"})` for concepts/flows, `context({name: "symbolName"})` for a named symbol, or `impact` for blast radius, on read-only callers, dependencies, imports, or execution flow.** Graph first; text search only for empty/`UNKNOWN`/literals.
- For security review, `explain({target: "fileOrSymbol"})` lists taint findings (source→sink flows; needs `analyze --pdg`).

## Never Do

- NEVER edit a function, class, or method before MCP/CLI impact analysis.
- NEVER ignore HIGH or CRITICAL risk warnings from impact analysis, and never read `UNKNOWN` as an all-clear — it means the walk could not answer, which is the one verdict that requires confirming by other means.
- NEVER rename symbols with find-and-replace — use `rename` which understands the call graph.
- NEVER commit before MCP/CLI graph change analysis.

## Resources

| Resource | Use for |
| --- | --- |
| `gitnexus://repo/Karl/context` | Codebase overview, check index freshness |
| `gitnexus://repo/Karl/clusters` | All functional areas |
| `gitnexus://repo/Karl/processes` | All execution flows |
| `gitnexus://repo/Karl/process/{name}` | Step-by-step execution trace |

## CLI

| Task | Read this skill file |
| --- | --- |
| Understand architecture / "How does X work?" | `.claude/skills/gitnexus-exploring/SKILL.md` |
| Blast radius / "What breaks if I change X?" | `.claude/skills/gitnexus-impact-analysis/SKILL.md` |
| Trace bugs / "Why is X failing?" | `.claude/skills/gitnexus-debugging/SKILL.md` |
| Rename / extract / split / refactor | `.claude/skills/gitnexus-refactoring/SKILL.md` |
| Tools, resources, schema reference | `.claude/skills/gitnexus-guide/SKILL.md` |
| Index, status, clean, wiki CLI commands | `.claude/skills/gitnexus-cli/SKILL.md` |

<!-- gitnexus:end -->
