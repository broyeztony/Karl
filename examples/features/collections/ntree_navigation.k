// Hierarchical ntree with stable parent/children/sibling navigation.

let t = ntree("root", { label: "Root", })

t.append("root", "api", { label: "API", })
t.append("root", "worker", { label: "Worker", })
t.append("api", "http", { label: "HTTP", })
t.append("api", "jobs", { label: "Jobs", })
t.prepend("worker", "queue", { label: "Queue", })
t.move("jobs", "worker", { index: 1, })

{
    size: t.size,
    root: t.root().id,
    childrenRoot: t.children("root").map(n -> n.id),
    siblingsJobs: t.siblings("jobs").map(n -> n.id),
    ancestorsJobs: t.ancestors("jobs").map(n -> n.id),
    descendantsRoot: t.descendants("root").map(n -> n.id),
    findQueue: t.find(n -> n.id == "queue"),
}
