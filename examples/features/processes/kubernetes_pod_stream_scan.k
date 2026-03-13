// Kubernetes pod scan using stream pipelines.
//
// Requires:
// - kubectl installed
// - current kube context configured
//
// Run (all namespaces):
//   karl run examples/features/processes/kubernetes_pod_stream_scan.k
// Run (single namespace):
//   karl run examples/features/processes/kubernetes_pod_stream_scan.k -- karl-health-demo

let streams = (import "examples/features/processes/process_stream_helpers.k")()
let topN = (xs, n) -> if xs.length <= n { xs } else { xs[..n] }
let toInt = s -> parseInt(s) ? 0

let getPodRows = ns -> {
    let args = if ns == null {
        ["get", "pods", "-A", "--no-headers"]
    } else {
        ["get", "pods", "-n", ns, "--no-headers"]
    }

    let p = proc({ command: "kubectl", args, }, {
        stdout: PIPE,
        stderr: PIPE,
        stdoutType: BYTES,
        stderrType: BYTES,
    }) ? {
        fail("kubectl spawn error: " + error.message)
    }

    let outTask = & (() -> p.stdout | lines() | collect())()
    let errTask = & streams.readStreamText(p.stderr)
    let st = wait p
    let rows = wait outTask ? []
    let err = wait errTask ? ""

    if st.ok { rows } else { fail("kubectl failed: " + err) }
}

let parseRow = (line, nsHint) -> {
    let cols = line.trim().split(" ").filter(c -> c != "")
    if cols.length < 5 { null }
    else {
        let namespace = if nsHint == null { cols[0] } else { nsHint }
        let pod = if nsHint == null { cols[1] } else { cols[0] }
        let ready = if nsHint == null { cols[2] } else { cols[1] }
        let status = if nsHint == null { cols[3] } else { cols[2] }
        let restarts = if nsHint == null { toInt(cols[4]) } else { toInt(cols[3]) }
        {
            namespace,
            pod,
            ready,
            status,
            restarts,
            problematic: (status != "Running" && status != "Completed") || restarts > 0,
        }
    }
}

let nsArg = if argv().length == 1 { argv()[0] } else { null }
let rows = getPodRows(nsArg).map(line -> parseRow(line, nsArg)).filter(r -> r != null)
let problematic = rows.filter(r -> r.problematic).sort((a, b) -> b.restarts - a.restarts)

{
    generatedAt: timeFormatRFC3339(now()),
    namespaceScope: if nsArg == null { "all" } else { nsArg },
    podCount: rows.length,
    problematicCount: problematic.length,
    topProblematic: topN(problematic, 20),
}
