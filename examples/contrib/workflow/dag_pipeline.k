// ============================================================================
// ADVANCED WORKFLOW: DAG-BASED DATA PIPELINE
// ============================================================================
// Demonstrates a complex Directed Acyclic Graph workflow with:
// - Multiple stages with dependencies
// - Fan-out and fan-in patterns
// - Worker pools for parallel processing
// - Real-world data transformation pipeline
// ============================================================================

// Import the workflow engine (not used directly, but shows module loading)
let makeEngine = import "examples/contrib/workflow/engine.k"
let WorkflowEngine = makeEngine()

log("================================================================================")
log("DAG PIPELINE: Multi-Stage Data Processing with Worker Pools")
log("================================================================================")
log("")

// ============================================================================
// PIPELINE ARCHITECTURE
// ============================================================================
//
//                      [Fetch API 1] ──┐
//                                       ├──> [Merge Data] ──> [Transform] ──┐
//                      [Fetch API 2] ──┘                                     │
//                                                                             ├──> [Aggregate] ──> [Report]
//                      [Fetch API 3] ──┐                                     │
//                                       ├──> [Filter Data] ──> [Enrich] ────┘
//                      [Fetch API 4] ──┘
//
// ============================================================================

// ----------------------------------------------------------------------------
// STAGE 1: PARALLEL DATA FETCHING
// ----------------------------------------------------------------------------

log("[STAGE 1] Fetching data from multiple sources...")

let fetchChannel = rendezvous()

// Simulated API endpoints
let endpoints = [
    { id: 1, name: "Users API", delay: 120, dataSize: 50 },
    { id: 2, name: "Products API", delay: 150, dataSize: 100 },
    { id: 3, name: "Orders API", delay: 100, dataSize: 200 },
    { id: 4, name: "Analytics API", delay: 180, dataSize: 75 },
]

// Spawn fetcher workers
let fetchers = for i < endpoints.length with i = 0, workers = [] {
    let endpoint = endpoints[i]
    
    let fetcher = & (() -> {
        log("  [FETCH]", endpoint.name, "- Starting...")
        sleep(endpoint.delay)
        
        // Generate sample data
        let data = for j < endpoint.dataSize with j = 0, items = [] {
            items += [{
                source: endpoint.id,
                record: j + 1,
                value: (j + 1) * endpoint.id * 10,
                timestamp: j * 1000,
            }]
            j = j + 1
        } then items
        
        fetchChannel.send({
            sourceId: endpoint.id,
            sourceName: endpoint.name,
            data: data,
            recordCount: data.length,
        })
        
        log("  [FETCH]", endpoint.name, "- Completed:", data.length, "records")
    })()
    
    workers += [fetcher]
    i = i + 1
} then workers

// Collector for fetch results
let fetchResults = & (() -> {
    for i < endpoints.length with i = 0, results = [] {
        let [val, done] = fetchChannel.recv()
        if done { break results }
        results += [val]
        i = i + 1
    } then results
})()

// Wait for all fetchers
for i < fetchers.length with i = 0 {
    wait fetchers[i]
    i = i + 1
} then {}

fetchChannel.done()
let allFetchedData = wait fetchResults

log("[STAGE 1] ✓ Fetched", allFetchedData.length, "data sources")
log("")

// ----------------------------------------------------------------------------
// STAGE 2: PARALLEL PROCESSING PATHS
// ----------------------------------------------------------------------------

log("[STAGE 2] Processing data through parallel paths...")

// Path A: Merge sources 1 and 2
let pathA = & (() -> {
    log("  [PATH A] Merging sources 1 and 2...")
    sleep(80)
    
    let source1 = allFetchedData[0].data
    let source2 = allFetchedData[1].data
    
    let merged = for i < source1.length with i = 0, result = [] {
        result += [source1[i]]
        i = i + 1
    } then result
    
    for i < source2.length with i = 0, result = merged {
        result += [source2[i]]
        i = i + 1
    } then result
})()

