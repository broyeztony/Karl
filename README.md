
<img src="assets/karl.png">

### The Karl programming language

```
// Closure:
// `a` is a function that returns another function.
// The inner function captures `n` from its definition scope.
let a = n -> {
    x -> n + x
}

// `for` as an expression:
// - The loop evaluates to a value.
// - `with` declares and initializes loop-local state.
// - The loop body mutates that state across iterations.
// - `then` introduces the *result expression* of the loop.
let value =
    for i < a(3)(10) with i = 0, b = 0 {
        log(i)
        b = i
        i++
    }
    // `then` accepts either:
    //   - a single expression, or
    //   - a block.
    //
    // In this case, `{ b, i, }` is an *object literal*.
    // The trailing comma is required to disambiguate:
    //   - `{ b }`      → returns `b` as the loop value
    //   - `{ b, }`     → returns an object `{ b: b }`
    then { b, i, }

log("value", value)

// ─────────────────────────────────────────────────────────────
// Concurrency / async tasks:
// - `&` starts an asynchronous computation and returns a task.
// - Tasks are composable via `.then(...)`.
let t0 = &http({
    method: "GET",
    url: "https://httpbin.org/delay/2",
    headers: map(),
})

let t1 = t0.then(res -> {
    // Destructure the resolved HTTP response.
    let { status, headers, body } = res

    log("status", status)
    log("Date", headers.get("Date"))

    // Runtime error handling:
    // `expr ? fallback` means “evaluate expr, or use fallback on error”.
    let b = decodeJson(body) ? {}

    // The last expression of the continuation
    // becomes the resolved value of `t1`.
    b.origin
})

// This executes immediately; the async task is still running.
log("⌛️")

// `wait` suspends until the task resolves and yields its value.
let origin = wait t1
origin
```