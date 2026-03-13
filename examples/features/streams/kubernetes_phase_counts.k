// Count pod phases with group_count().
// Optional env: K8S_NS=<namespace>

let ns = trim(env("K8S_NS")) ? ""
let scope = if ns == "" { ["-A"] } else { ["-n", ns] }
let args = ["get", "pods"] + scope + ["-o", "jsonpath={range .items[*]}{.status.phase}{\"\\n\"}{end}"]

let p = proc(
    { command: "kubectl", args: args, },
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let counts = p | lines() | filter(_ != "") | group_count()

let st = wait p
if !st.ok { fail("kubectl phase scan failed: exit " + str(st.code)) }

log(counts)
