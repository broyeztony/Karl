// Simple test showing truthy/falsy NOT working

log("Testing truthy/falsy...")

let config = {
    retryPolicy: { maxAttempts: 3 },
    storage: null,
}

// This should work but doesn't - requires explicit != null
log("Attempting: if config.retryPolicy { ... }")
if config.retryPolicy {
    log("✅ Has retry policy!")
} else {
    log("❌ No retry policy")
}

log("Test completed!")
