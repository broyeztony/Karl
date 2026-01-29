// 🎨 Karl Syntax Highlighting Test File
// This file demonstrates all syntax highlighting features!

// ═══════════════════════════════════════════════════════════
// KEYWORDS
// ═══════════════════════════════════════════════════════════

let x = 10
let name = "Karl"
let active = true

if x > 5 {
    log("Greater than 5")
} else {
    log("Less than or equal to 5")
}

for i < 10 with i = 0, sum = 0 {
    sum += i
    i++
} then sum

match value {
    case 0 -> "zero"
    case 1..10 -> "small"
    case _ -> "large"
}

// ═══════════════════════════════════════════════════════════
// ASYNC OPERATORS (Should stand out!)
// ═══════════════════════════════════════════════════════════

let task = & http({
    method: "GET",
    url: "https://example.com",
    headers: map(),
})

let fast = () -> "fast"
let slow = () -> { sleep(1000); "slow" }
let winner = wait | { fast(), slow() }

// ═══════════════════════════════════════════════════════════
// FUNCTIONS & ARROWS
// ═══════════════════════════════════════════════════════════

let add = (a, b) -> a + b
let multiply = (x, y) -> {
    x * y
}

let factorial = (n) -> {
    if n <= 1 { 1 } else { n * factorial(n - 1) }
}

// ═══════════════════════════════════════════════════════════
// BUILT-IN FUNCTIONS
// ═══════════════════════════════════════════════════════════

// I/O
log("Hello, Karl!")
let response = http({ method: "GET", url: "https://httpbin.org/json", headers: map() })
let data = decodeJson("{\"name\": \"Karl\"}")
let json = encodeJson({ name: "Karl", version: 1 })

// Concurrency
let ch = rendezvous()
ch.send("message")
let [msg, done] = ch.recv()
ch.done()

let promise = task.then(result -> result.status)
wait promise
sleep(1000)

// Collections
let items = [1, 2, 3, 4, 5]
let doubled = items.map(x -> x * 2)
let evens = items.filter(x -> x % 2 == 0)
let total = items.reduce((acc, x) -> acc + x, 0)
let sum = items.sum()
let found = items.find(x -> x > 3)
let sorted = items.sort((a, b) -> a - b)
let count = items.length

let myMap = map()
let mySet = set()

// Strings
let text = "  Hello, World!  "
let trimmed = text.trim()
let lower = text.toLower()
let upper = text.toUpper()
let parts = text.split(",")
let chars = text.chars()
let has = text.contains("World")
let starts = text.startsWith("Hello")
let ends = text.endsWith("!")
let replaced = text.replace("World", "Karl")

// ═══════════════════════════════════════════════════════════
// OPERATORS
// ═══════════════════════════════════════════════════════════

// Arithmetic
let a = 10 + 5
let b = 20 - 3
let c = 4 * 7
let d = 100 / 2
let e = 17 % 5

// Comparison
let eq = (x == y)
let neq = (x != y)
let lt = (x < y)
let gt = (x > y)
let lte = (x <= y)
let gte = (x >= y)

// Logical
let and = (true && false)
let or = (true || false)

// Assignment
let val = 10
val += 5
val -= 3
val *= 2
val /= 4
val %= 3

// Increment
val++
val--

// Range
let range = 1..10

// Error recovery
let parsed = decodeJson(body) ? { error: "failed", }
let result = divide(10, 0) ? { 0 }

// ═══════════════════════════════════════════════════════════
// STRINGS & NUMBERS
// ═══════════════════════════════════════════════════════════

let str1 = "double quotes"
let str2 = 'single quotes'
let escaped = "line 1\nline 2\ttab"

let decimal = 42
let float = 3.14159
let hex = 0xFF
let negative = -100

// ═══════════════════════════════════════════════════════════
// CONSTANTS
// ═══════════════════════════════════════════════════════════

let t = true
let f = false
let n = nil

// ═══════════════════════════════════════════════════════════
// COMPLEX EXAMPLE: Concurrent Pipeline
// ═══════════════════════════════════════════════════════════

let worker = (id, input_ch, output_ch) -> {
    for continue_loop < 100 with continue_loop = 0 {
        let [data, done] = input_ch.recv()
        
        if done {
            break {}
        } else {
            let task = & http({
                method: "GET",
                url: data.url,
                headers: map(),
            })
            
            let result = task.then(response -> {
                {
                    worker_id: id,
                    status: response.status,
                    body: decodeJson(response.body) ? {},
                }
            })
            
            output_ch.send(wait result)
        }
        
        continue_loop++
    } then {}
}

let ch_in = rendezvous()
let ch_out = rendezvous()

let tasks = for i < 3 with i = 0, ts = [] {
    ts += [& worker(i + 1, ch_in, ch_out)]
    i++
} then ts

// Feed work
for i < 5 with i = 0 {
    ch_in.send({ url: "https://example.com", })
    i++
} then {}
ch_in.done()

// Collect results
let results = for i < 5 with i = 0, rs = [] {
    let [r, done] = ch_out.recv()
    rs += [r]
    i++
} then rs

results

// ═══════════════════════════════════════════════════════════
// END OF SYNTAX TEST
// ═══════════════════════════════════════════════════════════
// If everything above looks colorful and beautiful, it works! ✨
