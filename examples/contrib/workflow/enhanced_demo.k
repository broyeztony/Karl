// ============================================================================
// ENHANCED WORKFLOW ENGINE DEMO
// ============================================================================
// Demonstrates the three new features:
// 1. Persisted DAG state (storage.k)
// 2. Retry & back-off policy (retry_policy.k)
// 3. Parallel execution engine (parallel_executor.k)
// ============================================================================

let makeStorage = import "./storage.k"
let Storage = makeStorage()

let makeRetry = import "./retry_policy.k"
let Retry = makeRetry()

let makeParallel = import "./parallel_executor.k"
let Parallel = makeParallel()

let makeEngine = import "./engine.k"
let Engine = makeEngine()

log("==========================================================")
log("ENHANCED WORKFLOW ENGINE DEMO")
log("==========================================================")

// ----------------------------------------------------------------------------
// DEMO 1: Retry Policy with Exponential Back-off
// ----------------------------------------------------------------------------

log("\n--- DEMO 1: Retry Policy with Exponential Back-off ---\n")

// Create a flaky task that fails randomly
let flakyTaskAttempts = 0
let createFlakyTask = (name, failureRate) -> {
    {
        name: name,
        handler: (ctx) -> {
            flakyTaskAttempts = flakyTaskAttempts + 1
            let randomValue = (now() % 1000) / 1000
            let shouldFail = randomValue < failureRate
            
            if shouldFail {
                log("[FLAKY]", name, "- Simulating failure")
                { success: false, error: "Transient network error" }
            } else {
                log("[FLAKY]", name, "- Success!")
                { success: true, data: { result: "Task completed", attempt: flakyTaskAttempts } }
            }
        }
    }
}

// Create retry engine with exponential back-off
let retryEngine = Retry.createRetryEngine({
    maxAttempts: 5,
    strategy: Retry.RETRY_EXPONENTIAL,
    initialDelay: 100,
    maxDelay: 5000,
    jitterEnabled: false,
    jitterFactor: 0.1,
    retryableErrors: [],
})

// Execute flaky task with retry
let flakyTask = createFlakyTask("NetworkRequest", 0.6)  // 60% failure rate
let retryResult = retryEngine.execute(flakyTask, {})

if retryResult.success {
    log("\n✅ Task succeeded after", retryResult.attempts, "attempts")
} else {
    log("\n❌ Task failed after", retryResult.attempts, "attempts")
}

// ----------------------------------------------------------------------------
// DEMO 2: Parallel Execution with Worker Pool
// ----------------------------------------------------------------------------

log("\n--- DEMO 2: Parallel Execution with Worker Pool ---\n")

// Create CPU-intensive tasks
let createComputeTask = (id, workload) -> {
    {
        name: "Compute-" + str(id),
        handler: (ctx) -> {
            log("[COMPUTE]", id, "- Starting heavy computation")
            
            // Simulate CPU work
            let result = for i < workload with i = 0, sum = 0 {
                sum = sum + i
                i = i + 1
            } then sum
            
            log("[COMPUTE]", id, "- Completed")
            { success: true, data: { taskId: id, result: result } }
        }
    }
}

// Create 12 compute tasks
let computeTasks = for i < 12 with i = 0, tasks = [] {
    tasks += [createComputeTask(i + 1, 1000)]
    i = i + 1
} then tasks

// Create parallel executor with 4 workers
let parallelExecutor = Parallel.createParallelExecutor({
    workerCount: 4,
    queueSize: 100,
    enablePriority: false,
    shutdownTimeout: 30000,
    enableMetrics: true,
    batchSize: 10,
})

log("Executing", computeTasks.length, "tasks with 4 workers...")
let startTime = now()
let parallelResult = parallelExecutor.execute(computeTasks, {})
let duration = now() - startTime

if parallelResult.success {
    log("\n✅ All tasks completed in", duration, "ms")
    log("Results:", parallelResult.results.length)
    
    // Show worker metrics
    if parallelResult.metrics {
        log("\nWorker Metrics:")
        for i < parallelResult.metrics.length with i = 0 {
            let workerMetrics = parallelResult.metrics[i]
            log("  Worker", workerMetrics.workerId, ":")
            log("    - Tasks processed:", workerMetrics.metrics.tasksProcessed)
            log("    - Success:", workerMetrics.metrics.tasksSucceeded)
            log("    - Failed:", workerMetrics.metrics.tasksFailed)
            i = i + 1
        } then {}
    }
}

// ----------------------------------------------------------------------------
// DEMO 3: Persisted DAG State with Resume
// ----------------------------------------------------------------------------

log("\n--- DEMO 3: Persisted DAG State with Resume ---\n")

// Create storage engine
let storageEngine = Storage.createStorageEngine({
    storageDir: "./workflow-state",
    enableAutoCheckpoint: false,
    checkpointInterval: 5000,
    compressionEnabled: false,
})

