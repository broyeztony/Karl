// Kubernetes health report using Karl process APIs.
//
// Requires:
// - kubectl installed
// - current kube context configured
//
// Run:
//   karl run examples/features/processes/kubernetes_health_report.k
// Optional rollout check:
//   karl run examples/features/processes/kubernetes_health_report.k -- <namespace> <deployment>

let kRun = args -> run(cmd({
    command: "kubectl",
    args: args,
}), {
    maxOutputBytes: 8000000,
    overflow: "truncate",
}) ? { fail("kubectl spawn error: " + error.message) }

let kMust = args -> {
    let st = kRun(args)
    if st.ok { st } else { fail("kubectl failed: " + st.error) }
}

let kJson = args -> jsonDecode(kMust(args).output) ? {
    fail("kubectl returned invalid JSON: " + error.message)
}

let scorePod = pod -> {
    let statuses = pod.status.containerStatuses ? []
    let restarts = statuses.reduce((acc, cs) -> acc + cs.restartCount, 0)
    let notReady = statuses.filter(cs -> cs.ready == false).length
    {
        name: pod.metadata.name,
        phase: pod.status.phase,
        restarts: restarts,
        notReady: notReady,
        problematic: pod.status.phase != "Running" || restarts > 3 || notReady > 0,
    }
}

let inspectNamespace = ns -> {
    let pods = kJson(["get", "pods", "-n", ns, "-o", "json"]).items.map(scorePod)
    let problematic = pods.filter(p -> p.problematic)
    let top = problematic
        .sort((a, b) -> (b.restarts + b.notReady) - (a.restarts + a.notReady))
        [..5]
    {
        namespace: ns,
        podCount: pods.length,
        problematicPods: problematic.length,
        topProblems: top,
    }
}

let optionalRollout = () -> {
    let args = argv()
    if args.length != 2 { null }
    else {
        let ns = args[0]
        let deploy = args[1]
        let st = kRun(["rollout", "status", "deploy/" + deploy, "-n", ns, "--timeout=25s"])
        {
            namespace: ns,
            deployment: deploy,
            ok: st.ok,
            code: st.code,
            output: st.output.trim(),
            error: st.error.trim(),
        }
    }
}

let report = (() -> {
    let version = kJson(["version", "-o", "json"])
    let nodeCount = kJson(["get", "nodes", "-o", "json"]).items.length
    let namespaces = kJson(["get", "ns", "-o", "json"]).items.map(ns -> ns.metadata.name)

    // Fan-out: one kubectl pod scan per namespace.
    let tasks = namespaces.map(ns -> & inspectNamespace(ns))
    let perNs = tasks.map(t -> wait t ? { namespace: "<error>", podCount: 0, problematicPods: 0, topProblems: [], error: error.message, })

    let ranked = perNs.sort((a, b) -> b.problematicPods - a.problematicPods)

    {
        generatedAt: timeFormatRFC3339(now()),
        clusterGitVersion: version.serverVersion.gitVersion,
        nodeCount: nodeCount,
        namespaceCount: namespaces.length,
        totalPods: perNs.reduce((acc, row) -> acc + row.podCount, 0),
        totalProblematicPods: perNs.reduce((acc, row) -> acc + row.problematicPods, 0),
        busiest: ranked[..10],
        rolloutCheck: optionalRollout(),
    }
})() ? { generatedAt: timeFormatRFC3339(now()), error: { kind: error.kind, message: error.message, }, }

log(jsonEncode(report))
