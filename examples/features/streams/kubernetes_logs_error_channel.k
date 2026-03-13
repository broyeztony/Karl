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

let pod1 = pods[0]
let pod2 = pods[1]
let pod3 = pods[2]

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

let at = (counts, pod, level) -> {
    let v = get(counts, key(pod, level))
    if v == null { 0 } else { v }
}

let spaces = n -> for i < n with i = 0, out = "" {
    out = out + " "
    i++
} then out

let fill = (ch, n) -> for i < n with i = 0, out = "" {
    out = out + ch
    i++
} then out

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

let render = counts -> {
    log("\u001b[2J\u001b[H")
    log("ns=" + ns + " selector=" + selector)
    log("pod1=" + pod1)
    log("pod2=" + pod2)
    log("pod3=" + pod3)
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
    let process = proc(
        "kubectl", "logs", "-f",
        "-n", ns,
        podName,
        "--all-containers=true",
        "--tail=100",
        { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
    )
    {
        process,
        stream: process
            | lines()
            | filter(_ != "")
            | map(line -> { kind: "log", pod: podLabel, level: levelOf(line), }),
    }
}

let f1 = feed(pod1, "pod1")
let f2 = feed(pod2, "pod2")
let f3 = feed(pod3, "pod3")

let producer = merge(f1.stream, f2.stream, f3.stream)

let signaler = & (() -> {
    let [signal, done] = signalWatch(["SIGINT", "SIGTERM"]).recv()
    if !done {
        log("stopping", signal)
        f1.process.abort() ? {}
        f2.process.abort() ? {}
        f3.process.abort() ? {}
    }
})()

let counts = map()
render(counts)

for true with pair = producer.read(), last = now() {
    let [event, eof] = pair
    if eof { break () }

    bump(counts, event.pod, event.level)

    let t = now()
    if t - last >= matrixMs {
        render(counts)
        last = t
    }
    pair = producer.read()
} then ()

f1.process.abort() ? {}
f2.process.abort() ? {}
f3.process.abort() ? {}

(wait f1.process) ? {}
(wait f2.process) ? {}
(wait f3.process) ? {}

signaler.cancel()
(wait signaler) ? {}

render(counts)
