# Recoverable Errors: Broad vs Narrow

## User Question
The error/recovery operator (a.k.a `?`) allows to define a block that is executed (evaluated) when a recoverable error occurs.

Is it:
1. advisable
2. desirable
3. feasible
to allow the ? operator for any situation where the runtime throws an error, be it caused by either executing builts-in / internals of the language or user-defined logic (which requires `fail(...)`)

As an example:
```
let m = map()
let vv = m.fhjdsksd // fhjdsksd is not a member of map
```
throws
```
runtime error: unknown map member: fhjdsksd
  at .playground/demo3.k:27:12
  27 | let vv = m.fhjdsksd
    |    
```
currently.

what if we could do:
```
let vv = m.fhjdsksd ? { "foo-bar" }
```

## Assistant Answer
Given the current specs and implementation, the answer is mostly "no" for the broad version, but "yes" for a narrower, explicit variant.

Right now:
- `? { ... }` is only valid after a call expression (`recover_expr = call_expr ["?" recovery_block]`). `SPECS/language.md` and `SPECS/interpreter.md`.
- Only specific builtins and `fail(...)` can produce *recoverable* errors; missing member/index is explicitly a *runtime error* that calls `exit(...)`. `SPECS/language.md`.
- The evaluator only catches `RecoverableError` in `evalRecoverExpression`; member access returns `RuntimeError`. `interpreter/eval.go`.

1) Advisable?
Generally no for "any runtime error." It blurs the boundary between expected failure and programmer bugs, and makes typos (like unknown members) silently recoverable. That makes debugging harder and can mask real defects, especially in a dynamic language without a type system. It also risks continuing after partial side effects (e.g., failures inside mixed mutation + computation), which can leave state inconsistent.

2) Desirable?
Only in narrow, explicit cases. Your example is a *missing method* on a map (`m.fhjdsksd`), which is likely a bug-not a recoverable condition. For dynamic keys, the intended pattern is `map.get(...)`/`map.has(...)` plus an explicit fallback:
```
let vv = if m.has("fhjdsksd") { m.get("fhjdsksd") } else { "foo-bar" }
```
If you want convenient fallbacks for missing properties or out-of-bounds access, it's cleaner to add explicit safe access or "get-or" helpers rather than making *all* runtime errors recoverable.

3) Feasible?
Yes, but it's a spec + parser + evaluator change:
- Grammar/parser: allow `? { ... }` after more than calls (e.g., any postfix expression or any expression) by relaxing `parseRecoverExpression` and `recover_expr` in `SPECS/language.md`.
- Evaluator: decide which runtime errors become recoverable and convert them to a recoverable form; or let `evalRecoverExpression` catch `RuntimeError` and construct an `error` object with a `kind` (you'd need to define `kind` for runtime errors).
- Spec updates: clarify which runtime errors are still fatal vs recoverable; keep `exit()` non-recoverable.

If you want a middle ground, a good option is:
- keep `?` for call expressions,
- add explicit safe accessors (`obj.get(...)`, `arr.get(...)`, `obj.has(...)`),
- or introduce a dedicated "safe" operator for member/index access only.

If you want, I can draft a concrete proposal (spec changes + parser/evaluator diffs) for one of these narrower designs.
