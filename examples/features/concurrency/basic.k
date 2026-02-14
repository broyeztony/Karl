// & spawns tasks; wait joins; | races.

let work = () -> 3
let t0 = & work()
let t1 = t0.then(v -> v + 1)
let result = wait t1

let fast = () -> 1
let slow = () -> { sleep(5); 2 }
let first = wait | { slow(), fast() }

let output = [result, first]
output
