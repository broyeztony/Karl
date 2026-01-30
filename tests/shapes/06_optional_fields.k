// Test: Optional fields handling

let AllOptional = import "tests/shapes/fixtures/AllOptional.shape"
let MixedFields = import "tests/shapes/fixtures/MixedFields.shape"

// Test case 1: All optional fields missing (should set to null)
let raw1 = "{}"
let parsed1 = decodeJson(raw1) as AllOptional ? {
    log("❌ FAIL: Empty object with all optional should not error")
    { field1: "fallback", field2: 0, field3: false }
}

if parsed1.field1 == null {
    if parsed1.field2 == null {
        if parsed1.field3 == null {
            log("✅ PASS: All missing optional fields set to null")
        } else {
            log("❌ FAIL: field3 should be null")
        }
    } else {
        log("❌ FAIL: field2 should be null")
    }
} else {
    log("❌ FAIL: field1 should be null")
}

// Test case 2: Some optional fields present
let raw2 = "{ \"field1\": \"present\" }"
let parsed2 = decodeJson(raw2) as AllOptional ? {
    log("❌ FAIL: Partial optional fields should not error")
    { field1: "fallback" }
}

if parsed2.field1 == "present" {
    if parsed2.field2 == null {
        if parsed2.field3 == null {
            log("✅ PASS: Some optional present, others null")
        } else {
            log("❌ FAIL: Absent optional should be null")
        }
    } else {
        log("❌ FAIL: Absent optional should be null")
    }
} else {
    log("❌ FAIL: Present field should have value")
}

// Test case 3: Mixed required/optional with all present
let raw3 = "{
    \"required1\": \"value1\",
    \"optional1\": \"value2\",
    \"required2\": 42
}"

let parsed3 = decodeJson(raw3) as MixedFields ? {
    log("❌ FAIL: All fields present should not error")
    { required1: "fallback", required2: 0 }
}

if parsed3.required1 == "value1" {
    if parsed3.optional1 == "value2" {
        log("✅ PASS: Mixed fields with all present")
    } else {
        log("❌ FAIL: Optional field not accessible")
    }
} else {
    log("❌ FAIL: Required field not accessible")
}

// Test case 4: Mixed with required missing (should error)
let raw4 = "{ \"optional1\": \"present\" }"
let parsed4 = decodeJson(raw4) as MixedFields ? {
    log("✅ PASS: Missing required field triggers error")
    { required1: "fallback", required2: 99 }
}

if parsed4.required1 == "fallback" {
    if parsed4.required2 == 99 {
        log("✅ PASS: Error recovery uses fallback for missing required")
    } else {
        log("❌ FAIL: Fallback not applied correctly")
    }
} else {
    log("❌ FAIL: Fallback not applied correctly")
}

// Test case 5: Mixed with optional missing (should not error)
let raw5 = "{ \"required1\": \"value\", \"required2\": 42 }"
let parsed5 = decodeJson(raw5) as MixedFields ? {
    log("❌ FAIL: Should not error with just optional missing")
    { required1: "fallback", required2: 0 }
}

if parsed5.optional1 == null {
    log("✅ PASS: Missing optional set to null in mixed shape")
} else {
    log("❌ FAIL: Optional should be null, got:", parsed5.optional1)
}

log("Test suite: Optional fields - Complete")
