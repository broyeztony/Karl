// readLine() returns one line at a time, null on EOF.
//
// Run this example with stdin enabled:
//   printf "alpha\nbeta\n" | karl run examples/features/stdin_readline.k -- --read
//   karl run examples/features/stdin_readline.k -- --read   (then type lines, `quit`, Enter)
//
// Without --read, the example stays non-blocking for automated runs.

let shouldRead = argv().length > 0 && argv()[0] == "--read"

let lines = if shouldRead {
    for true with acc = [] {
        let line = readLine()
        // Stop on EOF (Ctrl+D) or explicit sentinel.
        if line == null { break acc }
        if line == "quit" { break acc }
        acc += [line]
    } then acc
} else {
    ["pass -- --read to consume stdin"]
}

let output = { shouldRead, lines, }
output
