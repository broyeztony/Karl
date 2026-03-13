// Join pod stream with node stream by node name.
// Optional env: K8S_NS=<namespace> (default: all namespaces)

let ns = trim(env("K8S_NS")) ? ""
let scope = if ns == "" { ["-A"] } else { ["-n", ns] }
let podArgs = ["get", "pods"] + scope + ["-o", "jsonpath={range .items[*]}{.spec.nodeName}|{.metadata.namespace}|{.metadata.name}{\"\\n\"}{end}"]

let pPods = proc(
    { command: "kubectl", args: podArgs, },
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let pNodes = proc(
    "kubectl",
    "get",
    "nodes",
    "-o",
    "jsonpath={range .items[*]}{.metadata.name}|{range .status.conditions[?(@.type==\"Ready\")]}{.status}{end}{\"\\n\"}{end}",
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let pods = pPods
    | lines()
    | filter(_ != "")
    | map(line -> {
        let p = line.split("|")
        { node: p[0], namespace: p[1], pod: p[2], }
    })

let nodes = pNodes
    | lines()
    | filter(_ != "")
    | map(line -> {
        let p = line.split("|")
        { node: p[0], ready: p[1], }
    })

let joined = join(pods, nodes, _.node, _.node)
    | map(pair -> {
        let pod = pair[0]
        let node = pair[1]
        { pod: pod.namespace + "/" + pod.pod, node: pod.node, nodeReady: node.ready, }
    })
    | take(20)
    | collect()

let stPods = wait pPods
let stNodes = wait pNodes
if !stPods.ok { fail("kubectl pods failed: exit " + str(stPods.code)) }
if !stNodes.ok { fail("kubectl nodes failed: exit " + str(stNodes.code)) }

log({ total: joined.length, sample: joined[..5], })
