---
name: karl-api-first-stdlib
description: Design and review Karl standard library APIs for collections, streams, processes, and runtime primitives when users ask for new capabilities or ergonomics upgrades. Use this skill when an implementation currently requires user-land algorithms that should be promoted into optimized built-ins with deterministic semantics, explicit complexity contracts, and AI-friendly method shapes.
---

# Karl API-First Stdlib

## Overview

Design Karl APIs so users compose high-level primitives instead of re-implementing known algorithms in app code.
Prioritize runtime-level optimized methods with explicit behavior and complexity.

## Apply This Direction

1. Identify the repeated algorithm in user code.
2. Promote that algorithm to a built-in or method if it is a known, reusable problem.
3. Specify deterministic semantics before implementing.
4. Specify complexity targets (`O(1)`, `O(log n)`, `O(n)`) in docs/specs.
5. Keep names short, explicit, and discoverable by AI agents.
6. Add examples that solve real tasks with minimal user code.
7. Add tests for correctness, edge cases, determinism, and error behavior.
8. Benchmark when the API claims optimization.

## Non-Negotiable Rules

- Prefer API-level solutions over user-land algorithm boilerplate.
- Avoid forcing users to write textbook algorithms for common tasks.
- Make tie-breaking and ordering deterministic.
- Define null/error behavior explicitly.
- Keep return shapes stable and small.
- Avoid aliases unless explicitly requested.
- Do not ship speculative abstractions without concrete use cases.

## API Review Checklist

Read `/Users/tonybroyez/Documents/karl/skills/karl-api-first-stdlib/references/api-review-checklist.md` before finalizing a new API.

## Collection-Specific Guidance

For ordered trees:

- Provide search methods users actually need (`closest`, `floor`, `ceil`, `predecessor`, `successor`, `range`).
- Keep these operations runtime-optimized (`O(log n)` targets where applicable).
- Do not require users to extract `keys()` and implement binary search in app code.

For hierarchical trees:

- Use a dedicated node-tree API for semantic relationships (`parent`, `children`, `siblings`, `ancestors`, `descendants`).
- Do not expose AVL/Treap internal structure as stable parent/child semantics.

## Output Contract for API Proposals

When proposing a new Karl stdlib API, always include:

1. Signature(s)
2. Return shape(s)
3. Complexity targets
4. Determinism/tie rules
5. Error/null behavior
6. One realistic example
7. Minimal test matrix
