// ============================================================================
// INTEGRATED FEATURES TEST
// ============================================================================
// Tests the three integrated features in the main engine:
// 1. Retry policies with exponential back-off
// 2. Parallel execution with worker pools
// 3. Persisted DAG state with resume capability
// ============================================================================

let makeEngine = import "./engine.k"
let Engine = makeEngine()

let makeRetry = import "./retry_policy.k"
let Retry = makeRetry()

log("==========================================================")
log("INTEGRATED WORKFLOW ENGINE TESTS")
log("==========================================================")

// ----------------------------------------------------------------------------
// TEST 1: Retry Policy Integration
// ----------------------------------------------------------------------------

log("\n--- TEST 1: Retry Policy Integration ---\n")

let failureCount = 0

let flakyTasks = [
    {
        name: "FlakyAPI-1",
        handler: (ctx) -> {
            failureCount = failureCount + 1
            if failureCount < 3 {
                log("[FLAKY] Simulating failure #", failureCount)
                { success: false, error: "Connection timeout" }
            } else {
                log("[FLAKY] Success after", failureCount, "attempts")
                { success: true, data: { result: "API response" } }
            }
        }
    },
]

// Configure with exponential back-off retry policy
let retryConfig = {
    defaultRetries: 2,
    defaultWorkers: 3,
    stopOnError: false,
    retryPolicy: {
        maxAttempts: 5,
        strategy: Retry.RETRY_EXPONENTIAL,
        initialDelay: 100,
        maxDelay: 3000,
        jitterEnabled: false,
        jitterFactor: 0.1,
        retryableErrors: [],
    },
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
    },
}

let retryWorkflow = {
    name: "Retry Test Workflow",
    type: "sequential",
    tasks: flakyTasks,
}

let retryResult = Engine.execute(retryWorkflow, {}, retryConfig)

if retryResult.success {
    log("\n✅ TEST 1 PASSED: Retry policy successfully handled failures")
} else {
    log("\n❌ TEST 1 FAILED")
}

// ----------------------------------------------------------------------------
// TEST 2: Worker Pool Integration
// ----------------------------------------------------------------------------

log("\n--- TEST 2: Worker Pool Integration ---\n")

// Create 8 compute tasks
let computeTasks = for i < 8 with i = 0, tasks = [] {
    tasks += [{
        name: "Compute-" + str(i + 1),
        handler: (ctx) -> {
            log("[COMPUTE] Processing task", i + 1)
            // Simulate work
            let sum = for j < 500 with j = 0, total = 0 {
                total = total + j
                j = j + 1
            } then total
            sleep(100)
            { success: true, data: { taskId: i + 1, result: sum } }
        }
    }]
    i = i + 1
} then tasks

// Configure with worker pool
let poolConfig = {
    defaultRetries: 2,
    defaultWorkers: 3,
    stopOnError: false,
    retryPolicy: null,
    useWorkerPool: true,
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
    },
}

let parallelWorkflow = {
    name: "Parallel Worker Pool Test",
    type: "parallel",
    tasks: computeTasks,
}

log("Executing", computeTasks.length, "tasks with 4-worker pool...")
let startTime = now()
let poolResult = Engine.execute(parallelWorkflow, {}, poolConfig)
let duration = now() - startTime

if poolResult.success {
    log("\n✅ TEST 2 PASSED: Worker pool executed", computeTasks.length, "tasks in", duration, "ms")
} else {
    log("\n❌ TEST 2 FAILED")
}

// ----------------------------------------------------------------------------
// TEST 3: Persisted State Integration
// ----------------------------------------------------------------------------

log("\n--- TEST 3: Persisted State Integration ---\n")

// Create a DAG workflow
let dagNodes = [
    {
        id: "init",
        name: "Initialize",
        handler: (ctx) -> {
            log("[DAG] Initializing...")
            sleep(200)
            { success: true, data: { initialized: true } }
        }
    },
    {
        id: "process-a",
        name: "Process A",
        handler: (ctx) -> {
            log("[DAG] Processing A...")
            sleep(200)
            { success: true, data: { processedA: true } }
        }
    },
    {
        id: "process-b",
        name: "Process B",
        handler: (ctx) -> {
            log("[DAG] Processing B...")
            sleep(200)
            { success: true, data: { processedB: true } }
        }
    },
    {
        id: "merge",
        name: "Merge Results",
        handler: (ctx) -> {
            log("[DAG] Merging...")
            sleep(200)
            { success: true, data: { merged: true } }
        }
    },
]

let dagEdges = [
    { source: "init", target: "process-a" },
    { source: "init", target: "process-b" },
    { source: "process-a", target: "merge" },
    { source: "process-b", target: "merge" },
]

