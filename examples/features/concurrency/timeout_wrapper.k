// Timeout wrapper pattern (similar in spirit to Go's context.WithTimeout).
// - Start work in a task.
// - Race "wait work" against a timer.
// - If timer wins, cancel work and fail with timeout.

let withTimeout = (ms, fn) -> {
    let work = & fn()

    let first = wait !& {
        (() -> {
            let value = wait work
            let out = { kind: "ok", value: value, }
            out
        })(),
        (() -> {
            sleep(ms)
            let out = { kind: "timeout", }
            out
        })()
    }

    if first.kind == "timeout" {
        work.cancel()
        fail("timeout after " + str(ms) + "ms")
    }

    first.value
}

let fast = () -> { sleep(20); "fast-result" }
let slow = () -> { sleep(120); "slow-result" }

let ok = withTimeout(100, fast)
let timed = withTimeout(40, slow) ? { "fallback (" + error.message + ")" }

let output = { ok: ok, timed: timed, }
output
