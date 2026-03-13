// Mirror pod names to console and spill to disk while counting.

let snapshot = if (trim(env("K8S_SNAPSHOT")) ? "") == "" {
    "/tmp/karl-pods.snapshot"
} else {
    trim(env("K8S_SNAPSHOT")) ? ""
}

let p = proc("kubectl", "get", "pods", "-A", "-o", "name",
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, })

let total = p
    | lines()
    | filter(_ != "")
    | take(40)
    | spill(snapshot, { type: TEXT, })
    | tee(stdout())
    | count()

let st = wait p
if !st.ok { fail("kubectl get pods failed: exit " + str(st.code)) }

log({ total, snapshot, })
