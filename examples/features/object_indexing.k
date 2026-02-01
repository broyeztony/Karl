// Bracket indexing lets objects use non-identifier keys.

let headers = decodeJson("{\"Content-Type\":\"application/json\",\"User Agent\":\"Karl\"}")

let contentType = headers["Content-Type"]
let userAgent = headers["User Agent"]

headers["X-Request-Id"] = "req-123"
headers.server = "karl"

log("headers:", headers)

log({
    contentType,
    userAgent,
    requestId: headers["X-Request-Id"],
    server: headers.server,
})

let json = encodeJson(headers)
log("json:", json)