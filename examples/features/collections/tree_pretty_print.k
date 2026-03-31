// Built-in tree rendering (Unix tree-like).
// log(t) renders structure with ASCII branches.

let t = tree()
t.set(5, "e")
t.set(2, "b")
t.set(8, "h")
t.set(3, "c")

log(t)
