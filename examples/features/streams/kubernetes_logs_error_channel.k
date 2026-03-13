// Follow logs from one deployment and stream ERROR lines to a channel.
// Optional env:
// - K8S_NS (default: "default")
// - K8S_DEPLOY (default: "api")
// - K8S_ERROR_LIMIT (default: 30)

let envText = (key, fallback) -> {
    let value = trim(env(key)) ? ""
    if value == "" { fallback } else { value }
}

let ns = envText("K8S_NS", "default")
let deploy = envText("K8S_DEPLOY", "api")
let limit = parseInt(env("K8S_ERROR_LIMIT") ? "30") ? 30

let q = buffered(256)

let p = proc(
    "kubectl",
    "logs",
    "-f",
    "-n",
    ns,
    "deploy/" + deploy,
    "--all-containers=true",
    "--tail=200",
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let producer = & (p | lines() | filter(_.contains("ERROR")) | take(limit) | send(q))

let consumed = for true with n = 0 {
    let [line, done] = q.recv()
    if done { break n }
    log("error", n + 1, line)
    n++
} then ()

wait producer ? {}
p.abort() ? {}
wait p ? {}

log({ consumed, })
