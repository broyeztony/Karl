// Streaming demo: 2-command pipeline + stream reads + event channel.
//
// What this shows:
// - OS pipeline: producer command | transformer command
// - Karl consumes stdout/stderr incrementally through streams
// - A separate task waits process completion and closes the event stream
//
// Run:
//   karl run examples/features/processes/streaming_pipeline_channels.k

let stage = {
    command: "sh",
    args: ["-c", "i=1; while [ $i -le 12 ]; do echo \"evt-$i\"; i=$((i+1)); sleep 0.15; done | awk '{ print toupper($0); fflush() }'"],
}

let p = proc(stage, {
    stdIn: NULL,
    stdOut: PIPE,
    stdErr: PIPE,
    stdoutType: TEXT,
    stderrType: TEXT,
})

logt("pid", p.pid)

let events = channel()

// Convert a stream into structured events.
let pump = (kind, stream) -> for true with n = 0 {
    let [chunk, eof] = stream.read() ? { break n }
    if eof { break n }
    n++
    events.send({ kind, n, line: chunk.trim(), at: now(), })
} then n

let outTask = & (() -> pump("out", p.stdout))()
let errTask = & (() -> pump("err", p.stderr))()

// Wait process completion in background and emit final summary event.
& (() -> {
    let status = wait p
    let outCount = wait outTask ? { 0 }
    let errCount = wait errTask ? { 0 }
    events.send({ kind: "done", status: status, outCount: outCount, errCount: errCount, at: now(), })
    events.done()
})()

let startedAt = now()
// Single consumer loop over mixed stdout/stderr/done events.
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
