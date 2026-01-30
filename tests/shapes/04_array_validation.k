// Test: Array validation edge cases

let Items = import "tests/shapes/fixtures/Items.shape"
let Numbers = import "tests/shapes/fixtures/Numbers.shape"

// Test case 1: Empty array (should be valid)
let raw1 = "[]"
let parsed1 = decodeJson(raw1) as Items ? { 
    log("❌ FAIL: Empty array should be valid")
    decodeJson("[]") 
}
log("✅ PASS: Empty array validates correctly")

// Test case 2: Valid array of objects
let raw2 = "[
    { \"id\": 1, \"name\": \"Alice\" },
    { \"id\": 2, \"name\": \"Bob\" }
]"

let parsed2 = decodeJson(raw2) as Items ? { 
    log("❌ FAIL: Valid array should not error")
    decodeJson("[]") 
}

if parsed2[0].id == 1 {
    if parsed2[0].name == "Alice" {
        log("✅ PASS: Array of objects with all required fields")
    } else {
        log("❌ FAIL: Name not accessible")
    }
} else {
    log("❌ FAIL: ID not accessible")
}

// Test case 3: Array with missing required field (should error)
let raw3 = "[
    { \"id\": 1, \"name\": \"Alice\" },
    { \"id\": 2 }
]"

let didError3 = false
let parsed3 = decodeJson(raw3) as Items ? {
    didError3 = true
    decodeJson("[]")
}

if didError3 == true {
    log("✅ PASS: Array with invalid element triggers error recovery")
} else {
    log("❌ FAIL: Should have triggered error for missing required field")
}

// Test case 4: Primitive array with correct types
let raw4 = "[1, 2, 3, 4, 5]"
let parsed4 = decodeJson(raw4) as Numbers ? { 
    log("❌ FAIL: Valid number array should not error")
    decodeJson("[]") 
}

if parsed4[0] == 1 {
    if parsed4[4] == 5 {
        log("✅ PASS: Primitive array with correct types")
    } else {
        log("❌ FAIL: Array elements not correct")
    }
} else {
    log("❌ FAIL: Array elements not accessible")
}

// Test case 5: Primitive array with wrong type (should error)
let raw5 = "[1, 2, \"not a number\", 4]"
let didError5 = false
let parsed5 = decodeJson(raw5) as Numbers ? {
    didError5 = true
    decodeJson("[]")
}

if didError5 == true {
    log("✅ PASS: Array with wrong type element triggers error")
} else {
    log("❌ FAIL: Should have triggered error for type mismatch")
}

log("Test suite: Array validation - Complete")
