// Binary stream copy: source file -> stream pipeline -> destination file.
//
// Run:
//   karl run examples/features/processes/binary_stream_copy.k

let srcPath = "/tmp/karl-bytes-src.bin"
let dstPath = "/tmp/karl-bytes-dst.bin"

// Build a binary payload directly in Karl.
let payload = bytesJoin([
    toUtf8("\u0000\u0001\u0002\u0003\u0004"),
    toUtf8("KARL"),
    toUtf8("\n"),
    toUtf8("\u0005\u0006\u0007\u0008\u0009"),
])

let src = writer(srcPath)
let seeded = src.write(payload)
src.close()

// Byte-for-byte copy using a stream pipeline.
let copied = 
        read(srcPath) 
        | write(dstPath)

// Read destination and rebuild payload from stream chunks.
let outChunks = 
    read(dstPath) 
    | collect()
let out = bytesJoin(outChunks)

let identical = out eqv payload

{
    copied,
    seeded,
    chunks: len(outChunks),
    bytes: len(out),
    identical,
}
