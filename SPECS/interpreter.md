# Karl Interpreter Notes (Runtime Model)

This document defines the runtime semantics for evaluating the AST produced by the parser.
It is intended to let another agent implement the interpreter without re-deriving design intent.
It also notes the current Go implementation status where it diverges.

## Goals

- Match the behavior described in `SPECS/language.md`.
- Keep evaluation order deterministic and predictable.
- Provide a minimal, well-typed runtime model for later extensions.

## Runtime Value Model

### Primitive values

- **Int**: 64-bit signed integer.
- **Float**: 64-bit float.
- **Bool**: `true`/`false`.
- **String**: UTF-8 string.
- **Char**: single Unicode scalar (stored as string length 1).
- **Null**: absence of value.
- **Unit**: `()`; distinct from `null`.

### Composite values

- **Array**: ordered, mutable sequence.
- **Object**: string-keyed map, mutable.
- **Map**: dynamic key/value store; keys must be string, char, int, or bool.
- **Function**: closure with params + body + captured environment.
- **Range**: internal helper only; range expressions evaluate to arrays eagerly.
- **Task**: handle returned by `&` (and `|`) (waitable).
- **Rendezvous**: communication primitive with `send`/`recv` methods.
- **Set**: unordered collection of unique values (string/char/int/bool keys).

## Environment and Scope

- Lexical scoping with chained environments.
- New scope frames are created for:
  - Blocks `{ ... }`
  - Function bodies
  - Match arms
  - `for` loops (loop scope)
- `let` binds a new name in the current scope.
- Assignment updates the nearest existing binding; assigning to an unknown name is a runtime error.

## Evaluation Order

- Left-to-right for:
  - Function expressions and arguments
  - Array elements
  - Object entries
  - Infix operands
  - Match arm selection (first matching arm wins)
- Logical `&&` / `||` short-circuit.

## Core Expression Semantics

### Blocks

- A block evaluates each statement in order.
- The last expression in the block is the result.
- If there is no trailing expression, the result is `Unit`.

### Let statements

- Evaluate the RHS, then bind the pattern to the result.
- Pattern binding can introduce multiple names (destructuring).

### If

- Condition must evaluate to `Bool` (runtime error otherwise).
- Returns the chosen branch value.

### Match

- Evaluate the scrutinee.
- Try arms in order:
  - Pattern match.
  - If a guard exists, evaluate it (must be `Bool`).
  - First successful arm produces the result.
- If no arm matches: runtime error.

### For

Proposed algorithm:

1) Create a loop scope.
2) Initialize `with` bindings (once).
3) Evaluate the condition (must be `Bool`, runtime error otherwise).
4) If true: execute body, then repeat step 3.
5) If false initially, still execute the `then` block once.

**Break/continue**:
- `continue` skips to the next iteration.
- `break` exits the loop.
- `break` with no value triggers the `then` block immediately (if present) and returns its value.
- `break expr` returns `expr` as the loop result and does not evaluate `then`.
- If `then` is omitted and `break` has no value, the loop result is `Unit`.

### Assignment and mutation

- Assignment is an expression; it evaluates to the assigned value.
- Supported lvalues: identifiers, member access, index access.
- Index assignment rules:
  - Arrays require an integer index in bounds.
  - Objects/module objects require a string or char key (`obj["x-y"] = v`).
- `++`/`--` are postfix and evaluate to the updated value.

### Lambdas and calls

- Lambdas close over the lexical environment.
- Call evaluates function then arguments.
- Arity is enforced at runtime unless partial application is used.

#### Partial application with `_`

- `add(5, _)` produces a closure.
- Each `_` becomes a parameter (left-to-right).
- Non-placeholder arguments are evaluated at closure creation time and captured.

### Member access and indexing

- `obj.field` reads a property; missing property is a runtime error.
- `obj["field-name"]` reads a property by string/char key; missing property is a runtime error.
- `arr[i]` reads by index; out-of-bounds is a runtime error.
- Slice `arr[i..j]` returns a new array.

Slice semantics:
- Slices are half-open: `list[start..end]` includes `start` and excludes `end`.
- Missing `start` defaults to `0`; missing `end` defaults to `len`.
- Negative indices are allowed and count from the end (`-1` is the last element).
  - Example: `list[..-1]` returns all but the last element.
  - Example: `list[-3..]` returns the last 3 elements.