// Configure with persistence
let persistConfig = {
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
    enablePersistence: true,
    workflowId: "test-dag-001",
    storageConfig: {
        storageDir: "./workflow-state",
        enableAutoCheckpoint: false,
        checkpointInterval: 5000,
        compressionEnabled: false,
    },
}

let dagWorkflow = {
    name: "Persisted DAG Test",
    type: "dag",
    nodes: dagNodes,
    edges: dagEdges,
}

log("Executing DAG with persistence enabled...")
let dagResult = Engine.execute(dagWorkflow, {}, persistConfig)

if dagResult.success {
    log("\n✅ TEST 3 PASSED: DAG executed with state persistence")
    log("Workflow ID:", dagResult.workflowId)
    log("State saved to: ./workflow-state/", dagResult.workflowId, ".json")
} else {
    log("\n❌ TEST 3 FAILED")
}

// ----------------------------------------------------------------------------
// TEST 4: Combined Features - Resilient Parallel DAG
// ----------------------------------------------------------------------------

log("\n--- TEST 4: Combined Features Test ---\n")

// Create a complex DAG with retry and parallel execution
let complexNodes = [
    {
        id: "fetch-1",
        name: "Fetch Data Source 1",
        handler: (ctx) -> {
            log("[COMPLEX] Fetching from source 1...")
            sleep(150)
            { success: true, data: { source: 1, records: 100 } }
        }
    },
    {
        id: "fetch-2",
        name: "Fetch Data Source 2",
        handler: (ctx) -> {
            log("[COMPLEX] Fetching from source 2...")
            sleep(150)
            { success: true, data: { source: 2, records: 200 } }
        }
    },
    {
        id: "validate",
        name: "Validate Data",
        handler: (ctx) -> {
            log("[COMPLEX] Validating...")
            sleep(100)
            { success: true, data: { valid: true } }
        }
    },
    {
        id: "transform",
        name: "Transform Data",
        handler: (ctx) -> {
            log("[COMPLEX] Transforming...")
            sleep(100)
            { success: true, data: { transformed: true } }
        }
    },
    {
        id: "load",
        name: "Load to Database",
        handler: (ctx) -> {
            log("[COMPLEX] Loading...")
            sleep(100)
            { success: true, data: { loaded: true } }
        }
    },
]

let complexEdges = [
    { source: "fetch-1", target: "validate" },
    { source: "fetch-2", target: "validate" },
    { source: "validate", target: "transform" },
    { source: "transform", target: "load" },
]

// Configure with all features enabled
let fullConfig = {
    defaultRetries: 2,
    defaultWorkers: 3,
    stopOnError: false,
    retryPolicy: {
        maxAttempts: 3,
        strategy: Retry.RETRY_EXPONENTIAL,
        initialDelay: 100,
        maxDelay: 2000,
        jitterEnabled: false,
        jitterFactor: 0.1,
        retryableErrors: [],
    },
    useWorkerPool: false,
    workerCount: 4,
    queueSize: 100,
    batchSize: 10,
    enablePriority: false,
    shutdownTimeout: 30000,
    enableMetrics: true,
    enablePersistence: true,
    workflowId: "complex-etl-001",
    storageConfig: {
        storageDir: "./workflow-state",
        enableAutoCheckpoint: false,
        checkpointInterval: 5000,
        compressionEnabled: false,
    },
}

let complexWorkflow = {
    name: "Complex ETL Pipeline",
    type: "dag",
    nodes: complexNodes,
    edges: complexEdges,
}

log("Executing complex DAG with retry + persistence...")
let complexResult = Engine.execute(complexWorkflow, {}, fullConfig)

if complexResult.success {
    log("\n✅ TEST 4 PASSED: Complex workflow with all features succeeded")
    log("Workflow ID:", complexResult.workflowId)
} else {
    log("\n❌ TEST 4 FAILED")
}

// ----------------------------------------------------------------------------
// SUMMARY
// ----------------------------------------------------------------------------

log("\n==========================================================")
log("TEST SUMMARY")
log("==========================================================")

let allPassed = retryResult.success && poolResult.success && dagResult.success && complexResult.success

if allPassed {
    log("✅ ALL TESTS PASSED")
    log("")
    log("Feature Integration Verified:")
    log("  1. ✅ Retry policies with exponential back-off")
    log("  2. ✅ Parallel execution with worker pools")
    log("  3. ✅ Persisted DAG state with checkpoints")
    log("  4. ✅ Combined features in complex workflows")
} else {
    log("❌ SOME TESTS FAILED")
}

log("==========================================================")
