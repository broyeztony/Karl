// Each factory call returns a fresh module instance.

let makeCounter = import "examples/features/import_instances_module.k"
let a = makeCounter()
let b = makeCounter()

b.inc()
a.inc()
log(a.get(), b.get())

a.inc()
log(a.get(), b.get())

a.count = 4
log(a.get(), b.get())

b.get = () -> 99
log(a.get(), b.get())