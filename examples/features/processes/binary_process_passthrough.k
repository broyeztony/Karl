// Binary process passthrough: file -> process stdin -> process stdout -> file.
//
// Run:
//   karl run examples/features/processes/binary_process_passthrough.k

let srcPath = "/tmp/karl-bytes-proc-src.bin"
let dstPath = "/tmp/karl-bytes-proc-dst.bin"

let seed = run(cmd({
    command: "sh",
    args: ["-c", "printf '\\120\\113\\003\\004\\000\\377\\021\\042\\063\\012' > \"$1\"", "sh", srcPath],
}), {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

if !seed.ok { fail("seed failed") }

let p = proc(cmd({ command: "cat", }), {
    stdIn: PIPE,
    stdOut: PIPE,
    stdErr: NULL,
    stdinType: BYTES,
    stdoutType: BYTES,
})

let inTask = & (() -> pipe(reader(srcPath, { type: BYTES, }), p.stdin, { bufferSize: 4, }))()
let outTask = & (() -> pipe(p.stdout, writer(dstPath, { type: BYTES, }), { bufferSize: 4, }))()

let inStats = wait inTask
p.stdin.close() ? {}
let status = wait p
let outStats = wait outTask

let cmp = run(cmd({
    command: "sh",
    args: ["-c", "cmp -s \"$1\" \"$2\"", "sh", srcPath, dstPath],
}), {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

{
    pid: p.pid,
    status,
    input: inStats,
    output: outStats,
    identical: cmp.ok,
}
