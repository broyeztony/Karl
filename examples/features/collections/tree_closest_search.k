// Closest-node search with float keys and structural metrics.

let idx = tree("avl")
let seeds = [0.5, 1.1, 1.8, 2.4, 3.0, 3.6, 4.1, 4.9, 5.7, 6.2, 7.8, 8.4, 9.9]
seeds.forEach(k -> idx.set(k, { score: k, label: "s" + str(k), }))

log("tree:\n" + str(idx))
log("size:", idx.size, "maxDepth:", idx.maxDepth(), "maxWidth:", idx.maxWidth())

let queries = [0.2, 1.45, 2.7, 4.55, 6.0, 8.6, 10.1]
queries.forEach(q -> {
    log(
        "query:", q,
        "closest:", idx.closest(q),
        "floor:", idx.floor(q),
        "ceil:", idx.ceil(q),
    )
})

log("tie-upper at 2.1:", idx.closest(2.1, { tie: "upper", }))
