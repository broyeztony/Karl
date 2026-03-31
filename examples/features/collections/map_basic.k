// Maps: set/get/has/delete/keys/values

let counts = map()
counts.set("apples", 2)
counts.set("bananas", 1)

let hasApples = counts.has("apples")
let cherries = counts.get("cherries")
let removed = counts.delete("bananas")

{
    hasApples: hasApples,
    cherries: cherries,
    removed: removed,
    keys: counts.keys(),
    values: counts.values(),
}

