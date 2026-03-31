// Ordered tree collections (AVL/Treap)

let idx = tree() // default AVL
idx.set(5, "e")
idx.set(2, "b")
idx.set(8, "h")
idx.set(3, "c")

{
    kind: idx.kind(),
    size: idx.size,
    keys: idx.keys(),
    items: idx.items(),
    min: idx.min(),
    max: idx.max(),
    lowerBound4: idx.lowerBound(4),
    upperBound5: idx.upperBound(5),
}

