// Array helpers: map/filter/reduce/sum/find/sort.

let nums = [1, 2, 3, 4, 5]
let evens = nums.filter(x -> x % 2 == 0)
let doubled = nums.map(x -> x * 2)
let total = nums.sum()
let reduced = nums.reduce((acc, x) -> acc + x, 0)
let found = nums.find(x -> x > 3)
let sorted = [3, 1, 2].sort((a, b) -> a - b)
let len = nums.length
let output = [evens, doubled, total, reduced, found, sorted, len]
output
