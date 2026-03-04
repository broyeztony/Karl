// Binary chunk loop + reconstruction without pipe().
//
// Run:
//   karl run examples/features/processes/binary_chunk_loop_reconstruct.k

let srcPath = "/tmp/karl-bin-loop-src.bin"
let dstPath = "/tmp/karl-bin-loop-dst.bin"

let seed = run(cmd({
    command: "sh",
    args: [
        "-c",
        "printf '\\211PNG\\r\\n\\032\\n' > \"$1\"; dd if=/dev/urandom bs=1024 count=12 >> \"$1\" 2>/dev/null",
        "sh",
        srcPath,
    ],
}), {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

if !seed.ok { fail("seed failed") }

let p = proc(cmd({ command: "cat", args: [srcPath], }), {
    stdOut: PIPE,
    stdErr: NULL,
    stdoutType: BYTES,
})

let collected = for true with chunks = [], sizes = [] {
    let [chunk, eof] = p.stdout.read(700)
    if eof { break { chunks, sizes, } }
    chunks += [chunk]
    sizes += [len(chunk)]
} then ()

let st = wait p
if !st.ok { fail("cat failed") }

let payload = bytesJoin(collected.chunks)
let w = writer(dstPath, { type: BYTES, })
let written = w.write(payload)
w.close()

let cmp = run(cmd({
    command: "sh",
    args: ["-c", "cmp -s \"$1\" \"$2\"", "sh", srcPath, dstPath],
}), {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

{
    ok: cmp.ok,
    chunks: len(collected.chunks),
    bytes: written,
    firstChunkBytes: if len(collected.sizes) > 0 { collected.sizes[0] } else { 0 },
    identical: cmp.ok,
}
