## graphify

This project has a knowledge graph at graphify-out/ with god nodes, community structure, and cross-file relationships.

Rules:
- For codebase questions, first run `graphify query "<question>"` when graphify-out/graph.json exists. Use `graphify path "<A>" "<B>"` for relationships and `graphify explain "<concept>"` for focused concepts. These return a scoped subgraph, usually much smaller than GRAPH_REPORT.md or raw grep output.
- If graphify-out/wiki/index.md exists, use it for broad navigation instead of raw source browsing.
- Read graphify-out/GRAPH_REPORT.md only for broad architecture review or when query/path/explain do not surface enough context.
- After modifying code, run `graphify update .` to keep the graph current (AST-only, no API cost).

### Keeping the graph fresh

**Code changes** need nothing: `.github/workflows/graphify.yml` rebuilds the graph on every push to `main`. Locally, `graphify update .` does the same (AST-only, no API key).

**Doc, SPEC, README or image changes** need a manual refresh — CI runs AST only and cannot re-extract prose or image concepts. Pick whichever fits:

| Situation | Command |
|---|---|
| You are in Claude Code | `/graphify --update` (this session is the LLM; no key needed) |
| Headless, `claude` CLI installed | `graphify extract . --backend claude-cli` (bills to your Pro/Max plan) |
| Headless, with an API key | `GEMINI_API_KEY=… graphify extract . --backend gemini` |
| Fully offline | `graphify extract . --backend ollama` |

Then commit `graphify-out/graph.json` **and** any new `graphify-out/cache/semantic/` entries.

The semantic cache is committed on purpose. It is content-addressed, so you only pay extraction for files that actually changed — editing one SPEC costs one file, not the whole corpus. Always commit the cache entries your run produces so the next person inherits the hit. A graphify upgrade that changes the extraction prompt starts a fresh cache directory and everyone re-extracts once; that is expected.
