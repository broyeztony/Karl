// ============================================
// EXPRESSION-BASED LANGUAGE DESIGN - KARL
// ============================================

This is a spec for the Karl expression-based programming language.

## Philosophy

The language is **expression-based**: Everything evaluates to a value.

### Core Design Principles

1. **`->` has exactly one meaning**: mapping a binder (parameters or pattern) to an expression
   - Lambdas: `(x, y) -> x + y`
   - Case arms: `case Pattern -> expression`

2. **One pattern-matching construct**: `match` handles all branching

3. **Explicit failure**: `exit()` terminates; recoverable errors use `? { ... }` on fallible calls

4. **Everything is an expression**: blocks, loops, conditionals all evaluate to values

### Influences

- **JavaScript**: rapid prototyping, JSON-like data structures, no type ceremony
- **Scala**: elegant syntax, pattern matching, expression-oriented
- **Go**: concurrency as a language primitive (adapted for event-loop model)

// ============================================
// CANONICAL EXAMPLE
// ============================================

// This example demonstrates the "return to the left" philosophy:
// complex operations flow into a single binding

let finalValue =
    match compute(10, 101) {
        case 0 -> { ite: 0, acc: [] }
        case result -> {
            for x < result with x = 0, acc = [] {
                x++;
                acc += [x];
            } then { ite: x, acc }
        }
    }

// `finalValue` is the result of the match expression
// Either the loop result or the default branch value

// ============================================
// 1. CONTROL FLOW AS EXPRESSIONS
// ============================================

// if/else expressions
let value = if x < 10 { 1 } else { 0 }
// Conditions must evaluate to Bool; otherwise runtime error (exit).

// Multi-way if (else-if chains)
let grade = if score >= 90 { "A" }
            else if score >= 80 { "B" }
            else if score >= 70 { "C" }
            else { "F" }

// match expressions (the unified pattern-matching construct)
let value = match x {
    case 0      -> "zero"
    case 1..4   -> "small"
    case _      -> "other"
}

// Pattern matching with destructuring
let result = match value {
    case { type: "success", data } -> data
    case { type: "error", message } -> "Error: " + message
    case null -> "No value"
}

// Guarded patterns
let result = match value {
    case x if x > 0 -> "positive"
    case x if x < 0 -> "negative"
    case 0 -> "zero"
}
// Guards must evaluate to Bool; otherwise runtime error (exit).

// ============================================
// 2. LOOP EXPRESSIONS
// ============================================

// for loops as expressions
// - Variables in `with` clause are initialized once (optional)
// - Condition uses the same variables
// - Variables are updated explicitly in the loop body
// - `then` block determines the loop's return value when the loop ends normally
//   or when `break` has no value
// - If omitted, the loop returns Unit unless a `break expr` is executed
// - Condition must evaluate to Bool; otherwise runtime error (exit)
// - If condition is false initially, `then` block still executes once with initial values
// - If `then` is a single expression, its braces may be omitted
// - `break` (no value) triggers the `then` block immediately
// - `break expr` exits the loop and returns expr (bypasses `then`)
// - `continue` skips to the next iteration

let value = for x < 10 with x = 0, acc = [] {
    x++;
    acc += [x];
} then {
    { items: acc, count: x }
}

// Simpler example: sum 1 to 10
let sum = for i <= 10 with i = 1, total = 0 {
    total += i;
    i++;
} then total

// Multiple statements in then block
let result = for n > 0 with n = 100, steps = 0 {
    n = n / 2;
    steps++;
} then {
    let msg = "took " + steps + " steps";
    { steps, msg, }
}

// Breaking early (rendezvous example)
let found = for true with msg = null {
    let [value, done] = ch.recv()
    if done { break }
    msg = value
    match msg {
        case { kind: "done" } -> break
        case _ -> continue
    }
} then msg

// NOTE: for-in loops (iterators) - design TBD

// ============================================
// 3. ERROR HANDLING
// ============================================

// Model
// - There is no Result/Ok/Error type in the language.
// - Unrecoverable failures call exit(message) immediately.
// - Recoverable failures are only allowed from specific builtin calls (see below).
// - Use explicit checks for optional data (null or sentinel values).
// - Missing property access or out-of-bounds index access are runtime errors and call exit(...).
// - `wait` on a non-Task is a runtime error and calls exit(...).

