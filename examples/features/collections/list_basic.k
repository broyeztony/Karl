// Lists: map/filter/reduce/sort/find

let nums = [5, 1, 3, 3, 2]
let evens = nums.filter(x -> x % 2 == 0)
let doubled = nums.map(x -> x * 2)
let sorted = nums.sort((a, b) -> a - b)
let total = nums.reduce((acc, x) -> acc + x, 0)
let firstOver3 = nums.find(x -> x > 3)

{
    nums: nums,
    evens: evens,
    doubled: doubled,
    sorted: sorted,
    total: total,
    firstOver3: firstOver3,
}

