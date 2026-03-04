// stdout/stderr streams carry STRING chunks.
// For binary data, encode in the child process (hex/base64) before reading in Karl.

let collect = s -> for true with r = s.read(), acc = "" {
    let [chunk, eof] = r
    if eof { break acc }
    acc = acc + chunk
    r = s.read()
} then acc

let p = proc(
    cmd({
        command: "sh",
        args: ["-c", "printf '\\000\\001\\377\\020' | od -An -t x1 | tr -d ' \\n'; echo 'warn: sample stderr' 1>&2"],
    }),
    { stdOut: "pipe", stdErr: "pipe", stdIn: "null", }
)

let outTask = & (() -> collect(p.stdout))()
let errTask = & (() -> collect(p.stderr))()

let st = wait p
let outHex = wait outTask
let errText = wait errTask

log({
    ok: st.ok,
    code: st.code,
    stdout_hex: outHex,   // "0001ff10"
    stderr: errText,      // "warn: sample stderr\n"
})
