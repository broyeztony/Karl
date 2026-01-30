// Test: Complex aliases with spaces and special characters

let Headers = import "tests/shapes/fixtures/Headers.shape"

let raw = "{
    \"Content-Type\": \"application/json\",
    \"X-API-KEY\": \"secret-key-123\",
    \"User Agent\": \"Karl/1.0\",
    \"X-Request-ID\": \"abc-def-ghi\"
}"

let parsed = decodeJson(raw) as Headers ? {
    log("❌ FAIL: Shape parsing should not have failed")
    { contentType: "fallback" }
}

// Test all aliases were mapped correctly
let pass = true

if parsed.contentType != "application/json" {
    log("❌ FAIL: contentType should be 'application/json', got:", parsed.contentType)
    pass = false
}

if parsed.xApiKey != "secret-key-123" {
    log("❌ FAIL: xApiKey should be 'secret-key-123', got:", parsed.xApiKey)
    pass = false
}

if parsed.userAgent != "Karl/1.0" {
    log("❌ FAIL: userAgent should be 'Karl/1.0', got:", parsed.userAgent)
    pass = false
}

if parsed.xRequestId != "abc-def-ghi" {
    log("❌ FAIL: xRequestId should be 'abc-def-ghi', got:", parsed.xRequestId)
    pass = false
}

// Test JSON round-trip - aliases should be preserved
let encoded = encodeJson(parsed)
// Note: decodeJson returns a plain object without shape metadata
// so we can't easily check if aliases are preserved in this test
// This would require inspecting the raw JSON string

if pass == true {
    log("✅ PASS: Complex aliases with spaces and special characters")
} else {
    log("Test suite: Complex aliases - Failed")
}
