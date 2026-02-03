// Comprehensive test for truthy/falsy support

log("=== TRUTHY/FALSY COMPREHENSIVE TESTS ===\n")

// Test 1: Null is falsy
log("Test 1: Null is falsy")
let nullVal = null
if nullVal {
    log("❌ FAILED: null should be falsy")
} else {
    log("✅ PASSED: null is falsy")
}

// Test 2: false is falsy
log("\nTest 2: false is falsy")
if false {
    log("❌ FAILED: false should be falsy")
} else {
    log("✅ PASSED: false is falsy")
}

// Test 3: true is truthy
log("\nTest 3: true is truthy")
if true {
    log("✅ PASSED: true is truthy")
} else {
    log("❌ FAILED: true should be truthy")
}

// Test 4: 0 is falsy
log("\nTest 4: 0 is falsy")
if 0 {
    log("❌ FAILED: 0 should be falsy")
} else {
    log("✅ PASSED: 0 is falsy")
}

// Test 5: Non-zero numbers are truthy
log("\nTest 5: Non-zero numbers are truthy")
if 42 {
    log("✅ PASSED: 42 is truthy")
} else {
    log("❌ FAILED: 42 should be truthy")
}

if -1 {
    log("✅ PASSED: -1 is truthy")
} else {
    log("❌ FAILED: -1 should be truthy")
}

// Test 6: Empty string is falsy
log("\nTest 6: Empty string is falsy")
if "" {
    log("❌ FAILED: empty string should be falsy")
} else {
    log("✅ PASSED: empty string is falsy")
}

// Test 7: Non-empty strings are truthy
log("\nTest 7: Non-empty strings are truthy")
if "hello" {
    log("✅ PASSED: \"hello\" is truthy")
} else {
    log("❌ FAILED: \"hello\" should be truthy")
}

// Test 8: Empty array is falsy
log("\nTest 8: Empty array is falsy")
if [] {
    log("❌ FAILED: empty array should be falsy")
} else {
    log("✅ PASSED: empty array is falsy")
}

// Test 9: Non-empty arrays are truthy
log("\nTest 9: Non-empty arrays are truthy")
if [1, 2, 3] {
    log("✅ PASSED: [1, 2, 3] is truthy")
} else {
    log("❌ FAILED: [1, 2, 3] should be truthy")
}

// Test 10: Empty object is falsy
log("\nTest 10: Empty object is falsy")
if {} {
    log("❌ FAILED: empty object should be falsy")
} else {
    log("✅ PASSED: empty object is falsy")
}

// Test 11: Non-empty objects are truthy
log("\nTest 11: Non-empty objects are truthy")
if { a: 1 } {
    log("✅ PASSED: { a: 1 } is truthy")
} else {
    log("❌ FAILED: { a: 1 } should be truthy")
}

// Test 12: Negation operator works with truthy/falsy
log("\nTest 12: Negation operator")
if !null {
    log("✅ PASSED: !null is truthy")
} else {
    log("❌ FAILED: !null should be truthy")
}

if !0 {
    log("✅ PASSED: !0 is truthy")
} else {
    log("❌ FAILED: !0 should be truthy")
}

if !"" {
    log("✅ PASSED: !\"\" is truthy")
} else {
    log("❌ FAILED: !\"\" should be truthy")
}

if !42 {
    log("❌ FAILED: !42 should be falsy")
} else {
    log("✅ PASSED: !42 is falsy")
}

// Test 13: Logical AND with truthy/falsy
log("\nTest 13: Logical AND")
if 1 && 2 {
    log("✅ PASSED: 1 && 2 is truthy")
} else {
    log("❌ FAILED: 1 && 2 should be truthy")
}

if 0 && 1 {
    log("❌ FAILED: 0 && 1 should be falsy")
} else {
    log("✅ PASSED: 0 && 1 is falsy")
}

// Test 14: Logical OR with truthy/falsy
log("\nTest 14: Logical OR")
if 0 || 1 {
    log("✅ PASSED: 0 || 1 is truthy")
} else {
    log("❌ FAILED: 0 || 1 should be truthy")
}

if 0 || "" {
    log("❌ FAILED: 0 || \"\" should be falsy")
} else {
    log("✅ PASSED: 0 || \"\" is falsy")
}

// Test 15: For loop with truthy/falsy condition
log("\nTest 15: For loop with truthy/falsy")
let iterations = for count < 10 with count = 3, iter = 0 {
    iter = iter + 1
    count = count + 1
} then iter

if iterations == 7 {
    log("✅ PASSED: for loop with truthy/falsy condition works")
} else {
    log("❌ FAILED: for loop iterations =", iterations, "expected 7")
}

log("\n=== ALL TESTS COMPLETED ===")
