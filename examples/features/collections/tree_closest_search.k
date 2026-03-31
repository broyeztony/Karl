// Find exact or closest node for a numeric key using binary search over sorted keys.

let firstGE = (xs, target) -> {
    for lo < hi with lo = 0, hi = xs.length {
        let mid = floor((lo + hi) / 2)
        if xs[mid] < target { lo = mid + 1 } else { hi = mid }
    } then lo
}

let closestNode = (t, target) -> {
    let keys = t.keys()
    if keys.length == 0 { null } else {
        let i = firstGE(keys, target)

        if i < keys.length && keys[i] == target {
            { key: keys[i], value: t.get(keys[i]), exact: true, }
        } else {
            let left = if i > 0 { keys[i - 1] } else { null }
            let right = if i < keys.length { keys[i] } else { null }

            let chosen = if left == null {
                right
            } else if right == null {
                left
            } else if (target - left) <= (right - target) {
                left
            } else {
                right
            }

            { key: chosen, value: t.get(chosen), exact: false, }
        }
    }
}

let idx = tree()
let seeds = [10, 20, 35, 50, 80]
seeds.forEach(k -> idx.set(k, "v" + str(k)))

{
    q20: closestNode(idx, 20),
    q22: closestNode(idx, 22),
    q36: closestNode(idx, 36),
    q2 : closestNode(idx, 2),
    q99: closestNode(idx, 99),
}
