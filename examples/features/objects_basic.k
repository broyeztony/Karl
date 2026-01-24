// Object literals support shorthand and spread.

let x = 1
let y = 2
let base = { x, y, }
let extra = { ...base, y: 3, z: 4 }
extra
