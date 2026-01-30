// Test: Deep nesting with mixed required/optional fields

let DeepStructure = import "tests/shapes/fixtures/DeepStructure.shape"

// Test case 1: All fields present (should succeed)
let raw1 = "{
    \"level1\": {
        \"level2\": {
            \"level3\": {
                \"required\": \"value\",
                \"optional\": \"also present\"
            }
        }
    }
}"

let parsed1 = decodeJson(raw1) as DeepStructure ? {
    log("❌ FAIL: Valid deep structure should not error")
    { level1: { level2: { level3: { required: "fallback" } } } }
}

if parsed1.level1.level2.level3.required == "value" {
    log("✅ PASS: Deep nesting with all fields present")
} else {
    log("❌ FAIL: Deep required field not accessible")
}

// Test case 2: Optional field missing (should set to null)
let raw2 = "{
    \"level1\": {
        \"level2\": {
            \"level3\": {
                \"required\": \"value\"
            }
        }
    }
}"

let parsed2 = decodeJson(raw2) as DeepStructure ? {
    log("❌ FAIL: Missing optional field should not error")
    { level1: { level2: { level3: { required: "fallback" } } } }
}

if parsed2.level1.level2.level3.optional == null {
    log("✅ PASS: Missing optional field set to null")
} else {
    log("❌ FAIL: Optional field should be null, got:", parsed2.level1.level2.level3.optional)
}

// Test case 3: Required field missing (should error and use fallback)
let raw3 = "{
    \"level1\": {
        \"level2\": {
            \"level3\": {}
        }
    }
}"

let parsed3 = decodeJson(raw3) as DeepStructure ? {
    log("✅ PASS: Missing required field triggers error recovery")
    { level1: { level2: { level3: { required: "fallback-used" } } } }
}

if parsed3.level1.level2.level3.required == "fallback-used" {
    log("✅ PASS: Error recovery fallback works for deep nesting")
} else {
    log("❌ FAIL: Fallback should be 'fallback-used'")
}

log("Test suite: Deep nesting - Complete")
