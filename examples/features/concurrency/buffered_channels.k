// Example demonstrating buffered() builtin for buffered channels

log("=== Buffered Channel Example ===\n")

// Create a buffered channel with capacity 3
let chan = buffered(3)

log("Created buffered channel with capacity 3")

// Send 3 values without blocking (buffer has space)
chan.send("message 1")
log("Sent: message 1")

chan.send("message 2")
log("Sent: message 2")

chan.send("message 3")
log("Sent: message 3")

log("\nAll 3 messages sent without blocking!")
log("Now receiving messages...\n")

// Receive the values
let [msg1, done1] = chan.recv()
log("Received:", msg1)

let [msg2, done2] = chan.recv()
log("Received:", msg2)

let [msg3, done3] = chan.recv()
log("Received:", msg3)

log("\n✅ Buffered channels allow non-blocking sends up to buffer capacity!")
