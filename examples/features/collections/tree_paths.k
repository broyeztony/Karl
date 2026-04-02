// Root-to-node paths for ordered trees and hierarchical trees.

let idx = tree("avl")
let seeds = [50, 25, 75, 10, 30, 60, 80, 5, 15, 27, 33]
seeds.forEach(k -> idx.set(k, "v" + str(k)))

log("ordered tree:\n" + str(idx))
log("tree.path(33):", idx.path(33))
log("tree.path(61):", idx.path(61))

let hierarchy = ntree("root", { label: "root", })
hierarchy.append("root", "infra", { label: "infra", })
hierarchy.append("root", "apps", { label: "apps", })
hierarchy.append("infra", "db", { label: "db", })
hierarchy.append("infra", "cache", { label: "cache", })
hierarchy.append("apps", "api", { label: "api", })
hierarchy.append("api", "worker", { label: "worker", })

log("ntree:\n" + str(hierarchy))
log("ntree.path(\"worker\"):", hierarchy.path("worker").map(n -> n.id))
log("ntree.path(\"missing\"):", hierarchy.path("missing"))