// Recoverable errors: postfix catch block
// - Only call expressions may use `? { ... }`.
// - The call result is returned on success.
// - If the call fails with a recoverable error, the block runs and its value is returned.
// - Inside the block, `error` is implicitly bound to:
//   { kind: String, message: String }.
// - Non-recoverable errors still call exit(), even if wrapped in `? {}`.

// Builtins that can produce recoverable errors:
// decodeJson, readFile, writeFile, appendFile, deleteFile, exists, listDir, http, fail

// Example: recover from bad JSON
let raw = "{\"foo\":\"bar\"}"
let parsed = decodeJson(raw) ? {
    log("bad json:", error.message)
    { foo: "default" }
}

// Example: user-defined recoverable failure via fail()
let divide = (a, b) -> {
    if b == 0 { fail("division by zero") }
    a / b
}
let quotient = divide(10, 0) ? { 0 }

// Example: nested recoverable errors
let config = decodeJson(readFile("config.json")) ? {
    log("config error:", error.message)
    decodeJson("{\"mode\":\"safe\"}") ? { mode: "safe", }
}

// Explicit checks before calling fallible operations
let parseAge = (s) -> {
    if s == "" { exit("empty age") }
    parseInt(s)
}

let user = fetchUser(id)
if user == null { exit("user not found") }

// ============================================
// 4. FUNCTIONAL EXPRESSIONS
// ============================================

// Functions are first-class values: they can be assigned, passed, and returned.

// Lambda/anonymous functions
// `->` means: binder -> expression
let add = (x, y) -> x + y
let square = x -> x * x
let noArgs = () -> 42

// Higher-order functions
let result = [1, 2, 3].map(x -> x * 2)
let sum = [1, 2, 3].reduce((acc, x) -> acc + x, 0)
let filtered = [1, 2, 3, 4].filter(x -> x % 2 == 0)

// Partial application
let add5 = add(5, _)
let result = add5(10)  // 15

// Function composition
let compose = (f, g) -> x -> f(g(x))
let f = compose(square, add(1))
let result = f(5)  // (5 + 1)^2 = 36

// ============================================
// 5. CHAINING EXPRESSIONS
// ============================================

// Method chaining
let result = [1, 2, 3, 4, 5]
    .filter(x -> x % 2 == 0)
    .map(x -> x * x)
    .sort((a, b) -> b - a)
    .sum()

// Note: dot-call chaining is syntax sugar for built-in collection functions.
// User-defined objects have no implicit receiver; obj.method(arg) does not pass obj automatically.

// ============================================
// 6. BLOCK EXPRESSIONS
// ============================================

// Blocks return last expression
let value = {
    let a = 10;
    let b = 20;
    a + b  // returns 30
}

// Blocks can contain any statements/expressions
// If a block has no trailing expression, it evaluates to Unit
// Semicolons are optional statement separators inside blocks
// Unit literal is ()
let computed = {
    let temp = expensiveCalculation();
    let adjusted = temp * coefficient;
    if adjusted > threshold { adjusted } else { threshold }
}

// ============================================
// 6.5 ASSIGNMENT & MUTATION EXPRESSIONS
// ============================================

// Assignment is an expression and evaluates to the assigned value
let x = 1;
let y = x = 2;  // y is 2, x is 2

// Mutation operators are expressions and evaluate to the updated value
let n = 0;
let m = n++;  // m is 1, n is 1

// Assignment targets are identifiers, member access, or index access
user.name = "Ada"
items[0] += 1
user["display-name"] = "Ada L."

// Index access rules:
// - Arrays require integer indices.
// - Objects accept string/char keys (useful for non-identifier external fields).

// Equality:
// - `==` is strict identity for composite values (arrays/objects/maps/sets/tasks/rendezvous).
// - `eqv` is structural equivalence for arrays/objects/maps.
// - Other composite types use identity for both `==` and `eqv`.

// ============================================
// 6.6 LOOP CONTROL EXPRESSIONS
// ============================================

// `break` exits the nearest loop; `continue` skips to the next iteration
// `break` may optionally carry a value used as the loop's final result
let total = for i <= 10 with i = 1, acc = 0 {
    if i == 5 { break acc }
    acc += i
    i++
} then acc

// ============================================
// 7. PATTERN MATCHING EXPRESSIONS
// ============================================

// Destructuring assignment
let { x, y } = point
let [first, ...rest] = list
let (a, b) = tuple
// Trailing commas are allowed in object/array/tuple patterns.
let { x, y, } = point

