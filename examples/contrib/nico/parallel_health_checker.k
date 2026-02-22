// 🎯 ULTIMATE KARL SHOWCASE
// Real-World Parallel HTTP Health Checker
// Demonstrates ALL of Karl's power in one program!

// Test endpoints
let endpoints = [
    { name: "Fast API", url: "https://httpbin.org/get", },
    { name: "Delayed 1s", url: "https://httpbin.org/delay/1", },
    { name: "User Agent", url: "https://httpbin.org/user-agent", },
]

// Worker function: check a single URL
let check_endpoint = (endpoint, worker_id, ch) -> {
    log("🔧 Worker", worker_id, "checking:", endpoint.name)
    
    // Spawn async HTTP request
    let task = & http({
        method: "GET",
        url: endpoint.url,
        headers: map(),
    })
    
    // Handle the response with error recovery
    let result = task.then(response -> {
        let body_parsed = jsonDecode(response.body) ? { error: "parse failed", }
        
        {
            worker_id: worker_id,
            name: endpoint.name,
            url: endpoint.url,
            status: response.status,
            success: response.status == 200,
            body_preview: body_parsed,
        }
    })
    
    let final_result = wait result
    
    // Send result via channel
    ch.send(final_result)
    
    log("✅ Worker", worker_id, "done:", endpoint.name, "- Status", final_result.status)
    final_result
}

// Collector function: aggregate all results via channel
let collect_results = (ch, num_workers) -> {
    log("📊 Collector waiting for", num_workers, "workers...")
    
    for i < num_workers with i = 0, collected = [] {
        let [result, done] = ch.recv()
        
        if done {
            break collected
        } else {
            collected = collected += [result]
            log("📈 Collected", collected.length, "/", num_workers, "results")
        }
        
        i = i + 1
    } then {
        ch.done()
        collected
    }
}

// Calculate statistics from results
let compute_stats = (results) -> {
    for i < results.length with i = 0, stats = {
        total: results.length,
        successful: 0,
        failed: 0,
        status_codes: [],
    } {
        let r = results[i]
        
        if r.success {
            stats.successful = stats.successful + 1
        } else {
            stats.failed = stats.failed + 1
        }
        
        stats.status_codes = stats.status_codes += [r.status]
        
        i = i + 1
    } then stats
}

log("🌐 KARL ULTIMATE SHOWCASE: Parallel HTTP Health Checker")
log("═══════════════════════════════════════════════════════")

log("\n⚙️  Configuration:")
log("Endpoints:", endpoints.length)
log("Parallel workers:", endpoints.length)

// Create rendezvous channel for coordination
let ch = rendezvous()

log("\n🚀 Launching parallel HTTP checks...")

// Spawn worker for each endpoint CONCURRENTLY
let worker_tasks = for i < endpoints.length with i = 0, tasks = [] {
    let endpoint = endpoints[i]
    let worker_num = i + 1
    let task = & check_endpoint(endpoint, worker_num, ch)
    tasks = tasks += [task]
    i = i + 1
} then tasks

// Spawn collector task
let collector_task = & collect_results(ch, endpoints.length)

// Wait for all workers
log("\n⏳ Executing parallel requests...")
for i < worker_tasks.length with i = 0 {
    wait worker_tasks[i]
    i = i + 1
} then {}

// Get final aggregated results
let results = wait collector_task

// Compute statistics
let stats = compute_stats(results)

log("\n═══════════════════════════════════════════════════════")
log("📈 RESULTS")
log("═══════════════════════════════════════════════════════")

log("\n📊 Statistics:")
log("Total requests:", stats.total)
log("✅ Successful:", stats.successful)
log("❌ Failed:", stats.failed)
log("Status codes:", stats.status_codes)

log("\n🔍 Detailed Results:")
for i < results.length with i = 0 {
    let r = results[i]
    let emoji = if r.success { "✅" } else { "❌" }
    log(emoji, "Worker", r.worker_id, "-", r.name, "→ HTTP", r.status)
    i = i + 1
} then {}

log("\n✨ All parallel requests completed!")
log("\n🎯 This showcase demonstrates Karl's COMPLETE feature set:")
log("  ✓ Concurrent task spawning with &")
log("  ✓ Task composition with .then()")
log("  ✓ Blocking wait for task completion")
log("  ✓ Rendezvous channels for coordination")
log("  ✓ Producer-consumer pattern")
log("  ✓ HTTP requests (real I/O!)")
log("  ✓ Error handling with ?")
log("  ✓ JSON parsing")
log("  ✓ Closures capturing variables")
log("  ✓ Higher-order functions")
log("  ✓ For loops with state management")
log("  ✓ Object destructuring")
log("  ✓ Pattern matching manifestations")
log("  ✓ Functional data transformations")
log("")
log("Karl = Elegant Concurrency + Functional Power! 🚀")

{
    statistics: stats,
    results: results,
}
