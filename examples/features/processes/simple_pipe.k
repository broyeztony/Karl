
let plan = cmd({ command: "ls", }) | cmd({ command: "wc", args: ["-l"], })
let ph = proc(plan)

wait ph
