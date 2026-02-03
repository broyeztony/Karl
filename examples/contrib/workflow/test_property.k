// Test to see if we can access missing properties
let config1 = { a: 1 }
let config2 = { a: 1, b: null }

log("Testing property access...")

// This should fail
let test1 = config1.b

log("Test1:", test1)
