// Stream feature sweep:
// fromJson(), distinct(), sort(), group_count(), reduce_by_key(), top(), split(), exec()

let path = "/tmp/karl-stream-new-features.ndjson"

writeFile(path,
    "{\"team\":\"api\",\"ms\":120}\n" +
    "{\"team\":\"worker\",\"ms\":260}\n" +
    "{\"team\":\"api\",\"ms\":220}\n" +
    "{\"team\":\"web\",\"ms\":90}\n" +
    "{\"team\":\"worker\",\"ms\":180}\n"
)

let events = read(path, { type: TEXT, }) | lines() | fromJson()

let teams = events | map(_.team) | distinct() | sort() | collect()
let counts = events | map(_.team) | group_count()
let totals = events | reduce_by_key(_.team, 0, (acc, e) -> acc + e.ms)
let slowestTwo = events | map(_.ms) | top(2)
let byLatency = events | split(_.ms >= 200)
let cat = read(path, { type: TEXT, }) | exec("cat")

log({
    teams,
    count_api: counts.get("api"),
    count_worker: counts.get("worker"),
    total_api_ms: totals.get("api"),
    total_worker_ms: totals.get("worker"),
    slowest_two_ms: slowestTwo,
    slow_count: byLatency.left.length,
    fast_count: byLatency.right.length,
    exec_ok: cat.ok,
    exec_output_prefix: cat.output.split("\n")[0],
})
