// Binary stream copy: file reader -> file writer.
//
// Run:
//   karl run examples/features/processes/binary_stream_copy.k

let srcPath = "/tmp/karl-bytes-src.bin"
let dstPath = "/tmp/karl-bytes-dst.bin"

let seed = run(cmd({
    command: "sh",
    args: ["-c", "printf '\\000\\001\\002\\003\\004\\377\\376\\375\\374\\012' > \"$1\"", "sh", srcPath],
}), {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

if !seed.ok { fail("seed failed") }

let copied = pipe(
    reader(srcPath, { type: BYTES, }),
    writer(dstPath, { type: BYTES, }),
    { bufferSize: 3, }
)

let cmp = run(cmd({
    command: "sh",
    args: ["-c", "cmp -s \"$1\" \"$2\"", "sh", srcPath, dstPath],
}), {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

let hex = run(cmd({
    command: "sh",
    args: ["-c", "od -An -t x1 \"$1\" | tr -d ' \\n'", "sh", dstPath],
}), {
    maxOutputBytes: 4096,
    overflow: "truncate",
})

{
    copied,
    identical: cmp.ok,
    hex: hex.output,
}
