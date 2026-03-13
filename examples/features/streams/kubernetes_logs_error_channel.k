// Stream logs from 3 pods and refresh a level matrix every 2s (top-style).
// Optional env:
// - K8S_NS (default: "karl-health-demo")
// - K8S_SELECTOR (default: "app=chain-node")
// - K8S_MATRIX_MS (default: 2000)

let ns = trim(env("K8S_NS")) ? "karl-health-demo"
let selector = trim(env("K8S_SELECTOR")) ? "app=chain-node"
let matrixMs = parseInt(trim(env("K8S_MATRIX_MS")) ? "2000") ? 2000

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
let at = (counts, pod, level) -> get(counts, key(pod, level)) ? 0

let spaces = n -> for i < n with i = 0, out = "" { out = out + " "; i++ } then out
let fill = (ch, n) -> for i < n with i = 0, out = "" { out = out + ch; i++ } then out
let pad = (text, width) -> if text.length >= width { text } else { text + spaces(width - text.length) }

let levelW = 7
let countW = 6
let border =
    "+"
    + fill("-", levelW + 2)
    + "+"
    + fill("-", countW + 2)
    + "+"
    + fill("-", countW + 2)
    + "+"
    + fill("-", countW + 2)
    + "+"

let row = (a, b, c, d) ->
    "| "
    + pad(a, levelW)
    + " | "
    + pad(b, countW)
    + " | "
    + pad(c, countW)
    + " | "
    + pad(d, countW)
    + " |"

let render = (counts, names) -> {
    log("\u001b[2J\u001b[H")
    log("ns=" + ns + " selector=" + selector)
    log("pod1=" + names[0])
    log("pod2=" + names[1])
    log("pod3=" + names[2])
    log("")
    log(border)
    log(row("LEVEL", "POD1", "POD2", "POD3"))
    log(border)
    log(row("DEBUG", str(at(counts, "pod1", "DEBUG")), str(at(counts, "pod2", "DEBUG")), str(at(counts, "pod3", "DEBUG"))))
    log(row("INFO", str(at(counts, "pod1", "INFO")), str(at(counts, "pod2", "INFO")), str(at(counts, "pod3", "INFO"))))
    log(row("WARN", str(at(counts, "pod1", "WARN")), str(at(counts, "pod2", "WARN")), str(at(counts, "pod3", "WARN"))))
    log(row("ERROR", str(at(counts, "pod1", "ERROR")), str(at(counts, "pod2", "ERROR")), str(at(counts, "pod3", "ERROR"))))
    log(row("OTHER", str(at(counts, "pod1", "OTHER")), str(at(counts, "pod2", "OTHER")), str(at(counts, "pod3", "OTHER"))))
    log(border)
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

let signaler = & (() -> {
    let [signal, done] = signalWatch(["SIGINT", "SIGTERM"]).recv()
    if !done {
        log("stopping", signal)
        feeds.forEach(f -> f.proc.abort() ? {})
    }
})()

let counts = map()
let last = now()
render(counts, names)

// Single stream statement: pull 3 pods, merge, count, and render.
merge(feeds.map(f -> f.stream))
    | forEach(event -> {
        bump(counts, event.pod, event.level)
        let t = now()
        if t - last >= matrixMs {
            render(counts, names)
            last = t
        }
    })

feeds.forEach(f -> f.proc.abort() ? {})
feeds.forEach(f -> (wait f.proc) ? {})

signaler.cancel()
(wait signaler) ? {}

render(counts, names)
