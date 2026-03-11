// Merge pod + node feeds into one stream.
// Optional env: K8S_NS=<namespace> (default: all namespaces)

let ns = env("K8S_NS") ? ""

let podArgs = if ns == "" {
    ["get", "pods", "-A", "-o", "jsonpath={range .items[*]}{.metadata.namespace}/{.metadata.name}{\"\\n\"}{end}"]
} else {
    ["get", "pods", "-n", ns, "-o", "jsonpath={range .items[*]}{.metadata.namespace}/{.metadata.name}{\"\\n\"}{end}"]
}

let pPods = proc(
    { command: "kubectl", args: podArgs, },
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let pNodes = proc(
    "kubectl",
    "get",
    "nodes",
    "-o",
    "name",
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let sample = merge(
    pPods | lines() | filter(_ != "") | map("pod  " + _),
    pNodes | lines() | filter(_ != "") | map("node " + _),
)
    | take(40)
    | chunk(5)
    | collect()

let stPods = wait pPods
let stNodes = wait pNodes
if !stPods.ok { fail("kubectl pods failed: exit " + str(stPods.code)) }
if !stNodes.ok { fail("kubectl nodes failed: exit " + str(stNodes.code)) }

log({ groups: sample.length, sample: sample[..3], })
