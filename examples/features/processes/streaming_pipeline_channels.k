// Streaming demo: 2-command pipeline + Karl channels.
//
// What this shows:
// - OS pipeline: producer command | transformer command
// - Karl consumes stdout/stderr incrementally through channels
// - A separate task waits process completion and closes the event stream
//
// Run:
//   karl run examples/features/processes/streaming_pipeline_channels.k

let plan =
    cmd("sh", ["-c", "i=1; while [ $i -le 12 ]; do echo \"evt-$i\"; i=$((i+1)); sleep 0.15; done"]) |
    cmd("awk", ["{ print toupper($0); fflush() }"])

let p = proc({
    plan: plan,
    stdIn: "null",
    stdOut: "pipe",
    stdErr: "pipe",
})

let events = channel()
let outCh = stdOut(p)
let errCh = stdErr(p)

let outTask = & (() -> for true with r = outCh.recv(), n = 0 {
    let [chunk, closed] = r
    if closed { break n }
    n++
    events.send({ kind: "out", n: n, line: chunk.trim(), at: now(), })
    r = outCh.recv()
} then n)()

let errTask = & (() -> for true with r = errCh.recv(), n = 0 {
    let [chunk, closed] = r
    if closed { break n }
    n++
    events.send({ kind: "err", n: n, line: chunk.trim(), at: now(), })
    r = errCh.recv()
} then n)()

& (() -> {
    let status = wait p
    let outCount = wait outTask ? { 0 }
    let errCount = wait errTask ? { 0 }
    events.send({ kind: "done", status: status, outCount: outCount, errCount: errCount, at: now(), })
    events.done()
})()

let startedAt = now()
let summary = for true with r = events.recv(), seen = [], done = null {
    let [event, closed] = r
    if closed { break { lines: seen, done: done, elapsedMs: now() - startedAt, } }

    match event {
        case { kind: "out", n, line, at } -> {
            log("OUT", { n: n, line: line, tMs: at - startedAt, })
            seen += [line]
        }
        case { kind: "err", n, line, at } -> {
            log("ERR", { n: n, line: line, tMs: at - startedAt, })
        }
        case { kind: "done", status, outCount, errCount, at } -> {
            log("DONE", { status: status, outCount: outCount, errCount: errCount, tMs: at - startedAt, })
            done = status
        }
        case _ -> ()
    }

    r = events.recv()
} then { lines: seen, done: done, elapsedMs: now() - startedAt, }

summary