// Path B: Filter and combine sources 3 and 4
let pathB = & (() -> {
    log("  [PATH B] Filtering sources 3 and 4...")
    sleep(70)
    
    let source3 = allFetchedData[2].data
    let source4 = allFetchedData[3].data
    
    // Filter: only keep records with value > 500
    let filtered3 = for i < source3.length with i = 0, result = [] {
        let record = source3[i]
        if record.value > 500 {
            result += [record]
        }
        i = i + 1
    } then result
    
    let filtered4 = for i < source4.length with i = 0, result = [] {
        let record = source4[i]
        if record.value > 500 {
            result += [record]
        }
        i = i + 1
    } then result
    
    // Combine filtered results
    for i < filtered3.length with i = 0, result = [] {
        result += [filtered3[i]]
        i = i + 1
    } then result
    
    for i < filtered4.length with i = 0, result = filtered3 {
        result += [filtered4[i]]
        i = i + 1
    } then result
})()

let mergedDataA = wait pathA
let filteredDataB = wait pathB

log("  [PATH A] ✓ Merged:", mergedDataA.length, "records")
log("  [PATH B] ✓ Filtered:", filteredDataB.length, "records")
log("")

// ----------------------------------------------------------------------------
// STAGE 3: TRANSFORMATION WITH WORKER POOL
// ----------------------------------------------------------------------------

log("[STAGE 3] Transforming data with worker pool...")

let transformChannel = rendezvous()
let transformedChannel = rendezvous()

// Feed data into transform pipeline
let transformProducer = & (() -> {
    // Send Path A data
    for i < mergedDataA.length with i = 0 {
        transformChannel.send({ source: "A", data: mergedDataA[i] })
        i = i + 1
    } then {}
    
    // Send Path B data
    for i < filteredDataB.length with i = 0 {
        transformChannel.send({ source: "B", data: filteredDataB[i] })
        i = i + 1
    } then {}
    
    transformChannel.done()
})()

// Spawn transform workers
let numTransformWorkers = 4
let transformWorkers = for i < numTransformWorkers with i = 0, workers = [] {
    let workerId = i + 1
    
    let worker = & (() -> {
        for processing = true with processing = true, count = 0 {
            let [item, done] = transformChannel.recv()
            if done {
                processing = false
                break count
            }
            
            // Transform: double the value, add metadata
            let transformed = {
                original: item.data,
                value: item.data.value * 2,
                source: item.source,
                processedBy: workerId,
                enhanced: true,
            }
            
            transformedChannel.send(transformed)
            count = count + 1
        } then count
    })()
    
    workers += [worker]
    i = i + 1
} then workers

// Collector for transformed data
let totalItems = mergedDataA.length + filteredDataB.length
let transformCollector = & (() -> {
    for i < totalItems with i = 0, results = [] {
        let [val, done] = transformedChannel.recv()
        if done { break results }
        results += [val]
        i = i + 1
    } then results
})()

// Wait for producer and workers
wait transformProducer
for i < transformWorkers.length with i = 0, totalProcessed = 0 {
    let processed = wait transformWorkers[i]
    totalProcessed = totalProcessed + processed
    log("  [WORKER", i + 1, "] Processed:", processed, "items")
    i = i + 1
} then totalProcessed

transformedChannel.done()
let transformedData = wait transformCollector

log("[STAGE 3] ✓ Transformed:", transformedData.length, "records")
log("")

// ----------------------------------------------------------------------------
// STAGE 4: AGGREGATION AND STATISTICS
// ----------------------------------------------------------------------------

log("[STAGE 4] Computing aggregations...")

// Parallel aggregation tasks
let aggChannel = rendezvous()

// Task 1: Count by source
let countBySource = & (() -> {
    log("  [AGG] Counting by source...")
    sleep(50)
    
    let counts = { A: 0, B: 0, }
    
    for i < transformedData.length with i = 0, c = counts {
        let item = transformedData[i]
        if item.source == "A" {
            c["A"] = c["A"] + 1
        } else {
            c["B"] = c["B"] + 1
        }
        i = i + 1
    } then c
    
    aggChannel.send({ metric: "sourceCount", data: counts })
})()