- After normalization, `start` and `end` must be within `[0, len]`; otherwise runtime error.
- If `start >= end`, the result is an empty array.

### Range

- Range expressions evaluate eagerly to arrays.
- `a..b` uses a default `step` of `1`.
- Integer/char ranges are inclusive of the end.
- Range endpoints must be integers or chars; float ranges are a runtime error.
- Char ranges produce arrays of `Char`.

### Query expressions

```
from x in source
  where ...
  orderby ...
  select ...
```

Execution:
1) Evaluate `source` (must be array).
2) Apply each `where` predicate in order.
3) If `orderby` exists, sort by key.
4) `select` runs for each element, producing output array.

Ordering requires comparable keys; otherwise runtime error.

### Import

- `import "path"` returns a zero-argument factory function.
- Calling the factory evaluates the module in a fresh environment and returns an object of its top-level `let` bindings.
- The returned object is a live view of the module environment; assignments update that instance.
- Path resolution is relative to the project root.
- Recommended implementation:
  - Resolve the path to an absolute filename.
  - Parse and cache the module program.
  - Each factory call evaluates the cached program in a new environment and returns its bindings.
- Current implementation uses the evaluator's `projectRoot`; if unset, it falls back to the process working directory.
- Dependency manager support is out of scope for now.

### exit(message)

- Terminates the entire program immediately.

### Recoverable errors (`? { ... }`)

- Only call expressions may use `? { ... }`.
- If the call succeeds, its value is returned.
- If the call fails with a recoverable error, the block runs and its value is returned.
- Inside the block, `error` is bound to `{ kind: String, message: String }`.
- Non-recoverable runtime errors still call `exit()` immediately.

Recoverable error sources:
- `decodeJson`, `readFile`, `writeFile`, `appendFile`, `deleteFile`, `exists`, `listDir`, `http`, `fail`.

Example:

```
let parsed = decodeJson(raw) ? {
    log("bad json:", error.message)
    { foo: "default" }
}
```

## Pattern Matching Semantics

- **Wildcard** `_`: always matches, binds nothing.
- **Identifier**: always matches and binds.
- **Literal**: matches by value.
- **Range pattern**: value is within range.
- **Object pattern**: matches if all specified keys exist and their values match recursively.
  Extra keys are ignored.
- **Array pattern**: matches element-wise; rest pattern captures remaining items.
- **Tuple pattern**: matches fixed length.
Call patterns are not supported.

Tuple representation:

- Tuples are represented as arrays.
- Tuple patterns `(a, b)` match arrays of length 2.
- There is no tuple literal; use arrays as the runtime carrier.

## Concurrency (Event Loop)

This section describes the target event-loop model. The current Go runtime uses goroutines
instead (see "Current Go Implementation").

### Runtime architecture (event loop)

The interpreter runs a **single-threaded event loop** that schedules cooperative tasks.
There is no preemption; tasks yield only at well-defined points (`wait`, `recv`, `sleep`).
Concurrency follows CSP-style message passing via rendezvous channels.

Core data structures (conceptual):

```
EventLoop
  tasks:    map<TaskID, Task>
  runQ:     queue<TaskID>        // ready to run
  waiting:
    onTask: map<TaskID, []TaskID>   // waiters for task completion
    onRendezvous: map<RendezvousID, []TaskID>   // waiters for recv
    onTime: min-heap<deadline, TaskID>
```

Task state machine:

```
NEW -> READY -> RUNNING -> WAITING -> READY -> ...
                          |             |
                          v             v
                       COMPLETED     CANCELED
                          |
                          v
                        FAILED
```

Each `Task` stores:

- `id`, `status`
- `env` (root lexical environment)
- `frames` (interpreter call/expr stack)
- `result` (value or error)
- `waitingOn` (task id, rendezvous id, or timer deadline)

Task execution model:

- Each task is evaluated by a **step function** that advances one or more AST frames.
- A frame captures:
  - the AST node being evaluated,
  - local environment reference,
  - partial results needed to resume (e.g., left operand already evaluated).
- When a frame reaches a yield point (`wait`, `recv`, `sleep`), the task returns control
  to the event loop and records what it is waiting on.

**Scheduling loop (simplified):**

1) Pop a task id from `runQ`.
2) Execute until:
   - it completes (stores result, wakes waiters), or
   - it hits a yield point (moves to WAITING), or
   - it fails (runtime error; marks FAILED).
