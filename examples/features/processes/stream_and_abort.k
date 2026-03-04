// Long-running process control in Karl: pid/running/abort/wait.
//
// Run:
//   karl run examples/features/processes/stream_and_abort.k

// Read up to `limit` bytes from a TEXT stream for preview/logging.
let takeBytes = (s, limit) -> for true with acc = "", n = 0 {
    let [chunk, eof] = s.read() ? { break { bytes: n, sample: acc[..200], } }
    if eof { break { bytes: n, sample: acc[..200], } }

    let next = acc + chunk
    n = next.length
    if n >= limit { break { bytes: limit, sample: next[..200], } }

    acc = next
} then { bytes: n, sample: acc[..200], }

let p = proc(cmd({
    command: "yes",
    args: ["karl"],
}), {
    // Use TEXT mode to keep output preview logic simple in this demo.
    stdOut: PIPE,
    stdErr: PIPE,
    stdoutType: TEXT,
    stderrType: TEXT,
}) ? {
    fail("cannot start long-running command: " + error.message)
}

let outTask = & (() -> takeBytes(p.stdout, 8000))()
let runningBeforeAbort = p.running

sleep(120)

// Abort the process, then wait to collect final status.
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

{
    pid: p.pid,
    runningBeforeAbort,
    runningAfterWait: p.running,
    status: st,
    stdout: out,
}
