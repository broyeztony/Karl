// Count pod phases with group_count().
// Optional env: K8S_NS=<namespace>

let ns = env("K8S_NS") ? ""

let args = if ns == "" {
    ["get", "pods", "-A", "-o", "jsonpath={range .items[*]}{.status.phase}{\"\\n\"}{end}"]
} else {
    ["get", "pods", "-n", ns, "-o", "jsonpath={range .items[*]}{.status.phase}{\"\\n\"}{end}"]
}

let p = proc(
    { command: "kubectl", args: args, },
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let counts = p | lines() | filter(_ != "") | group_count()

let st = wait p
if !st.ok { fail("kubectl phase scan failed: exit " + str(st.code)) }

log(counts)