// Structural patterns in match
let value = match data {
    case [x, y, z]  -> x + y + z
    case { x, y }   -> x * y
    case x          -> x
}

// Tuple patterns match arrays of the same length (tuples are arrays at runtime).
// There is no tuple literal; use arrays as the carrier value.
let pair = [lat, lon]
let (latValue, lonValue) = pair

// Nested patterns
let result = match response {
    case { status: 200, body: { data } } -> data
    case { status: 404 }                  -> exit("Not found")
    case { status }                       -> exit("Status: " + status)
}

// ============================================
// 8. OBJECT & STRUCT EXPRESSIONS
// ============================================

// Object literals
let person = {
    name: "Alice",
    age: 30,
    greet: () -> "Hello, " + name
}
// Object keys must be identifiers; use Map for dynamic/string keys.

// Struct initialization (typed objects)
// - Syntax sugar; the type name is not enforced at runtime.
let point = Point { x: 10, y: 20 }

// Object spread
let updated = { ...person, age: 31 }

// Shorthand property names
let x = 10;
let y = 20;
let point = { x, y, }  // same as { x: x, y: y }
// Note: shorthand-only objects require a trailing comma to disambiguate from blocks.

// ============================================
// 8.5 STRING EXPRESSIONS
// ============================================

// String and char literals support escapes:
// \\ \" \' \n \r \t \b \f \uXXXX (4 hex digits)

let text = " Hello Karl "
let trimmed = text.trim()
let words = trimmed.split(" ")
let chars = trimmed.chars()
let hasHello = trimmed.contains("Hello")
let starts = trimmed.startsWith("Hello")
let ends = trimmed.endsWith("key")
let lower = trimmed.toLower()
let upper = trimmed.toUpper()
let replaced = trimmed.replace("Karl", "World")
let length = trimmed.length
// length counts Unicode scalar values (not bytes)

// ============================================
// 9. MAP EXPRESSIONS
// ============================================

// Maps are for dynamic keys (objects are fixed, string-keyed shapes).
let counts = map()
counts.set("apples", 2)
counts.set("bananas", 1)

let hasApples = counts.has("apples")     // true
let missing = counts.get("cherries")     // null
let removed = counts.delete("bananas")   // true
let keys = counts.keys()                 // array of keys (order not guaranteed)
let values = counts.values()             // array of values (order not guaranteed)

// Map key types: string, char, int, bool.

// ============================================
// 9.5 SET EXPRESSIONS
// ============================================

// Sets store unique values (string, char, int, bool).
let uniq = set()
uniq.add(1)
uniq.add(2)
uniq.add(2)

let hasOne = uniq.has(1)                 // true
let removed = uniq.delete(1)             // true
let values = uniq.values()               // array of values (order not guaranteed)
let size = uniq.size                     // integer

// Note: set() constructs a Set. set(map, key, value) still sets a Map entry.

// ============================================
// 10. RANGE & SLICE EXPRESSIONS
// ============================================

// Range expressions
let numbers = 1..10
let chars = 'a'..'z'
let evens = 2..100 step 2

// Range evaluation:
// - Ranges evaluate eagerly to arrays.
// - Integer/char ranges are inclusive of the end.
// - Range endpoints must be integers or chars; float ranges are not allowed.

// Range endpoints must be explicit in standalone ranges.
// Open-ended ranges are allowed only inside slice expressions.

// Slice expressions
let sub = list[1..5]
let reversed = list[..-1]
let tail = list[1..]
let head = list[..5]
let prefix = "some string"[0..4]   // "some"

// Object key indexing
let contentType = headers["Content-Type"]

// Slice semantics:
// - Half-open bounds: list[start..end] includes start, excludes end.
// - Missing start defaults to 0; missing end defaults to length.
// - Negative indices count from the end (e.g., list[..-1] drops the last element).
// - Slicing works for arrays and strings.
// - String slicing is rune-based (Unicode-safe), not byte-based.

// ============================================
// 11. RECURSIVE EXPRESSIONS
// ============================================

// Recursive functions
let factorial = (n) -> if n <= 1 { 1 } else { n * factorial(n - 1) }

let fib = (n) -> match n {
    case 0 -> 0
    case 1 -> 1
    case n -> fib(n - 1) + fib(n - 2)
}

// ============================================
// 12. CONCURRENCY EXPRESSIONS
// ============================================

