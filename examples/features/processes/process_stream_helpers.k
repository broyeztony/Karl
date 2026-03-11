// Reusable helpers for process stdout/stderr stream capture.

let readStreamText = s -> for true with acc = "" {
    let [chunk, eof] = s.read()
    if eof { break acc }
    let part = decodeUtf8(chunk) ? chunk
    acc = acc + part
} then acc

let waitCaptured = p -> {
    let outTask = & readStreamText(p.stdout)
    let errTask = & readStreamText(p.stderr)
    let st = wait p
    {
        ok: st.ok,
        code: st.code,
        signal: st.signal,
        timedOut: st.timedOut,
        aborted: st.aborted,
        durationMs: st.durationMs,
        output: wait outTask ? {
            fail("stdout capture error: " + error.message)
        },
        error: wait errTask ? {
            fail("stderr capture error: " + error.message)
        },
    }
}

let spawnCaptured = (spec, opts) -> {
    let p = proc(spec, opts) ? {
        fail("process spawn error: " + error.message)
    }
    waitCaptured(p)
}
