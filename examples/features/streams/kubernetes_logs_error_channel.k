// Stream logs from 3 pods and refresh a level matrix every 2s (top-style).
// Optional env:
// - K8S_NS (default: "karl-health-demo")
// - K8S_SELECTOR (default: "app=chain-node")
// - K8S_MATRIX_MS (default: 2000)

let matrix = (import "examples/features/streams/kubernetes_log_matrix_render.k")()

let ns = trim(env("K8S_NS")) ? "karl-health-demo"
let selector = trim(env("K8S_SELECTOR")) ? "app=chain-node"
let matrixMs = parseInt(trim(env("K8S_MATRIX_MS"))) ? 2000

let listPods = () -> {
    let st = run(
        "kubectl", "get", "pods",
        "-n", ns,
        "-l", selector,
        "--field-selector=status.phase=Running",
        "-o", "jsonpath={range .items[*]}{.metadata.name}{\"\\n\"}{end}",
    ) ? { fail("kubectl get pods failed: " + error.message) }

    if !st.ok { fail("kubectl get pods failed: exit " + str(st.code) + " " + st.error) }
    st.output.split("\n").filter(_ != "")
}

let pods = listPods()
if pods.length < 3 { fail("expected at least 3 running pods for selector " + selector) }

let levelOf = line -> if line.contains("ERROR") {
    "ERROR"
} else if line.contains("WARN") {
    "WARN"
} else if line.contains("INFO") {
    "INFO"
} else if line.contains("DEBUG") {
    "DEBUG"
} else {
    "OTHER"
}

let key = (pod, level) -> pod + "|" + level
let bump = (counts, pod, level) -> {
    let k = key(pod, level)
    let cur = get(counts, k)
    set(counts, k, if cur == null { 1 } else { cur + 1 })
}

let feed = (podName, podLabel) -> {
    let p = proc(
        "kubectl", "logs", "-f",
        "-n", ns,
        podName,
        "--all-containers=true",
        "--tail=100",
        { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
    )
    {
        proc: p,
        stream: p | lines() | filter(_ != "") | map(line -> { pod: podLabel, level: levelOf(line), }),
    }
}

let feeds = [
    feed(pods[0], "pod1"),
    feed(pods[1], "pod2"),
    feed(pods[2], "pod3"),
]
let names = [pods[0], pods[1], pods[2]]
let ui = matrix.create(ns, selector, names)

let signaler = & (() -> {
    let [signal, done] = signalWatch(["SIGINT", "SIGTERM"]).recv()
    if !done {
        log("stopping", signal)
        feeds.forEach(f -> f.proc.abort() ? {})
    }
})()

let counts = map()
let last = now()
ui.render(counts)

// Single stream statement: pull 3 pods, merge, count, and render.
merge(feeds.map(f -> f.stream))
    | forEach(event -> {
        bump(counts, event.pod, event.level)
        let t = now()
        if t - last >= matrixMs {
            ui.render(counts)
            last = t
        }
    })

feeds.forEach(f -> f.proc.abort() ? {})
feeds.forEach(f -> (wait f.proc) ? {})

signaler.cancel()
(wait signaler) ? {}

ui.render(counts)
