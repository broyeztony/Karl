# Karl Tooling Naming

This document defines a coherent naming system for Karl tooling.

Principles:
- Concrete metaphors over abstract jargon
- Short CLI-friendly names
- Consistent tone across the toolchain
- Serious engineering voice (minimal cleverness)

## 1) Core build and package workflow

### Package manager: `lager`

Meaning: storage, repository, cellar.

Examples:
- `karl lager add foo`
- `karl lager update`

### Build system: `forge`

Examples:
- `karl forge`
- `karl forge --release`

## 2) Language internals (advanced users)

### Intermediate representation: `skein`

Usage:
- "Karl lowers to Skein before optimization."

### Optimizer: `temper`

Usage:
- "Skein is optimized by Temper."

### Runtime: `loom`

Usage:
- "Karl programs execute on Loom."

Rationale:
- Strong execution/scheduling metaphor
- Natural fit for concurrency narratives

## 3) Developer experience tools

### Formatter: `plain`

Positioning:
- One clear style
- Minimal formatting bikeshedding

CLI:
- `karl plain`

### Linter: `stern`

Positioning:
- Strict, serious, fair diagnostics

Usage:
- "Stern warnings are enabled by default."

### Documentation generator: `folio`

Positioning:
- Structured technical pages and references

## 4) Testing and verification

### Test runner: `probe`

Positioning:
- Precise, signal-focused test execution

### Fuzz/property testing: `strain`

Positioning:
- Pushes assumptions and edge behavior

## 5) Debugging and observability

### Debugger: `trace`

Positioning:
- Explicit and familiar naming

### Profiler: `weight`

Positioning:
- Highlights computational heaviness and cost

## 6) Distribution and release

### Artifact bundler: `crate`

Positioning:
- Physical, stable artifact metaphor

### Release tool: `seal`

Positioning:
- Final publication step
- Implies integrity and intentional release

## CLI naming map

- `karl lager`   -> package manager
- `karl forge`   -> build system
- `karl plain`   -> formatter
- `karl stern`   -> linter
- `karl folio`   -> docs generator
- `karl probe`   -> test runner
- `karl strain`  -> fuzz/property testing
- `karl trace`   -> debugger
- `karl weight`  -> profiler
- `karl crate`   -> bundler
- `karl seal`    -> release tool

## Notes

- These names are a product vocabulary spec, not a mandatory implementation plan.
- Internal tool names (`skein`, `temper`, `loom`) can exist in docs and logs before CLI exposure.
