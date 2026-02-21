// Race (!& { ... }) returns the first *completion* (success or failure) and cancels losers.
// A common pattern is timeouts: race work against a timer that fails.

let state = { hits: 0 }

let slow = () -> {
    sleep(120)
    state.hits = state.hits + 1
    "slow"
}

let fast = () -> "fast"

let winner = wait !& { slow(), fast() }

// If losers were not canceled, slow would increment hits after ~120ms.
sleep(200)
let afterRaceHits = state.hits

// Timeout: fail after N ms.
let timeout = (ms) -> { sleep(ms); fail("timeout") }

let maybeSlow = () -> { sleep(100); "ok" }
let timed = (wait !& { maybeSlow(), timeout(30) }) ? {
    "timed out (" + error.message + ")"
}

// Race returns the first completion even if it's a failure.
let failFast = () -> fail("boom")
let okSlow = () -> { sleep(50); "ok" }
let firstCompletion = (wait !& { failFast(), okSlow() }) ? {
    "recovered: " + error.message
}

let output = { winner, afterRaceHits, timed, firstCompletion, }
output