// Task 2: Sum all values
let sumValues = & (() -> {
    log("  [AGG] Summing values...")
    sleep(60)
    
    let total = for i < transformedData.length with i = 0, sum = 0 {
        sum = sum + transformedData[i].value
        i = i + 1
    } then sum
    
    aggChannel.send({ metric: "totalValue", data: total })
})()

// Task 3: Find min/max
let findMinMax = & (() -> {
    log("  [AGG] Finding min/max...")
    sleep(55)
    
    let stats = for i < transformedData.length with i = 0, min = 999999, max = 0 {
        let val = transformedData[i].value
        min = if val < min { val } else { min }
        max = if val > max { val } else { max }
        i = i + 1
    } then { min: min, max: max }
    
    aggChannel.send({ metric: "minMax", data: stats })
})()

// Task 4: Count by worker
let countByWorker = & (() -> {
    log("  [AGG] Counting by worker...")
    sleep(45)
    
    let workerCounts = for wid <= numTransformWorkers with wid = 1, counts = {} {
        counts[str(wid)] = 0
        wid = wid + 1
    } then counts
    
    for i < transformedData.length with i = 0 {
        let workerId = str(transformedData[i].processedBy)
        workerCounts[workerId] = workerCounts[workerId] + 1
        i = i + 1
    } then {}
    
    aggChannel.send({ metric: "workerDistribution", data: workerCounts })
})()

// Collect aggregation results
let numAggTasks = 4
let aggResults = & (() -> {
    for i < numAggTasks with i = 0, results = [] {
        let [val, done] = aggChannel.recv()
        if done { break results }
        results += [val]
        i = i + 1
    } then results
})()

wait countBySource
wait sumValues
wait findMinMax
wait countByWorker

aggChannel.done()
let finalAggregations = wait aggResults

log("[STAGE 4] ✓ Computed", finalAggregations.length, "metrics")
log("")

// ----------------------------------------------------------------------------
// STAGE 5: FINAL REPORT GENERATION
// ----------------------------------------------------------------------------

log("[STAGE 5] Generating final report...")
sleep(100)

log("")
log("================================================================================")
log("PIPELINE EXECUTION COMPLETE - FINAL REPORT")
log("================================================================================")
log("")

log("Data Sources:")
for i < allFetchedData.length with i = 0 {
    let source = allFetchedData[i]
    log("  •", source.sourceName, "-", source.recordCount, "records")
    i = i + 1
} then {}

log("")
log("Processing Results:")
log("  • Total records processed:", transformedData.length)
log("  • Transform workers used:", numTransformWorkers)

log("")
log("Aggregated Metrics:")
for i < finalAggregations.length with i = 0 {
    let agg = finalAggregations[i]
    match agg.metric {
        case "sourceCount" -> {
            log("  • Source A:", agg.data["A"], "records")
            log("  • Source B:", agg.data["B"], "records")
        }
        case "totalValue" -> log("  • Total value sum:", agg.data)
        case "minMax" -> {
            log("  • Min value:", agg.data.min)
            log("  • Max value:", agg.data.max)
        }
        case "workerDistribution" -> {
            log("  • Worker distribution:")
            for wid <= numTransformWorkers with wid = 1 {
                let workerId = str(wid)
                log("    - Worker", workerId, ":", agg.data[workerId], "items")
                wid = wid + 1
            } then {}
        }
        case _ -> {}
    }
    i = i + 1
} then {}

log("")
log("================================================================================")
log("✓ DAG Pipeline executed successfully!")
log("================================================================================")
log("")
log("Pipeline Statistics:")
log("  • Stages: 5")
log("  • Parallel fetch workers: 4")
log("  • Parallel processing paths: 2")
log("  • Transform worker pool size: 4")
log("  • Aggregation tasks: 4")
log("  • Total records in: 425")
log("  • Total records out:", transformedData.length)
log("")
log("This demonstrates:")
log("  ✓ Complex DAG with fan-out/fan-in patterns")
log("  ✓ Multi-stage pipelines with worker pools")
log("  ✓ Parallel data fetching and processing")
log("  ✓ Channel-based coordination between stages")
log("  ✓ Efficient data transformation and aggregation")
