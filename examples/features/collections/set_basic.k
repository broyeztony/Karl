// Sets: unique values + membership

let uniq = set()
uniq.add(1)
uniq.add(2)
uniq.add(2)

let hasOne = uniq.has(1)
let removedOne = uniq.delete(1)

{
    hasOne: hasOne,
    removedOne: removedOne,
    size: uniq.size,
    values: uniq.values(),
}

