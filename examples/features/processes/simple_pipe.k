// stdout/stderr streams carry STRING chunks.
// For binary data, encode in the child process (hex/base64) before reading in Karl.

let readAll = s -> for true with acc = "" {
    let [chunk, eof] = s.read() ? { break acc }
    if eof { break acc }
    acc = acc + chunk
} then acc

// Force TEXT mode because this demo aggregates string output.
let p = proc(
    {
        command: "sh",
        args: ["-c", "printf '\\000\\001\\377\\020' | od -An -t x1 | tr -d ' \\n'; echo 'warn: sample stderr' 1>&2"],
    }, { stdout: PIPE, stderr: PIPE, stdin: NULL, stdoutType: TEXT, stderrType: TEXT, }
)

let outTask = & (() -> readAll(p.stdout))()
let errTask = & (() -> readAll(p.stderr))()

let st = wait p
let outHex = wait outTask
let errText = wait errTask

log({
    ok: st.ok,
    code: st.code,
    stdoutHex: outHex,   // "0001ff10"
    stderr: errText,     // "warn: sample stderr\n"
})
