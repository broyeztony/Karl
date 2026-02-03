// Simple retry test

import "examples/contrib/workflow/engine.k"

log("Testing retry policy...")

let task = {
    name: "TestTask",
    execute: () -> {
        log("Task executing")
        { success: true, result: "done" }
    },
}

let config = {
    retryPolicy: {
        maxAttempts: 3,
        backoffMs: 100,
        retryableErrors: [],
    },
    useWorkerPool: false,
    workerCount: 1,
    enableMetrics: false,
    enablePersistence: false,
    workflowId: "test-1",
    storageConfig: null,
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
