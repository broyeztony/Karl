// 🎯 Monte Carlo Pi Estimation - Parallel Computing Showcase
// Demonstrates Karl's concurrent programming power:
// - Multiple parallel worker tasks
// - Rendezvous channels for coordination
// - Producer-consumer pattern
// - Error handling with graceful degradation
// - Functional composition
// - Statistical analysis

// Generate random float between 0 and 1 (using simple PRNG)
let random_seed = 12345
let next_random = () -> {
    random_seed = (random_seed * 1103515245 + 12345) % 2147483648
    random_seed % 1000
}

// Check if a random point falls inside the unit circle
let point_in_circle = () -> {
    let x = next_random()
    let y = next_random()
    // Scale to 0-1 range and check if x² + y² <= 1
    // Using integer math: (x² + y²) <= 1000000
    x * x + y * y <= 1000000
}

// Monte Carlo worker: generates samples and sends results
let monte_carlo_worker = (ch, samples, worker_id) -> {
    log("🔧 Worker", worker_id, "starting with", samples, "samples")
    
    for i < samples with i = 0, inside = 0 {
        if point_in_circle() {
            inside = inside + 1
        }
        i = i + 1
    } then {
        let result = {
            worker_id: worker_id,
            total: samples,
            inside: inside,
        }
        ch.send(result)
        log("✅ Worker", worker_id, "completed:", inside, "inside out of", samples)
        result
    }
}

// Results aggregator: collects from all workers
let aggregate_results = (ch, num_workers) -> {
    log("📊 Aggregator waiting for", num_workers, "workers...")
    
    for i < num_workers with i = 0, total_samples = 0, total_inside = 0, results = [] {
        let [result, done] = ch.recv()
        
        if done {
            break { total: total_samples, inside: total_inside, results: results, }
        } else {
            results = results += [result]
            total_samples = total_samples + result.total
            total_inside = total_inside + result.inside
        }
        
        i = i + 1
    } then {
        ch.done()
        { total: total_samples, inside: total_inside, results: results, }
    }
}

// Estimate pi from total samples
let estimate_pi = (total, inside) -> {
    // Pi ≈ 4 * (inside / total)
    // Using integer math: scale up before division
    // (inside * 4000) / total gives us pi * 1000
    let numerator = inside * 4000
    // Manual division by finding how many times total fits into numerator
    for i < numerator with i = 0, quotient = 0 {
        if i * total >= numerator {
            break quotient
        }
        quotient = quotient + 1
        i = i + 1
    } then quotient
}

// Create histogram of worker contributions
let worker_histogram = (results) -> {
    results.map(r -> {
        worker: r.worker_id,
        samples: r.total,
        inside: r.inside,
        // Calculate percentage: (inside * 100) / total
        percentage: (r.inside * 100 * 10) % r.total,
    })
}

// Main execution
log("🎯 Monte Carlo Pi Estimation - Parallel Computing Demo")
log("═══════════════════════════════════════════════════════")

// Configuration
let num_workers = 5
let samples_per_worker = 2000

log("\n⚙️  Configuration:")
log("Workers:", num_workers)
log("Samples per worker:", samples_per_worker)
log("Total samples:", num_workers * samples_per_worker)

// Create communication channel
let ch = rendezvous()

log("\n🚀 Launching parallel workers...")

// Spawn all workers concurrently
let worker_tasks = for i < num_workers with i = 0, tasks = [] {
    let worker_num = i + 1
    let task = & monte_carlo_worker(ch, samples_per_worker, worker_num)
    tasks = tasks += [task]
    i = i + 1
} then tasks

// Spawn aggregator
let aggregator_task = & aggregate_results(ch, num_workers)

// Wait for all workers to complete
log("\n⏳ Computing...")
for i < worker_tasks.length with i = 0 {
    wait worker_tasks[i]
    i = i + 1
} then {}

// Get final results from aggregator
let final_results = wait aggregator_task

// Calculate pi estimate
let pi_estimate_raw = estimate_pi(final_results.total, final_results.inside)

log("\n═══════════════════════════════════════════════════════")
log("📈 RESULTS")
log("═══════════════════════════════════════════════════════")
log("\nTotal samples:", final_results.total)
log("Points inside circle:", final_results.inside)
log("Points outside circle:", final_results.total - final_results.inside)
log("Ratio (inside/total):", final_results.inside, "/", final_results.total)

// Calculate accuracy
let actual_pi_scaled = 3141  // Pi ≈ 3.141...
let error = if pi_estimate_raw > actual_pi_scaled {
    pi_estimate_raw - actual_pi_scaled
} else {
    actual_pi_scaled - pi_estimate_raw
}

log("\n🎯 Pi Estimate:", pi_estimate_raw, "(scaled by 1000)")
log("📏 Actual Pi:  ", actual_pi_scaled, "(scaled by 1000)")
log("📊 Error:      ", error, "(", error, "/ 1000 )")

log("\n👷 Worker Breakdown:")
let histogram = worker_histogram(final_results.results)
for i < histogram.length with i = 0 {
    let h = histogram[i]
    log("  Worker", h.worker, ":", h.inside, "/", h.samples, "inside")
    i = i + 1
} then {}

log("\n✨ Parallel computation completed successfully!")
log("This showcases Karl's:")
log("  • Concurrent task spawning with &")
log("  • Rendezvous channels for coordination")
log("  • Producer-consumer pattern")
log("  • Functional iteration and mapping")
log("  • Statistical computing")

{
    configuration: {
        workers: num_workers,
        samples_per_worker: samples_per_worker,
        total_samples: final_results.total,
    },
    results: {
        inside_circle: final_results.inside,
        outside_circle: final_results.total - final_results.inside,
        pi_estimate: pi_estimate_raw,
        actual_pi: actual_pi_scaled,
        error: error,
    },
    worker_breakdown: histogram,
}
