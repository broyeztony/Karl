// Test loop variable scoping
let loopResult = for i < 3 with i = 0, sum = 0 {
    sum = sum + i
    i = i + 1
} then { total: sum }

log("loopResult:", loopResult)

let data = for i < 3 with i = 0, arr = [] {
    arr += [i * 2]
    i = i + 1
} then arr

log("data:", data)

{
    result: loopResult.total,
    numbers: data,
}
