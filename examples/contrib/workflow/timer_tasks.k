// ============================================================================
// TIMER TASKS DEMO - Scheduled and Interval-Based Workflows
// ============================================================================
// Demonstrates the timer task features of the workflow engine:
// - Delayed task execution
// - Interval-based repeated execution
// - Coordinated timer-based workflows
// - Real-world scheduling scenarios
// ============================================================================

let makeEngine = import "examples/contrib/workflow/engine.k"
let WorkflowEngine = makeEngine()

log("================================================================================")
log("TIMER TASKS: Scheduled Execution and Interval Processing")
log("================================================================================")
log("")

// ============================================================================
// EXAMPLE 1: Simple Delayed Task
// ============================================================================

log("[EXAMPLE 1] Simple delayed task execution")
log("")

let delayedTask = {
    name: "Delayed Notification",
    type: "timer",
    task: {
        name: "Send Notification",
        type: "timer",
        delay: 500,
        handler: (context) -> {
            log("  📧 Sending notification...")
            sleep(100)
            log("  ✓ Notification sent successfully!")
            {
                success: true,
                data: {
                    message: "Notification delivered",
                    timestamp: 1234567890,
                }
            }
        }
    }
}

let result1 = WorkflowEngine.execute(delayedTask, {}, WorkflowEngine.defaultConfig)
log("Result:", if result1.success { "✓ Success" } else { "✗ Failed" })
log("")

// ============================================================================
// EXAMPLE 2: Interval Task - Health Check Monitor
// ============================================================================

log("[EXAMPLE 2] Interval-based health check monitoring")
log("")

let healthCheckTask = {
    name: "Health Check Monitor",
    type: "interval",
    task: {
        name: "Check System Health",
        type: "interval",
        interval: 300,
        maxRepetitions: 5,
        handler: (ctx) -> {
            let iteration = ctx.iteration
            log("  🏥 Health check #", iteration)
            
            // Simulate health check
            sleep(50)
            
            // Randomly succeed/fail for demo
            let healthy = iteration != 3  // Fail on iteration 3
            
            if healthy {
                log("  ✓ System healthy")
                {
                    success: true,
                    data: {
                        status: "healthy",
                        iteration: iteration,
                        cpuUsage: 45 + iteration * 5,
                        memoryUsage: 60 + iteration * 3,
                    }
                }
            } else {
                log("  ⚠️  System unhealthy - high memory usage")
                {
                    success: false,
                    data: {
                        status: "unhealthy",
                        iteration: iteration,
                        error: "Memory threshold exceeded"
                    }
                }
            }
        }
    }
}

let result2 = WorkflowEngine.execute(healthCheckTask, {}, WorkflowEngine.defaultConfig)
log("Overall result:", if result2.success { "✓ All checks passed" } else { "⚠️  Some checks failed" })
log("Total iterations:", result2.iterations)
log("")

// ============================================================================
// EXAMPLE 3: Coordinated Timer Workflow - Batch Processing
// ============================================================================

log("[EXAMPLE 3] Coordinated batch processing with timers")
log("")

// Simulate a batch processing system that:
// 1. Waits for data to accumulate (timer)
// 2. Processes batches at intervals
// 3. Sends periodic status updates

let batchChannel = rendezvous()
let statusChannel = rendezvous()

// Data accumulation phase
let dataAccumulator = & (() -> {
    log("  📊 Starting data accumulation phase...")
    sleep(400)
    
    let batchData = for i < 10 with i = 0, items = [] {
        items += [{
            id: i + 1,
            value: (i + 1) * 100,
            timestamp: i * 50,
        }]
        i = i + 1
    } then items
    
    batchChannel.send(batchData)
    log("  ✓ Accumulated", batchData.length, "items for processing")
})()

