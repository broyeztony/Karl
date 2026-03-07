// Binary process passthrough: file -> process stdin -> process stdout -> file.
//
// Run:
//   karl run examples/features/processes/binary_process_passthrough.k

let srcPath = "/tmp/karl-bytes-proc-src.bin"
let dstPath = "/tmp/karl-bytes-proc-dst.bin"

// Seed a binary payload we will stream through a process.
let seed = run({
    command: "sh",
    args: ["-c", "printf '\\120\\113\\003\\004\\000\\377\\021\\042\\063\\012' > \"$1\"", "sh", srcPath],
}, {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

if !seed.ok { fail("seed failed") }

// cat echoes stdin -> stdout. Keep stdio in BYTES mode.
let p = proc({ command: "cat", }, {
    stdin: PIPE,
    stdout: PIPE,
    stderr: NULL,
    stdinType: BYTES,
    stdoutType: BYTES,
})

let pump = (src, dst, chunkSize) -> for true with bytes = 0, chunks = 0 {
    let [chunk, eof] = src.read(chunkSize)
    if eof { break { bytes: bytes, chunks: chunks, } }
    bytes += dst.write(chunk)
    chunks += 1
} then ()

let srcReader = reader(srcPath, { type: BYTES, })
let dstWriter = writer(dstPath, { type: BYTES, })

// Run input and output transfer concurrently to avoid stdin/stdout deadlocks.
let inTask = & pump(srcReader, p.stdin, 4)
let outTask = & pump(p.stdout, dstWriter, 4)

// Once input transfer is done, close stdin to let cat exit cleanly.
let inStats = wait inTask
srcReader.close() ? {}
p.stdin.close() ? {}
let status = wait p
let outStats = wait outTask
dstWriter.close() ? {}

// Validate payload identity.
let cmp = run({
    command: "sh",
    args: ["-c", "cmp -s \"$1\" \"$2\"", "sh", srcPath, dstPath],
}, {
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
