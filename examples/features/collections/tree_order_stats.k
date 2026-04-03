// Ordered-tree advanced methods: bulkLoad, rank/select, neighborhood, pops, validate.

let idx = tree("avl")
idx.bulkLoad([
  { key: 10.0, value: { name: "alpha", }, },
  [20.0, { name: "beta", }],
  [30.0, { name: "gamma", }],
  [40.0, { name: "delta", }],
  [50.0, { name: "epsilon", }],
])

log("initial:", idx.items())

let query = 22.7
log("rank30", idx.rank(30.0))
log("select2", idx.select(2))
log("neighbors", idx.kClosest(query, 2))
log("tieUpper", idx.kClosest(35.0, 2, { tie: "upper", }))
log("validate", idx.validate())

log("popMin:", idx.popMin())
log("popMax:", idx.popMax())
log("remaining:", idx.items())
