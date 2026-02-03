// Test DAG without persistence

let makeEngine = import "examples/contrib/workflow/engine.k"
let Engine = makeEngine()

log("Testing DAG without persistence...")

let dagNodes = [
    {
        id: "a",
        name: "Task A",
        handler: (ctx) -> {
            log("[A] Running")
            { success: true, data: { result: "A" } }
        }
    },
    {
        id: "b",
        name: "Task B",
        handler: (ctx) -> {
            log("[B] Running")
            { success: true, data: { result: "B" } }
        }
    },
]

let dagEdges = [
    { source: "a", target: "b" },
]

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

let workflow = {
    name: "Simple DAG",
    type: "dag",
    nodes: dagNodes,
    edges: dagEdges,
}

log("Executing DAG...")
let result = Engine.execute(workflow, {}, config)

log("Result:", result)

if result.success {
    log("✅ Test passed")
} else {
    log("❌ Test failed")
}
