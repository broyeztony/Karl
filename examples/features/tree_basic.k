// Ordered tree collections (AVL/Treap)

let idx = tree() // default AVL
idx.set(5, "e")
idx.set(2, "b")
idx.set(8, "h")
idx.set(3, "c")

log("kind:", idx.kind())
log("size:", idx.size)
log("keys:", idx.keys())
log("items:", idx.items())
log("min:", idx.min())
log("max:", idx.max())
log("lowerBound(4):", idx.lowerBound(4))
log("upperBound(5):", idx.upperBound(5))

let removed = idx.delete(3)
log("deleted 3:", removed)
log("keys after delete:", idx.keys())
