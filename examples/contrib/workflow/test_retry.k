// Simple retry test

let makeEngine = import "./engine.k"
let Engine = makeEngine()

log("Testing retry policy...")

let task = {
    name: "TestTask",
    handler: (ctx) -> {
        log("Task executing")
        { success: true, data: { result: "done" } }
    },
}

let config = {
    defaultRetries: 2,
    defaultWorkers: 3,
    stopOnError: false,
    retryPolicy: {
        maxAttempts: 3,
        strategy: "exponential",
        initialDelay: 100,
        maxDelay: 5000,
        jitterEnabled: false,
        jitterFactor: 0.1,
        retryableErrors: [],
    },
    useWorkerPool: false,
    workerCount: 1,
    queueSize: 100,
    batchSize: 10,
    enablePriority: false,
    shutdownTimeout: 30000,
    enableMetrics: false,
    enablePersistence: false,
    workflowId: "test-1",
    storageConfig: {
        storageDir: "./workflow-state",
        enableAutoCheckpoint: false,
    },
}

let workflow = {
    name: "Simple Test",
    type: "sequential",
    tasks: [task],
}

log("Executing workflow...")
let result = Engine.execute(workflow, {}, config)

if result.success {
    log("✅ Test passed!")
} else {
    log("❌ Test failed:", result.error)
}
