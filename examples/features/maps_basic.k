// Map supports set/get/has/delete and keys/values.

let counts = map()
counts.set("apples", 2)
counts.set("bananas", 1)
let hasApples = counts.has("apples")
let missing = counts.get("cherries")
let removed = counts.delete("bananas")
let keys = counts.keys()
let values = counts.values()
let output = [hasApples, missing, removed, keys.length, values.length]
output