// Spawn expressions - all tasks run concurrently
// & { ... } returns a Task handle for an array of results
let tasks = & {
    task1(),
    task2(),
    task3()
}
let results = wait tasks

// Race expression - first to complete wins
let fastest = wait | {
    fetchFromServer1(),
    fetchFromServer2()
}

// --------------------------------------------
// Concurrency model (event-loop based)
// --------------------------------------------
// - Concurrency is cooperative and scheduled by a single event loop.
// - The model is CSP-style: tasks communicate via rendezvous channels (created with rendezvous() or channel()).
// - & call spawns a task and returns a Task handle.
// - & { call1(), call2(), ... } spawns all calls concurrently and returns a Task handle of results in order.
// - wait task waits for completion and yields the task result.
// - | { call1(), call2(), ... } returns a Task handle for the first completed result.
// - wait on that handle yields the first result; losing tasks are cancelled automatically
//   (cooperative cancellation).
// - task.cancel() requests cancellation for a task.
// - Tasks communicate by returning values, shared immutable data, or rendezvous channels.
// - If you need to know which task completed, return a tagged value from each task.

// Passing data by returning values
let results = wait & {
    fetchUser(1),
    fetchUser(2)
}
let merged = { a: results[0], b: results[1] }

// Capturing shared immutable data
let baseUrl = "https://api"
let pages = wait & {
    fetch(baseUrl + "/p1"),
    fetch(baseUrl + "/p2")
}

// Rendezvous channels (synchronous; send/recv block)
// channel() is an alias of rendezvous().
let ch = channel()

wait & { producer(ch), consumer(ch) }

let producer = (ch) -> {
    ch.send({ kind: "data", value: 1 })
    ch.send({ kind: "data", value: 2 })
    ch.done()
}

let consumer = (ch) -> {
    for true with res = ch.recv(), acc = [] {
        let [msg, done] = res
        if done { break acc }
        match msg {
            case { kind: "data", value } -> { acc += [value] }
        }
        res = ch.recv()
    } then acc
}

// Streaming consumer: process messages as they arrive (infinite stream)
let streamConsumer = (ch) -> {
    for true with res = ch.recv() {
        let [msg, done] = res
        if done { break () }
        match msg {
            case { kind: "data", value } -> { process(value) }
            case { kind: "tick" }        -> { heartbeat() }
        }
        res = ch.recv()
    } then ()
}

// Channel (rendezvous) API
// - ch.send(value) blocks until a receiver accepts the value (returns Unit)
// - ch.recv() blocks until a value arrives or the rendezvous is closed; returns [value, done]
// - ch.done() closes the rendezvous
// - sending on a closed rendezvous is a runtime error and calls exit(...)
// - Task values are waitable handles returned by & and |
// - Task handles support then():
//   let t = & doWork()
//   t.then(res -> log("done", res))

// ============================================
// 13. QUERY EXPRESSIONS (SQL-like)
// ============================================

// Query expressions for collection manipulation
let results = from user in users
    where user.age > 18
    select { name: user.name, age: user.age }

// With ordering
let sorted = from user in users
    where user.active
    orderby user.name
    select user

// ============================================
// 14. IMPORTS
// ============================================

// Import returns a factory; calling it returns module bindings.
// Imports are expressions and can be used anywhere.
let makeUtil = import "examples/features/import_module.k"
let util = makeUtil()
log(util.greet("karl"))

// Import rules:
// - The path is a string literal.
//   - If it starts with `./` or `../`, it is resolved relative to the importing file's directory.
//   - Otherwise, it is resolved from the project root (currently: the process working directory).
// - The factory returns an object containing all top-level let bindings.
// - There is no export keyword yet; all top-level lets are exported.
// - The returned object is live: assigning to its properties updates the module instance.
// - Name collisions are the caller's responsibility (avoid duplicate paths).
// - Dependency manager support is out of scope for now.

// ============================================
// MINIMAL GRAMMAR (EBNF)
// ============================================

// Note: This is a minimal grammar sketch for the whole language.
// It focuses on core constructs and intended precedence, not full lexical detail.

program         = { statement } ;

statement       = let_stmt | expr_stmt ;
let_stmt        = "let" pattern "=" expr [ ";" ] ;
expr_stmt       = expr [ ";" ] ;

expr            = if_expr
                | match_expr
                | for_expr
                | lambda_expr
                | loop_ctrl
                | assign ;

