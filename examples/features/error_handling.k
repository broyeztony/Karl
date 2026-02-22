// Recoverable errors use ? { ... } with implicit error.

let raw = "{\"foo\": }"
let parsed = jsonDecode(raw) ? {
    let msg = error.message
    { error: msg, }
}

let divide = (a, b) -> {
    if b == 0 { fail("division by zero") }
    a / b
}
let result = divide(10, 0) ? { 0 }

let output = [parsed, result]
output
