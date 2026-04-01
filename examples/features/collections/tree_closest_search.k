// Closest-node search is now built into ordered trees.

let idx = tree()
let seeds = [10, 20, 35, 50, 80]
seeds.forEach(k -> idx.set(k, "v" + str(k)))

{
    q20: idx.closest(20),
    q22: idx.closest(22),
    q36: idx.closest(36),
    q2: idx.closest(2),
    q99: idx.closest(99),
    q25upper: idx.closest(25, { tie: "upper", }),
}