if_expr         = "if" expr block [ "else" ( if_expr | block | expr ) ] ;

match_expr      = "match" expr "{" { match_arm } "}" ;
match_arm       = "case" pattern [ "if" expr ] "->" expr ;

for_expr        = "for" expr [ "with" for_bindings ] block [ "then" then_block ] ;
for_bindings    = binding { "," binding } ;
binding         = pattern "=" expr ;
then_block      = block | expr ; // braces optional for single expression

lambda_expr     = binder "->" expr ;
binder          = pattern | "(" [ pattern { "," pattern } ] ")" ;

loop_ctrl       = "break" [ expr ]
                | "continue" ;

assign          = logic_or
                | lvalue assign_op expr ;
lvalue          = IDENT { ( "." IDENT | "[" expr "]" ) } ;
assign_op       = "=" | "+=" | "-=" | "*=" | "/=" | "%=" ;

logic_or        = logic_and { "||" logic_and } ;
logic_and       = equality { "&&" equality } ;
equality        = comparison { ( "==" | "!=" | "eqv" ) comparison } ;
comparison      = range { ( "<" | "<=" | ">" | ">=" ) range } ;
range           = add [ ".." add [ "step" add ] ] ;
add             = mul { ( "+" | "-" ) mul } ;
mul             = unary { ( "*" | "/" | "%" ) unary } ;
unary           = ( "!" | "-" ) unary
                | wait_expr
                | import_expr
                | spawn_expr
                | postfix ;

wait_expr      = "wait" unary ;
import_expr    = "import" STRING ;
spawn_expr      = "&" spawn_target ;
spawn_target    = call_expr
                | "{" [ call_expr { "," call_expr } ] "}" ;

postfix         = recover_expr ;
recover_expr    = call_expr [ "?" recovery_block ] ;
recovery_block  = block ; // brace expression; object literal allowed by disambiguation
call_expr       = primary { call | member | index | inc_dec } ;
call            = "(" [ expr { "," expr } [ "," ] ] ")" ;
member          = "." ( IDENT | "then" ) ;
index           = "[" expr "]" ;
inc_dec         = "++" | "--" ;

primary         = literal
                | IDENT
                | "_" // placeholder for partial application
                | "(" expr ")"
                | block
                | object
                | array
                | query_expr
                | race_expr
                | struct_init ;

block           = "{" { statement } [ expr ] "}" ;

object          = "{" [ object_entry { "," object_entry } [ "," ] ] "}" ;
object_entry    = IDENT [ ":" expr ] | "..." expr ;
array           = "[" [ expr { "," expr } [ "," ] ] "]" ;

race_expr       = "|" "{" [ call_expr { "," call_expr } ] "}" ;

struct_init     = IDENT object ;

query_expr      = "from" IDENT "in" expr
                  { "where" expr }
                  [ "orderby" expr ]
                  "select" expr ;

pattern         = "_" | literal | IDENT | range_pattern
                | "{" [ pattern_entry { "," pattern_entry } [ "," ] ] "}"
                | "[" [ pattern { "," pattern } ] [ "," "..." pattern ] [ "," ] "]"
                | "(" [ pattern { "," pattern } ] [ "," ] ")" ;
pattern_entry   = IDENT [ ":" pattern ] ;

range_pattern   = literal ".." literal ;
literal         = NUMBER | STRING | CHAR | "true" | "false" | "null" | "()" ;

// ============================================
// DISAMBIGUATION RULES
// ============================================

// 1. Block vs object literal
// - "{" starts a block by default.
// - A "{" is parsed as an object literal when:
//   - it is empty ("{}"),
//   - it contains any top-level ":" (key/value entry),
//   - it contains any top-level "..." spread, or
//   - it contains a trailing comma before "}" at top level (e.g., "{ x, }", "{ x, y, }").
// - Shorthand-only objects must use a trailing comma to disambiguate:
//   "{ name, }" forces object literal parsing.
// - Struct init uses IDENT object (e.g., "Point { x: 1 }") and follows the same object rules.

// 2. Statement separation
// - Semicolons are optional.
// - Newlines are whitespace and do not terminate statements.
// - Newlines are ignored inside (), [], and {}.


// ============================================
// FUTURE WORK (NOT IMPLEMENTED)
// ============================================

// 1. For-in loops: iterating collections (syntax TBD)
// 2. Type annotations: optional types for bindings and function signatures
