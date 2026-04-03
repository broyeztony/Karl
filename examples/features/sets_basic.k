// Set keeps unique values only.

let uniq = set()
uniq.add(1)
uniq.add(2)
uniq.add(2)
let hasOne = uniq.has(1)
let removed = uniq.delete(1)
let values = uniq.values()
let size = uniq.size
let output = [hasOne, removed, size, values.length]
output
