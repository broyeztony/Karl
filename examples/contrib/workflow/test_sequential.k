// Minimal test to debug the boolean issue

let makeEngine = import "./engine.k"
let Engine = makeEngine()

log("Testing basic retry...")

let testTask = {
    name: "SimpleTest",
    handler: (ctx) -> {
        log("Handler executing")
        { success: true, data: { result: "OK" } }
    }
}

let workflow = {
    name: "Simple Test",
    type: "sequential",
    tasks: [testTask]
}

let config = {
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
    },
}

log("Executing workflow...")
let result = Engine.execute(workflow, {}, config)

log("Result:", result)

if result.success {
    log("✅ Test passed")
} else {
    log("❌ Test failed")
}
