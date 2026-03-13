// Streaming text analytics with Unix pipeline composition in Karl.
//
// Run:
//   karl run examples/features/processes/text_analytics_pipeline.k

let streams = (import "examples/features/processes/process_stream_helpers.k")()

let corpus =
    "Karl makes process orchestration concise.\n" +
    "Karl pipelines are explicit and composable.\n" +
    "Karl can supervise workers, scrape logs, and ship reports.\n" +
    "Composable APIs matter for production automation.\n" +
    "Production automation benefits from clear syntax.\n"

let stage = {
    command: "sh",
    args: ["-c", "tr A-Z a-z | tr -cs 'a-z' '\\n' | sort | uniq -c | sort -nr | head -n 8"],
}

// Force TEXT mode: this pipeline is textual analytics.
let p = proc(stage, {
    stdin: PIPE,
    stdout: PIPE,
    stderr: PIPE,
    stdinType: TEXT,
    stdoutType: TEXT,
    stderrType: TEXT,
})

// Feed input corpus, then close stdin so the pipeline can complete.
let inStream = p.stdin
inStream.write(corpus)
inStream.close()

let st = streams.waitCaptured(p)

let output = if st.ok {
    {
        ok: true,
        durationMs: st.durationMs,
        topWords: st.output.trim().split("\n"),
    }
} else {
    {
        ok: false,
        code: st.code,
        signal: st.signal,
        stderr: st.error,
    }
}

output
