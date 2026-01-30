// Test: Type mismatch handling

let StringField = import "tests/shapes/fixtures/StringField.shape"
let IntField = import "tests/shapes/fixtures/IntField.shape"
let BoolField = import "tests/shapes/fixtures/BoolField.shape"
let FloatField = import "tests/shapes/fixtures/FloatField.shape"

log("Testing type mismatches...")

// Test 1: String shape with int value (should error)
let raw1 = "{ \"value\": 42 }"
let parsed1 = decodeJson(raw1) as StringField ? {
    log("✅ PASS: String shape rejects int value")
    { value: "fallback" }
}

if parsed1.value == "fallback" {
    log("✅ PASS: Error recovery works for string/int mismatch")
} else {
    log("❌ FAIL: Should have used fallback")
}

// Test 2: Int shape with string value (should error)
let raw2 = "{ \"value\": \"not a number\" }"
let parsed2 = decodeJson(raw2) as IntField ? {
    log("✅ PASS: Int shape rejects string value")
    { value: 0 }
}

if parsed2.value == 0 {
    log("✅ PASS: Error recovery works for int/string mismatch")
} else {
    log("❌ FAIL: Should have used fallback")
}

// Test 3: Bool shape with int value (should error)
let raw3 = "{ \"value\": 1 }"
let parsed3 = decodeJson(raw3) as BoolField ? {
    log("✅ PASS: Bool shape rejects int value")
    { value: false }
}

if parsed3.value == false {
    log("✅ PASS: Error recovery works for bool/int mismatch")
} else {
    log("❌ FAIL: Should have used fallback")
}

// Test 4: Float shape with string value (should error)
let raw4 = "{ \"value\": \"3.14\" }"
let parsed4 = decodeJson(raw4) as FloatField ? {
    log("✅ PASS: Float shape rejects string value")
    { value: 0.0 }
}

if parsed4.value == 0.0 {
    log("✅ PASS: Error recovery works for float/string mismatch")
} else {
    log("❌ FAIL: Should have used fallback")
}

// Test 5: Correct types should work
let raw5 = "{ \"value\": \"correct\" }"
let parsed5 = decodeJson(raw5) as StringField ? {
    log("❌ FAIL: Valid string should not error")
    { value: "fallback" }
}

if parsed5.value == "correct" {
    log("✅ PASS: Correct string type validates")
} else {
    log("❌ FAIL: Should have parsed correctly")
}

log("Test suite: Type mismatches - Complete")
