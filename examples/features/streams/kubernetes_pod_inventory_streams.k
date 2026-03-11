// Build a pod inventory with stream transforms.
// Optional env: K8S_NS=<namespace>

let ns = env("K8S_NS") ? ""

let args = if ns == "" {
    ["get", "pods", "-A", "-o", "jsonpath={range .items[*]}{.metadata.namespace}|{.metadata.name}|{.status.phase}{\"\\n\"}{end}"]
} else {
    ["get", "pods", "-n", ns, "-o", "jsonpath={range .items[*]}{.metadata.namespace}|{.metadata.name}|{.status.phase}{\"\\n\"}{end}"]
}

let p = proc(
    { command: "kubectl", args: args, },
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let pods = p
    | lines()
    | filter(_ != "")
    | map(line -> {
        let p = line.split("|")
        { namespace: p[0], pod: p[1], phase: p[2], }
    })
    | collect()

let st = wait p
if !st.ok { fail("kubectl get pods failed: exit " + str(st.code)) }

let sample = if pods.length <= 8 { pods } else { pods[..8] }
log({ total: pods.length, sample, })
