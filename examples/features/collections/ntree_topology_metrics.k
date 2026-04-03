// Hierarchical topology helpers: depth/height/lca/pathBetween/subtreeSize.

let t = ntree("root", { label: "Root", })

t.append("root", "a", { label: "A", })
t.append("root", "b", { label: "B", })
t.append("root", "c", { label: "C", })
t.append("a", "a1", { label: "A1", })
t.append("a", "a2", { label: "A2", })
t.append("a2", "a21", { label: "A21", })
t.append("c", "c1", { label: "C1", })

log({
  depthA21: t.depth("a21"),
  heightRoot: t.height("root"),
  lcaNear: t.lca("a1", "a21"),
  lcaFar: t.lca("a1", "c1"),
  pathNear: t.pathBetween("a1", "a21").map(n -> n.id),
  pathFar: t.pathBetween("a1", "c1").map(n -> n.id),
  subtreeA: t.subtreeSize("a"),
  subtreeRoot: t.subtreeSize("root"),
})