3) If no ready tasks exist:
   - advance timers; move due tasks to `runQ`.
   - if still empty, the program is idle (end if all tasks are done).

### Yield points

Tasks yield only at:

- `wait task` (wait for task completion)
- `ch.send(value)` (wait for receiver)
- `ch.recv()` (wait for message)
- `sleep(ms)` (wait for timer)

All other expression evaluation is synchronous.
Tasks that never hit a yield point will monopolize the event loop.

### Spawn and wait

- `& call()` spawns a task and returns a Task handle.
- `& { call1(), call2(), ... }` spawns tasks concurrently and returns a Task handle for ordered results.
- `wait task` waits and yields the result.
- `wait` on non-Task is a runtime error (exit).

Implementation details:

- `& call()`:
  - Create a new `Task` with its own frame stack to evaluate the call.
  - Enqueue it on `runQ`.
  - Return a Task handle referencing the task id.
- `wait task`:
  - If `task` is COMPLETE, return its result immediately.
  - Otherwise, move current task to WAITING and register it under `waiting.onTask[taskID]`.

**Group tasks (`& { ... }`)**:

- Create child tasks for each call expression.
- Create a hidden "join task" that waits on all children and collects results in input order.
- The `& { ... }` expression returns a handle to the join task.

Diagram:

```
& { A(), B(), C() }
  -> spawn TaskA, TaskB, TaskC
  -> spawn JoinTask
JoinTask waits: TaskA, TaskB, TaskC
JoinTask result: [A, B, C] (preserves input order)
```

### Race

- `| { call1(), call2(), ... }` spawns tasks and returns a Task handle.
- `wait` on that handle yields the first completed result.
- Remaining tasks are intended to be cancelled; current runtime lets them continue
  (results are ignored).

Implementation details (target behavior):

- Create child tasks for each call expression.
- Create a hidden "race task" that waits for the first child to complete, returns that value,
  and signals cancellation for the remaining children.

Diagram:

```
| { A(), B(), C() }
  -> spawn TaskA, TaskB, TaskC
  -> spawn RaceTask
RaceTask returns first result (A or B or C)
RaceTask cancels remaining tasks
```

Cancellation semantics (cooperative, target behavior):

- Cancellation is **requested** immediately for losing tasks.
- A task only stops when it reaches the next yield point (`wait`, `recv`, `sleep`).
- If a task never yields, it will run to completion but its result is discarded.

Example:

```
let busy = () -> {
    for true {  // no wait/recv/sleep inside
        work()
    }
}

let slow = () -> {
    sleep(1000)
    "slow"
}

let fastest = wait | { busy(), slow() }
// fastest is "slow"
// busy() keeps running until it reaches a yield point (it never does),
// so it continues but its result is ignored.
```

### Rendezvous

- `rendezvous()` returns a Rendezvous channel for task communication.
- `ch.send(value)` enqueues a value and returns Unit.
- `ch.recv()` returns `[value, done]`; `done` is true when the rendezvous is closed (value is `null`).
- `ch.done()` closes the rendezvous (no further sends allowed).

Implementation details (current runtime):

- Rendezvous stores:
  - `ch` (unbuffered channel)
  - `closed` flag
- `send`:
  - If closed: runtime error (exit).
  - Else: block until a receiver accepts the value.
- `recv`:
  - Block until a value arrives or the rendezvous is closed.
  - If closed and empty: return `[null, true]`.


### Timers

`sleep(ms)` is a built-in function that yields:

- It registers the current task in `waiting.onTime` with a deadline.
- When the deadline is reached, the task is moved back to `runQ`.

## Current Go Implementation (Status)

- Implemented in `interpreter/` with a recursive evaluator over the AST.
- Tasks are backed by Go goroutines (not a custom event loop yet).
- `wait` blocks the goroutine; `sleep` uses `time.Sleep`; `send`/`recv` block on a Go channel.
- Race tasks return the first result but do not cancel losers yet (results are ignored).
- Runtime errors call `exit()` immediately, including inside spawned tasks.
- Recoverable errors are only produced by specific builtins and can be handled with `? { ... }`.

## Built-in Functions (Assumed)

