// Test: Alias collision - internal key should win when both present

let Response = import "tests/shapes/fixtures/Response.shape"

// Test case 1: Both internal and external keys present
let raw1 = "{ \"data\": \"internal\", \"external-data\": \"external\" }"
let parsed1 = decodeJson(raw1) as Response ? { data: "fallback" }

// Internal key should win
if parsed1.data == "internal" {
    log("✅ PASS: Alias collision - internal key wins")
} else {
    log("❌ FAIL: Expected 'internal', got:", parsed1.data)
}

// Test case 2: Only external key present
let raw2 = "{ \"external-data\": \"external\" }"
let parsed2 = decodeJson(raw2) as Response ? { data: "fallback" }

// Should use alias mapping
if parsed2.data == "external" {
    log("✅ PASS: Alias mapping works with external key only")
} else {
    log("❌ FAIL: Expected 'external', got:", parsed2.data)
}

// Test case 3: Only internal key present
let raw3 = "{ \"data\": \"internal\" }"
let parsed3 = decodeJson(raw3) as Response ? { data: "fallback" }

// Should use internal key directly
if parsed3.data == "internal" {
    log("✅ PASS: Internal key works without alias")
} else {
    log("❌ FAIL: Expected 'internal', got:", parsed3.data)
}

log("Test suite: Alias collision - Complete")
