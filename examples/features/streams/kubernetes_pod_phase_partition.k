// Split pods into running vs non-running with split(pred).
// Optional env: K8S_NS=<namespace> (default: all namespaces)

let ns = trim(env("K8S_NS")) ? ""
let scope = if ns == "" { ["-A"] } else { ["-n", ns] }
let args = ["get", "pods"] + scope + ["-o", "jsonpath={range .items[*]}{.metadata.namespace}|{.metadata.name}|{.status.phase}{\"\\n\"}{end}"]

let p = proc(
    { command: "kubectl", args: args, },
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let grouped = p
    | lines()
    | filter(_ != "")
    | map(line -> {
        let cols = line.split("|")
        { namespace: cols[0], pod: cols[1], phase: cols[2], }
    })
    | split(_.phase == "Running")

let st = wait p
if !st.ok { fail("kubectl get pods failed: exit " + str(st.code)) }

let sample = xs -> if xs.length <= 5 { xs } else { xs[..5] }

log(toJson({
    runningCount: grouped.left.length,
    notRunningCount: grouped.right.length,
    runningSample: sample(grouped.left),
    notRunningSample: sample(grouped.right),
}))
