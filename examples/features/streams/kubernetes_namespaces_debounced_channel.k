// Debounce bursty namespace stream values and bridge to a channel sink.

let q = buffered(128)

let p = proc(
    "kubectl",
    "get",
    "pods",
    "-A",
    "-o",
    "jsonpath={range .items[*]}{.metadata.namespace}{\"\\n\"}{end}",
    { stdout: PIPE, stderr: NULL, stdoutType: TEXT, }
)

let producer = & (p | lines() | filter(_ != "") | debounce(30) | toChannel(q))

let values = for true with out = [] {
    let [v, done] = q.recv()
    if done { break out }
    out += [v]
} then ()

wait producer
let st = wait p
if !st.ok { fail("kubectl namespaces stream failed: exit " + str(st.code)) }

log({ count: values.length, values, })
