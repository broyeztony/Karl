// Ordered-tree search helpers: floor/ceil/predecessor/successor/range.

let idx = tree("avl")
let seeds = [10, 20, 35, 50, 80]
seeds.forEach(k -> idx.set(k, "v" + str(k)))

{
    floor36: idx.floor(36),
    ceil36: idx.ceil(36),
    pred35: idx.predecessor(35),
    succ35: idx.successor(35),
    rangeInclusive: idx.range(20, 50),
    rangeExclusive: idx.range(20, 50, { includeFrom: false, includeTo: false, }),
    rangeLimited: idx.range(10, 80, { limit: 3, }),
}