- `rendezvous()` -> Rendezvous
- `sleep(ms)` -> Unit (yields)
- `now()` -> Int (epoch ms)
- `exit(message)` -> no return (terminates)
- `fail(message)` -> no return (recoverable error)
- `log(...values)` -> Unit
- `str(value)` -> String
- `parseInt(string)` -> Int
- `encodeJson(value)` -> String
- `decodeJson(text)` -> Value
- `readFile(path)` -> String
- `writeFile(path, data)` -> Unit
- `appendFile(path, data)` -> Unit
- `deleteFile(path)` -> Unit
- `exists(path)` -> Bool
- `listDir(path)` -> Array<String>
- `http({ method, url, headers, body, })` -> { status, headers, body, }
- `done(rendezvous)` -> Unit (closes rendezvous)
- `map()` -> Map
- `set()` -> Set
- `map(list, fn)` remains the array map function.
- `sort(list, cmp)` -> Array (returns new array)
- `split(string, sep)` -> Array
- `chars(string)` -> Array
- `trim(string)` -> String
- `toLower(string)` -> String
- `toUpper(string)` -> String
- `contains(string, substr)` -> Bool
- `startsWith(string, prefix)` -> Bool
- `endsWith(string, suffix)` -> Bool
- `replace(string, old, new)` -> String
- `get(map, key)` -> value or `null`
- `set(map, key, value)` -> Map
- `add(set, value)` -> Set
- `has(map, key)` -> Bool
- `has(set, value)` -> Bool
- `delete(map, key)` -> Bool
- `delete(set, value)` -> Bool
- `keys(map)` -> Array
- `values(map)` -> Array
- `values(set)` -> Array
- `abs(number)` -> Number
- `sqrt(number)` -> Float
- `pow(base, exp)` -> Float
- `sin(number)` -> Float
- `cos(number)` -> Float
- `tan(number)` -> Float
- `floor(number)` -> Int
- `ceil(number)` -> Int
- `min(a, b)` -> Number
- `max(a, b)` -> Number
- `clamp(value, min, max)` -> Number
- `rand()` -> Int
- `randInt(min, max)` -> Int (inclusive)
- `randFloat(min, max)` -> Float

Notes:
- `http` accepts `headers` as an object or map; the response `headers` is a `Map` so header names like `Content-Type` are accessible via `headers.get("Content-Type")`.

## Equality Semantics

- `==` is strict identity for composite values (arrays/objects/maps/sets/tasks/rendezvous/functions).
- `eqv` is deep for arrays/objects/maps; other composite types use identity.
- For primitive types, `==` and `eqv` are the same.
- Mixed-type comparisons return `false` (no implicit coercion).

## Built-in Methods and Properties (Assumed)

Arrays:
- `length` (property)
- `map`, `filter`, `reduce`, `sum`, `find`, `sort`

Strings:
- `length` (property)
- `length` counts Unicode scalar values (runes), not bytes.
- `split`, `chars`, `trim`, `toLower`, `toUpper`, `contains`, `startsWith`, `endsWith`, `replace`

Maps:
- `get`, `set`, `has`, `delete`, `keys`, `values`

Sets:
- `size` (property)
- `add`, `has`, `delete`, `values`

Rendezvous:
- `send`, `recv`, `done`

Tasks:
- `wait` handled by language keyword
- `then(fn)` runs `fn` when the task completes and returns a new Task handle.

Member call semantics (no implicit receiver):

- `obj.field` is a plain property lookup.
- Calling a property does **not** pass `obj` implicitly.
- Built-in collection methods use dot-call **syntax sugar** only:
  - `list.map(fn)` desugars to `map(list, fn)`
  - `list.filter(fn)` desugars to `filter(list, fn)`
  - `list.reduce(fn, init)` desugars to `reduce(list, fn, init)`
  - `list.sort(fn)` desugars to `sort(list, fn)`
  - `text.split(sep)` desugars to `split(text, sep)` (same for other string helpers)
  - `m.get(key)` desugars to `get(m, key)` (same for `set`/`has`/`delete`/`keys`/`values`)
  - `s.add(value)` desugars to `add(s, value)` (same for `has`/`delete`/`values`)

User-defined objects do not have method receivers; functions must take the object explicitly.

## CLI Usage

The CLI can evaluate Karl source or print its AST:

- `karl parse <file.k> [--format=pretty|json]`
- `karl run <file.k>`
- `cat <file.k> | karl run -`

## Known Limitations / Notes

- The interpreter is a direct AST evaluator (no bytecode/JIT or optimization passes).
- Tasks are backed by goroutines; task scheduling order is nondeterministic.
- Range expressions are eager and allocate full arrays.
- No tail-call optimization; deep recursion can overflow the Go stack.
