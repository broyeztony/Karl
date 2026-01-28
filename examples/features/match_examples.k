// Match expression examples: ranges, structural patterns, and guards.

// Range + wildcard
let value = 3
let label = match value {
    case 0 -> "zero"
    case 1..4 -> "small"
    case _ -> "other"
}

// Guard on a simple expression
let parity = match value {
    case _ if value % 2 == 0 -> "even"
    case _ -> "odd"
}

// Object pattern with guard
let response = { status: 404, body: "not found" }
let message = match response {
    case { status: 200, body } -> body
    case { status } if status >= 400 -> "error"
    case _ -> "unknown"
}

// Array/tuple patterns with guards
let pair = [2, 5]
let compare = match pair {
    case (a, b) if a == b -> "equal"
    case (a, b) if a > b -> "a>b"
    case (a, b) -> "a<b"
}

let output = [label, parity, message, compare]
output
