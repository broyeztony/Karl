// Long-running process control in Karl: pid/running/abort/wait.
//
// Run:
//   karl run examples/features/processes/stream_and_abort.k

let takeBytes = (ch, limit) -> for true with r = ch.recv(), acc = "", n = 0 {
    let [chunk, closed] = r
    if closed { break { bytes: n, sample: acc[..200], } }

    let next = acc + chunk
    n = next.length
    if n >= limit { break { bytes: limit, sample: next[..200], } }

    acc = next
    r = ch.recv()
} then { bytes: n, sample: acc[..200], }

let stage = cmd({
    command: "yes",
    args: ["karl"],
})

let p = proc(stage, {
    stdOut: "pipe",
    stdErr: "pipe",
}) ? {
    fail("cannot start long-running command: " + error.message)
}

let outTask = & (() -> takeBytes(p.stdout, 8000))()
let runningBeforeAbort = p.running

sleep(120)

p.abort() ? {
    // If already stopped, keep going; wait will still return final status.
    ()
}

let st = wait p ? {
    {
        ok: false,
        code: -1,
        signal: null,
        timedOut: false,
        aborted: true,
        durationMs: 0,
        waitError: error.message,
    }
}

let out = wait outTask ? { { bytes: 0, sample: "", } }

let output = {
    pid: p.pid,
    runningBeforeAbort: runningBeforeAbort,
    runningAfterWait: p.running,
    status: st,
    stdoutPreview: out,
}

output