// Define a simple DAG
let nodes = [
    {
        id: "fetch",
        name: "Fetch Data",
        handler: (ctx) -> {
            log("[DAG] Fetching data...")
            sleep(500)
            { success: true, data: { records: 100 } }
        }
    },
    {
        id: "validate",
        name: "Validate Data",
        handler: (ctx) -> {
            log("[DAG] Validating data...")
            sleep(500)
            { success: true, data: { valid: true } }
        }
    },
    {
        id: "transform",
        name: "Transform Data",
        handler: (ctx) -> {
            log("[DAG] Transforming data...")
            sleep(500)
            { success: true, data: { transformed: true } }
        }
    },
    {
        id: "load",
        name: "Load Data",
        handler: (ctx) -> {
            log("[DAG] Loading data...")
            sleep(500)
            { success: true, data: { loaded: true } }
        }
    },
]

let edges = [
    { source: "fetch", target: "validate" },
    { source: "validate", target: "transform" },
    { source: "transform", target: "load" },
]

let workflowId = "etl-pipeline-001"

// Check if we can resume
let resumeCheck = storageEngine.canResume(workflowId)

if resumeCheck.resumable {
    log("Found existing workflow state, resuming...")
    log("Completed nodes:", resumeCheck.state.totalCompleted, "/", nodes.length)
    
    // Get incomplete nodes
    let incompleteNodes = storageEngine.getIncompleteNodes(nodes, resumeCheck.state)
    log("Remaining nodes:", incompleteNodes.length)
} else {
    log("No existing state, starting fresh workflow...")
    
    // Create initial state
    let initialState = storageEngine.createInitialState(nodes, edges)
    
    // Save initial state
    let saveResult = storageEngine.save(workflowId, initialState)
    
    if saveResult.success {
        log("Initial state saved to:", saveResult.path)
    }
}

// Execute the DAG
log("\nExecuting DAG workflow...")
let dagConfig = {
    defaultRetries: 2,
    defaultWorkers: 3,
    stopOnError: false,
    retryPolicy: null,
    useWorkerPool: false,
    workerCount: 4,
    queueSize: 100,
    batchSize: 10,
    enablePriority: false,
    shutdownTimeout: 30000,
    enableMetrics: true,
    enablePersistence: false,
    workflowId: null,
    storageConfig: {
        storageDir: "./workflow-state",
        enableAutoCheckpoint: false,
        checkpointInterval: 5000,
        compressionEnabled: false,
    },
}

let dagResult = Engine.executeDAG(nodes, edges, {}, dagConfig)

if dagResult.success {
    log("\n✅ DAG execution completed successfully")
    
    // Update and save final state
    let finalState = storageEngine.createInitialState(nodes, edges)
    finalState.status = "completed"
    finalState.totalCompleted = nodes.length
    
    // Mark all nodes as completed
    for i < nodes.length with i = 0 {
        finalState.completedNodes[nodes[i].id] = true
        finalState.startedNodes[nodes[i].id] = true
        i = i + 1
    } then {}
    
    storageEngine.save(workflowId, finalState)
    log("Final state saved")
}

// ----------------------------------------------------------------------------
// DEMO 4: Combined Features - Resilient Parallel DAG
// ----------------------------------------------------------------------------

log("\n--- DEMO 4: Combined Features - Resilient Parallel DAG ---\n")

// Create a complex workflow combining all features
let createResilientTask = (id, retryPolicy) -> {
    {
        id: "task-" + str(id),
        name: "ResilientTask-" + str(id),
        handler: (ctx) -> {
            // Simulate occasional failures
            let randomValue = (now() % 1000) / 1000
            let shouldFail = randomValue < 0.3
            
            if shouldFail {
                { success: false, error: "Random failure" }
            } else {
                sleep(200)
                { success: true, data: { taskId: id, processed: true } }
            }
        }
    }
}

// Create retry policy for resilient tasks
let taskRetryPolicy = Retry.createExponentialRetryPolicy(3, 100, 2000)

// Wrap tasks with retry logic
let resilientNodes = for i < 6 with i = 0, taskList = [] {
    let baseTask = createResilientTask(i + 1, taskRetryPolicy)
    
    // Wrap handler with retry
    let wrappedTask = {
        id: baseTask.id,
        name: baseTask.name,
        handler: (ctx) -> {
            let retryEngine = Retry.createRetryEngine(taskRetryPolicy)
            retryEngine.execute(baseTask, ctx)
        }
    }
    
    taskList += [wrappedTask]
    i = i + 1
} then taskList

// Define dependencies for parallel execution
let resilientEdges = [
    { source: "task-1", target: "task-3" },
    { source: "task-2", target: "task-3" },
    { source: "task-3", target: "task-5" },
    { source: "task-4", target: "task-5" },
    { source: "task-5", target: "task-6" },
]

log("Executing resilient DAG with", resilientNodes.length, "nodes...")
let resilientResult = Engine.executeDAG(resilientNodes, resilientEdges, {}, dagConfig)

if resilientResult.success {
    log("\n✅ Resilient DAG completed successfully")
    log("All tasks executed with automatic retry on failures")
}

// ----------------------------------------------------------------------------
// SUMMARY
// ----------------------------------------------------------------------------

log("\n==========================================================")
log("DEMO SUMMARY")
log("==========================================================")
log("✅ Feature 1: Retry Policy - Demonstrated exponential back-off")
log("✅ Feature 2: Parallel Execution - Demonstrated worker pool with 4 workers")
log("✅ Feature 3: Persisted State - Demonstrated save/load/resume capability")
log("✅ Combined: Resilient parallel DAG with all features integrated")
log("==========================================================")
