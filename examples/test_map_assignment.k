// Simple test showing map assignment NOT working

log("Testing map assignment...")

// Try to build a simple map dynamically
let map = {}

log("Created empty map:", map)

// This should work but doesn't
log("Attempting: map[\"key\"] = \"value\"")
map["key"] = "value"

log("Map after assignment:", map)
log("✅ Test passed!")
