// stdout/stderr streams carry STRING chunks.
// For binary data, encode in the child process (hex/base64) before reading in Karl.
let streams = (import "examples/features/processes/process_stream_helpers.k")()

// Force TEXT mode because this demo aggregates string output.
let p = proc(
    {
        command: "sh",
        args: ["-c", "printf '\\000\\001\\377\\020' | od -An -t x1 | tr -d ' \\n'; echo 'warn: sample stderr' 1>&2"],
    }, { stdout: PIPE, stderr: PIPE, stdin: NULL, stdoutType: TEXT, stderrType: TEXT, }
)

let st = streams.waitCaptured(p)

log({
    ok: st.ok,
    code: st.code,
    stdoutHex: st.output,   // "0001ff10"
    stderr: st.error,       // "warn: sample stderr\n"
})
