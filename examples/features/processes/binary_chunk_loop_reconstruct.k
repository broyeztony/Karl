// Binary chunk loop + reconstruction.
//
// Run:
//   karl run examples/features/processes/binary_chunk_loop_reconstruct.k

let srcPath = "/tmp/karl-bin-loop-src.bin"
let dstPath = "/tmp/karl-bin-loop-dst.bin"

// Build a realistic binary payload: PNG signature + random bytes.
let seed = run({
    command: "sh",
    args: [
        "-c",
        "printf '\\211PNG\\r\\n\\032\\n' > \"$1\"; dd if=/dev/urandom bs=1024 count=12 >> \"$1\" 2>/dev/null",
        "sh",
        srcPath,
    ],
}, {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

if !seed.ok { fail("seed failed") }

// Stream BYTES from process stdout.
let p = proc({ command: "cat", args: [srcPath], }, {
    stdout: PIPE,
    stderr: NULL,
    stdoutType: BYTES,
})

// Manual BYTES read loop:
// - read fixed-size chunks
// - keep chunk values for reconstruction
// - keep chunk sizes for quick diagnostics
let collected = for true with chunks = [], sizes = [] {
    let [chunk, eof] = p.stdout.read(700)
    if eof { break { chunks, sizes, } }
    chunks += [chunk]
    sizes += [len(chunk)]
} then ()

let st = wait p
if !st.ok { fail("cat failed") }

// Rebuild full payload from chunk list.
let payload = bytesJoin(collected.chunks)
let w = writer(dstPath, { type: BYTES, })
let written = w.write(payload)
w.close()

// Validate output equals input.
let cmp = run({
    command: "sh",
    args: ["-c", "cmp -s \"$1\" \"$2\"", "sh", srcPath, dstPath],
}, {
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
