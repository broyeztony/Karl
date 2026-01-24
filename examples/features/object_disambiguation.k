// Object literals vs blocks (trailing comma disambiguates shorthand).

let x = 1
let y = 2
let other = { z: 3 }

// Object literals
let o1 = { x: 1, y: 2 }
let o2 = { x, y, }          // shorthand + trailing comma
let o3 = { ...other, x, y, } // spread forces object
let o4 = { }                // empty object
let o5 = ({ x, y, })        // explicit grouping, still needs trailing comma

// Blocks (same braces, no object signals)
let b1 = { x }              // block returns x
let b2 = { let sum = x + y; sum }

let loopObj = for i < 1 with i = 0 {
    i++
} then { x, y, }            // object in then block (trailing comma)

let loopBlock = for i < 1 with i = 0 {
    i++
} then { x }                // block in then

let output = [o1.x, o2.y, b1, b2, loopObj.x, loopBlock]
output
