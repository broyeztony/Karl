// Streaming text analytics with Unix pipeline composition in Karl.
//
// Run:
//   karl run examples/features/processes/text_analytics_pipeline.k

let collect = s -> for true with acc = "" {
    let [chunk, eof] = s.read() ? { break acc }
    if eof { break acc }
    acc = acc + chunk
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

// Force TEXT mode: this pipeline is textual analytics.
let p = proc(plan, {
    stdIn: PIPE,
    stdOut: PIPE,
    stdErr: PIPE,
    stdinType: TEXT,
    stdoutType: TEXT,
    stderrType: TEXT,
})

let outTask = & (() -> collect(p.stdout))()
let errTask = & (() -> collect(p.stderr))()

// Feed input corpus, then close stdin so the pipeline can complete.
let inStream = p.stdin
inStream.write(corpus)
inStream.close()

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
