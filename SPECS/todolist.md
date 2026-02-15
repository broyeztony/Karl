# Karl TODO

## Current priorities

- handle ENV vars
- read from command-line (input / scanner)
- add a httpServer built-in
- build a debugger (breakpoints, step in/over/out, stack/locals inspection; CLI first, DAP later)
- build a notebook system for Karl like Jupyter using the repl server
- Keep tests green as syntax/runtime changes land (`gotest`).
- Parser: consider treating newlines as statement boundaries to reduce adjacency ambiguity.
- Extend test coverage when new syntax is added (parser + interpreter + examples).
- Brainstorm objects versus maps versus mutability versus shapes
- Recover block that run for any situation where the runtime throws an expection? ✅
- string interpolation
- make a <task> cancelable ✅


## Known review points

- Disambiguation rules (block vs object) must stay consistent with examples and tests ✅
- Import/factory behavior and live exports should remain explicit in specs.
