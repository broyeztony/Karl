# API Review Checklist

Use this checklist before implementing or merging new stdlib APIs.

## 1) Problem Fit

- Is this a common, reusable problem?
- Are users currently writing this algorithm manually?
- Will moving it into stdlib reduce LoC and errors materially?

## 2) Surface Quality

- Is naming short and explicit?
- Are signatures uniform and easy to discover?
- Are return shapes minimal and stable?
- Is there any avoidable aliasing?

## 3) Semantic Precision

- Is behavior deterministic?
- Are tie-break rules explicit?
- Are ordering guarantees explicit?
- Are null vs runtime error cases explicit?

## 4) Performance Contract

- Are complexity targets stated?
- Does implementation meet those targets?
- Is there benchmark coverage for performance-critical paths?

## 5) AI Ergonomics

- Can an agent solve the task by composing built-ins only?
- Does the API remove algorithmic boilerplate from user code?
- Do docs include realistic, concise examples?

## 6) Release Readiness

- Spec updated
- Runtime updated
- Tests added (happy path + edge cases)
- Examples added
- No behavior ambiguity remains
