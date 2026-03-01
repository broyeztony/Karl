// Streaming text analytics with Unix pipeline composition in Karl.
//
// Run:
//   karl run examples/features/processes/text_analytics_pipeline.k

let collect = ch -> for true with r = ch.recv(), acc = "" {
    let [chunk, closed] = r
    if closed { break acc }
    acc = acc + chunk
    r = ch.recv()
} then acc

let corpus =
    "Karl makes process orchestration concise.\n" +
    "Karl pipelines are explicit and composable.\n" +
    "Karl can supervise workers, scrape logs, and ship reports.\n" +
    "Composable APIs matter for production automation.\n" +
    "Production automation benefits from clear syntax.\n"

let plan =
    cmd({ command: "tr", args: ["A-Z", "a-z"], }) |
    cmd({ command: "tr", args: ["-cs", "a-z", "\\n"], }) |
    cmd({ command: "sort", args: [], }) |
    cmd({ command: "uniq", args: ["-c"], }) |
    cmd({ command: "sort", args: ["-nr"], }) |
    cmd({ command: "head", args: ["-n", "8"], })

let p = proc(plan, {
    stdIn: "pipe",
    stdOut: "pipe",
    stdErr: "pipe",
})

let outTask = & (() -> collect(p.stdout))()
let errTask = & (() -> collect(p.stderr))()

let inCh = p.stdin
inCh.send(corpus)
inCh.done()

let st = wait p
let out = wait outTask ? { "" }
let err = wait errTask ? { "" }

let output = if st.ok {
    {
        ok: true,
        durationMs: st.durationMs,
        topWords: out.trim().split("\n"),
    }
} else {
    {
        ok: false,
        code: st.code,
        signal: st.signal,
        stderr: err,
    }
}

output
