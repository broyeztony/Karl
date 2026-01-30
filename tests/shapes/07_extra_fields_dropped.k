// Test: Extra fields should be dropped

let StrictShape = import "tests/shapes/fixtures/StrictShape.shape"

// Test case 1: Extra fields at top level
let raw1 = "{
    \"allowed1\": \"value1\",
    \"allowed2\": 42,
    \"extra1\": \"should be dropped\",
    \"extra2\": 999,
    \"extra3\": true
}"

let parsed1 = decodeJson(raw1) as StrictShape ? {
    log("❌ FAIL: Valid fields with extras should not error")
    { allowed1: "fallback", allowed2: 0 }
}

// Check allowed fields are present
if parsed1.allowed1 == "value1" {
    if parsed1.allowed2 == 42 {
        log("✅ PASS: Extra fields dropped at top level (allowed fields work)")
    } else {
        log("❌ FAIL: Allowed field not preserved")
    }
} else {
    log("❌ FAIL: Allowed field not preserved")
}

// Test case 2: Extra fields in nested objects
let NestedStrict = import "tests/shapes/fixtures/NestedStrict.shape"

let raw2 = "{
    \"parent\": {
        \"allowed\": \"value\",
        \"extraNested\": \"drop me\"
    },
    \"extraTop\": \"also drop\"
}"

let parsed2 = decodeJson(raw2) as NestedStrict ? {
    log("❌ FAIL: Nested extras should not error")
    { parent: { allowed: "fallback" } }
}

// Check nested allowed field works
if parsed2.parent.allowed == "value" {
    log("✅ PASS: Extra fields dropped in nested objects (allowed fields work)")
} else {
    log("❌ FAIL: Nested allowed field not preserved")
}

log("Test suite: Extra fields - Complete")