// Batch processor (interval-based)
let batchProcessor = & (() -> {
    // Wait for data
    let [data, done] = batchChannel.recv()
    if done {
        { success: false, error: "No data received" }
    } else {
        log("  🔄 Starting batch processing...")
        
        // Process in batches
        let batchSize = 3
        let numBatches = (data.length + batchSize - 1) / batchSize
        
        for batchNum < numBatches with batchNum = 0, processed = 0 {
            let startIdx = batchNum * batchSize
            let endIdx = if (startIdx + batchSize) < data.length {
                startIdx + batchSize
            } else {
                data.length
            }
            
            log("  📦 Processing batch", batchNum + 1, "/", numBatches, "- items", startIdx + 1, "to", endIdx)
            
            // Process batch
            sleep(200)
            
            let batchProcessed = endIdx - startIdx
            processed = processed + batchProcessed
            
            // Send status update
            statusChannel.send({
                batch: batchNum + 1,
                totalBatches: numBatches,
                processed: processed,
                total: data.length,
                progress: (processed * 100) / data.length,
            })
            
            // Delay between batches
            if batchNum < numBatches - 1 {
                sleep(150)
            }
            
            batchNum = batchNum + 1
        } then {
            success: true,
            processedCount: processed
        }
    }
})()

// Status monitor (interval consumer)
let statusMonitor = & (() -> {
    log("  📈 Status monitor active")
    
    for monitoring = true with monitoring = true, updates = 0 {
        let [status, done] = statusChannel.recv()
        if done {
            break updates
        }
        
        log("  ├─ Progress:", status.progress, "% - Batch", status.batch, "/", status.totalBatches)
        updates = updates + 1
    } then updates
})()

// Wait for completion
wait dataAccumulator
let processingResult = wait batchProcessor
statusChannel.done()
let statusUpdates = wait statusMonitor

log("  ✓ Batch processing completed")
log("  → Processed:", processingResult.processedCount, "items")
log("  → Status updates sent:", statusUpdates)
log("")

// ============================================================================
// EXAMPLE 4: Multi-Timer Coordination - Distributed Task Scheduler
// ============================================================================

log("[EXAMPLE 4] Multi-timer coordination for distributed tasks")
log("")

let taskResults = rendezvous()

// Task 1: Quick task (short delay)
let quickTask = & (() -> {
    log("  ⏱️  Quick task starting (250ms delay)")
    sleep(250)
    log("  ✓ Quick task completed")
    taskResults.send({ task: "quick", duration: 250, result: "success" })
})()

// Task 2: Medium task
let mediumTask = & (() -> {
    log("  ⏱️  Medium task starting (500ms delay)")
    sleep(500)
    log("  ✓ Medium task completed")
    taskResults.send({ task: "medium", duration: 500, result: "success" })
})()

// Task 3: Slow task
let slowTask = & (() -> {
    log("  ⏱️  Slow task starting (750ms delay)")
    sleep(750)
    log("  ✓ Slow task completed")
    taskResults.send({ task: "slow", duration: 750, result: "success" })
})()

// Coordinator
let coordinator = & (() -> {
    log("  🎯 Coordinator waiting for all tasks...")
    
    for i < 3 with i = 0, results = [] {
        let [result, done] = taskResults.recv()
        if done { break results }
        
        results += [result]
        log("  ├─ Received:", result.task, "task result")
        
        i = i + 1
    } then results
})()

// Wait for all tasks
wait quickTask
wait mediumTask
wait slowTask

taskResults.done()
let allResults = wait coordinator

log("  ✓ All", allResults.length, "tasks completed")
log("")

// ============================================================================
// SUMMARY
// ============================================================================

log("================================================================================")
log("✓ Timer Tasks Demo Complete")
log("================================================================================")
log("")
log("Features demonstrated:")
log("  ✓ Simple delayed task execution")
log("  ✓ Interval-based repeated execution")
log("  ✓ Coordinated batch processing with timers")
log("  ✓ Multi-timer task coordination")
log("  ✓ Health check monitoring pattern")
log("  ✓ Batch processing with status updates")
log("")
log("Use cases:")
log("  • Scheduled notifications and alerts")
log("  • Periodic health checks and monitoring")
log("  • Batch processing with configurable intervals")
log("  • Distributed task coordination")
log("  • Rate-limited API calls")
log("  • Polling and retry mechanisms")
log("")
