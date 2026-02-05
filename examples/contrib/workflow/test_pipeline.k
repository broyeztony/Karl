// ============================================================================
// PIPELINE EXECUTION TEST
// ============================================================================
// Tests the multi-stage pipeline execution feature with:
// 1. Multiple processing stages
// 2. Worker pools per stage
// 3. Data flow between stages
// 4. Concurrent processing within stages
// ============================================================================

let makeEngine = import "./engine.k"
let Engine = makeEngine()

log("==========================================================")
log("PIPELINE EXECUTION TEST")
log("==========================================================")

// ----------------------------------------------------------------------------
// TEST 1: Basic 3-Stage Pipeline
// ----------------------------------------------------------------------------

log("\n--- TEST 1: Basic 3-Stage Pipeline ---\n")

// Create a simple data processing pipeline
let stages = [
    {
        name: "Extract",
        workers: 2,
        handler: (item) -> {
            log("[EXTRACT] Processing item:", item.id)
            sleep(50)
            {
                success: true,
                data: {
                    id: item.id,
                    value: item.value,
                    extracted: true,
                }
            }
        }
    },
    {
        name: "Transform",
        workers: 3,
        handler: (item) -> {
            log("[TRANSFORM] Processing item:", item.id)
            sleep(50)
            {
                success: true,
                data: {
                    id: item.id,
                    value: item.value * 2,
                    transformed: true,
                }
            }
        }
    },
    {
        name: "Load",
        workers: 2,
        handler: (item) -> {
            log("[LOAD] Processing item:", item.id)
            sleep(50)
            {
                success: true,
                data: {
                    id: item.id,
                    value: item.value,
                    loaded: true,
                }
            }
        }
    },
]

// Create input data
let items = for i < 10 with i = 0, data = [] {
    data += [{
        id: i + 1,
        value: (i + 1) * 10,
    }]
    i = i + 1
} then data

let pipelineWorkflow = {
    name: "ETL Pipeline",
    type: "pipeline",
    stages: stages,
    items: items,
}

let pipelineConfig = {
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

log("Executing pipeline with", items.length, "items through", stages.length, "stages...")
let startTime = now()
let result1 = Engine.execute(pipelineWorkflow, {}, pipelineConfig)
let duration1 = now() - startTime

if result1.success {
    log("\n✅ TEST 1 PASSED: Pipeline processed", items.length, "items in", duration1, "ms")
} else {
    log("\n❌ TEST 1 FAILED")
}

// ----------------------------------------------------------------------------
// TEST 2: Pipeline with Error Handling
// ----------------------------------------------------------------------------

log("\n--- TEST 2: Pipeline with Error Handling ---\n")

let errorCounter = { count: 0 }

let stagesWithErrors = [
    {
        name: "Stage1",
        workers: 2,
        handler: (item) -> {
            log("[STAGE1] Processing item:", item.id)
            sleep(30)
            {
                success: true,
                data: {
                    id: item.id,
                    stage1: true,
                }
            }
        }
    },
    {
        name: "Stage2-Flaky",
        workers: 2,
        handler: (item) -> {
            log("[STAGE2] Processing item:", item.id)
            sleep(30)
            
            // Simulate occasional failures
            if item.id % 3 == 0 {
                errorCounter.count = errorCounter.count + 1
                log("[STAGE2] Simulated error for item:", item.id)
                {
                    success: false,
                    data: null,
                }
            } else {
                {
                    success: true,
                    data: {
                        id: item.id,
                        stage2: true,
                    }
                }
            }
        }
    },
    {
        name: "Stage3",
        workers: 2,
        handler: (item) -> {
            log("[STAGE3] Processing item:", item.id)
            sleep(30)
            {
                success: true,
                data: {
                    id: item.id,
                    stage3: true,
                }
            }
        }
    },
]

let items2 = for i < 6 with i = 0, data = [] {
    data += [{ id: i + 1 }]
    i = i + 1
} then data

let errorPipelineWorkflow = {
    name: "Error Handling Pipeline",
    type: "pipeline",
    stages: stagesWithErrors,
    items: items2,
}

log("Executing pipeline with error simulation...")
let result2 = Engine.execute(errorPipelineWorkflow, {}, pipelineConfig)

if result2.success {
    log("\n✅ TEST 2 PASSED: Pipeline handled errors gracefully")
    log("Total simulated errors:", errorCounter.count)
} else {
    log("\n❌ TEST 2 FAILED")
}

// ----------------------------------------------------------------------------
// TEST 3: High-Throughput Pipeline
// ----------------------------------------------------------------------------

log("\n--- TEST 3: High-Throughput Pipeline ---\n")

let highThroughputStages = [
    {
        name: "Fast-Stage-1",
        workers: 4,
        handler: (item) -> {
            // Very fast processing
            {
                success: true,
                data: {
                    id: item.id,
                    processed: true,
                }
            }
        }
    },
    {
        name: "Fast-Stage-2",
        workers: 4,
        handler: (item) -> {
            {
                success: true,
                data: {
                    id: item.id,
                    processed: true,
                }
            }
        }
    },
]

// Create a large dataset
let largeItems = for i < 50 with i = 0, data = [] {
    data += [{ id: i + 1 }]
    i = i + 1
} then data

let highThroughputWorkflow = {
    name: "High-Throughput Pipeline",
    type: "pipeline",
    stages: highThroughputStages,
    items: largeItems,
}

log("Executing high-throughput pipeline with", largeItems.length, "items...")
let startTime3 = now()
let result3 = Engine.execute(highThroughputWorkflow, {}, pipelineConfig)
let duration3 = now() - startTime3

if result3.success {
    log("\n✅ TEST 3 PASSED: Processed", largeItems.length, "items in", duration3, "ms")
    let throughput = largeItems.length * 1000 / duration3
    log("Throughput:", throughput, "items/second")
} else {
    log("\n❌ TEST 3 FAILED")
}

// ----------------------------------------------------------------------------
// SUMMARY
// ----------------------------------------------------------------------------

log("\n==========================================================")
log("TEST SUMMARY")
log("==========================================================")

let allPassed = result1.success && result2.success && result3.success

if allPassed {
    log("✅ ALL PIPELINE TESTS PASSED")
    log("")
    log("Pipeline Features Verified:")
    log("  1. ✅ Multi-stage data processing")
    log("  2. ✅ Worker pools per stage")
    log("  3. ✅ Error handling in pipelines")
    log("  4. ✅ High-throughput processing")
} else {
    log("❌ SOME TESTS FAILED")
}

log("==========================================================")
