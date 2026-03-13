// Stream NDJSON over HTTP, decode each row, and aggregate.

let port = trim(env("STREAM_HTTP_PORT")) ? "18091"
let addr = "127.0.0.1:" + port

let srv = httpServe({
    addr: addr,
    routes: [{
        method: "GET",
        path: "/events",
        handler: req -> {
            status: 200,
            body:
                "{\"team\":\"api\",\"ms\":120}\n" +
                "{\"team\":\"worker\",\"ms\":260}\n" +
                "{\"team\":\"api\",\"ms\":220}\n",
        },
    }],
})

let counts = http("http://" + addr + "/events", { type: TEXT, })
    | lines()
    | json()
    | group_count(_.team)

httpServerStop(srv)

log({ api: counts.get("api"), worker: counts.get("worker"), })
