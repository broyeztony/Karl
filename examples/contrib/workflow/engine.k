// ============================================================================
// WORKFLOW ENGINE - Task-based Execution with DAG Support
// ============================================================================
// A comprehensive workflow orchestration engine for Karl demonstrating:
// - Task-based workflow definitions
// - DAG (Directed Acyclic Graph) execution
// - Parallel and sequential task execution with worker pools
// - Advanced retry policies with exponential back-off
// - Persisted state for workflow resumption
// - Error handling and circuit breakers
// - Status tracking and result aggregation
// ============================================================================

// Import enhanced modules (relative paths)
let makeRetry = import "./retry_policy.k"
let RetryModule = makeRetry()

let makeParallel = import "./parallel_executor.k"
let ParallelModule = makeParallel()

let makeStorage = import "./storage.k"
let StorageModule = makeStorage()

// ----------------------------------------------------------------------------
// TIMER TASK SUPPORT
// ----------------------------------------------------------------------------

// Timer task: executes a task after a delay
let createTimerTask = (name, delay, handler) -> {
    {
        name: name,
        type: "timer",
        delay: delay,
        handler: handler,
    }
}

// Interval task: executes a task repeatedly at intervals
let createIntervalTask = (name, interval, maxRepetitions, handler) -> {
    {
        name: name,
        type: "interval",
        interval: interval,
        maxRepetitions: maxRepetitions,
        handler: handler,
    }
}

// Execute a timer task
let executeTimerTask = (timerTask, context) -> {
    log("[TIMER]", timerTask.name, "- Waiting", timerTask.delay, "ms")
    sleep(timerTask.delay)
    
    log("[TIMER]", timerTask.name, "- Executing")
    let result = timerTask.handler(context) ? {
        { success: false, error: "Timer task execution failed" }
    }
    
    if result.success {
        log("[TIMER]", timerTask.name, "- Completed successfully")
    } else {
        log("[TIMER]", timerTask.name, "- Failed")
    }
    
    result
}

// Execute an interval task
let executeIntervalTask = (intervalTask, context) -> {
    log("[INTERVAL]", intervalTask.name, "- Starting with interval:", intervalTask.interval, "ms")
    
    for i < intervalTask.maxRepetitions with i = 0, results = [], allSuccess = true {
        log("[INTERVAL]", intervalTask.name, "- Iteration", i + 1, "/", intervalTask.maxRepetitions)
        
        let result = intervalTask.handler({ iteration: i + 1, context: context }) ? {
            { success: false, error: "Interval task execution failed" }
        }
        
        results += [result]
        
        if !result.success {
            allSuccess = false
        }
        
        // Sleep between iterations (but not after the last one)
        if i < intervalTask.maxRepetitions - 1 {
            sleep(intervalTask.interval)
        }
        
        i = i + 1
    } then {
        success: allSuccess,
        iterations: intervalTask.maxRepetitions,
        results: results
    }
}

// ----------------------------------------------------------------------------
// TASK EXECUTION ENGINE
// ----------------------------------------------------------------------------

// Execute a single task with error handling and retry logic
let executeTask = (task, context, config) -> {
    // Use advanced retry policy if configured, otherwise fall back to simple retry
    if config.retryPolicy != null {
        let retryEngine = RetryModule.createRetryEngine(config.retryPolicy)
        retryEngine.execute(task, context)
    } else {
        // Legacy simple retry logic
        let maxRetries = config.defaultRetries
        
        for i < maxRetries + 1 with i = 0, result = { success: false, error: "Not executed" }, success = false {
            if success { break result }
            
            let taskResult = task.handler(context) ? {
                { success: false, error: "Task execution failed" }
            }
            
            if taskResult.success {
                success = true
                result = taskResult
            } else {
                result = taskResult  // Keep last failure result
                if i < maxRetries {
                    log("[RETRY]", task.name, "- Attempt", i + 1)
                }
            }
            i = i + 1
        } then result
    }
}

// ----------------------------------------------------------------------------
// SEQUENTIAL WORKFLOW EXECUTOR
// ----------------------------------------------------------------------------

let executeSequential = (tasks, initialContext, config) -> {
    log("[WORKFLOW] Starting sequential execution of", tasks.length, "tasks")
    
    for i < tasks.length with i = 0, context = initialContext, results = [] {
        let task = tasks[i]
        log("[TASK]", task.name, "- Starting")
        
        let taskResult = executeTask(task, context, config)
        
        if taskResult.success {
            log("[TASK]", task.name, "- Success")
            results += [{
                name: task.name,
                success: true,
                data: taskResult.data,
            }]
            
            // Update context with task output for next task
            context = {
                previous: context,
                data: taskResult.data,
            }
        } else {
            log("[TASK]", task.name, "- Failed after retries")
            
            results += [{
                name: task.name,
                success: false,
                error: "Task failed",
            }]
            
            // Stop on first failure if configured
            if config.stopOnError {
                break { success: false, results: results, failedAt: task.name }
            }
        }
        
        i = i + 1
    } then { success: true, results: results }
}

// ----------------------------------------------------------------------------
// PARALLEL WORKFLOW EXECUTOR
// ----------------------------------------------------------------------------

let executeParallel = (tasks, context, config) -> {
    log("[WORKFLOW] Starting parallel execution of", tasks.length, "tasks")
    
    // Use advanced parallel executor if configured
    if config.useWorkerPool {
        log("[WORKFLOW] Using worker pool with", config.workerCount, "workers")
        
        let parallelExecutor = ParallelModule.createParallelExecutor({
            workerCount: config.workerCount,
            queueSize: config.queueSize,
            batchSize: null,  // Will use default
            enablePriority: config.enablePriority,
            shutdownTimeout: null,  // Will use default
            enableMetrics: config.enableMetrics,
        })
        
        parallelExecutor.execute(tasks, context)
    } else {
        // Legacy parallel execution (one goroutine per task)
        let resultChan = rendezvous()
        
        // Spawn workers for each task
        let workers = for i < tasks.length with i = 0, taskList = [] {
            let task = tasks[i]
            let taskId = i
            
            let worker = & (() -> {
                log("[TASK]", task.name, "- Starting (parallel)")
                let result = executeTask(task, context, config)
                
                if result.success {
                    log("[TASK]", task.name, "- Success (parallel)")
                } else {
                    log("[TASK]", task.name, "- Failed (parallel)")
                }
                
                let response = {
                    taskId: taskId,
                    name: task.name,
                    success: result.success,
                    data: result.data,
                }
                
                resultChan.send(response)
            })()
            
            taskList += [worker]
            i = i + 1
        } then taskList
        
        // Collect results
        let aggregator = & (() -> {
            for i < tasks.length with i = 0, results = [] {
                let [val, done] = resultChan.recv()
                if done { break results }
                
                results += [val]
                i = i + 1
            } then results
        })()
        
        // Wait for all workers to complete
        for i < workers.length with i = 0 {
            wait workers[i]
            i = i + 1
        } then {}
        
        resultChan.done()
        let allResults = wait aggregator
        
        let response = { success: true, results: allResults }
        response
    }
}

// ----------------------------------------------------------------------------
// DAG EXECUTOR - Directed Acyclic Graph with Dependencies
// ----------------------------------------------------------------------------

let executeDAG = (nodes, edges, initialContext, config) -> {
    log("[WORKFLOW] Starting DAG execution with", nodes.length, "nodes")
    
    // Initialize storage if persistence is enabled
    let storage = if config.enablePersistence {
        StorageModule.createStorageEngine(config.storageConfig)
    } else {
        null
    }
    
    let workflowId = if config.workflowId != null { config.workflowId } else { "workflow-" + str(now()) }
    
    // Check if we can resume from saved state
    let resumeState = if storage != null {
        let resumeCheck = storage.canResume(workflowId)
        if resumeCheck.resumable {
            log("[WORKFLOW] Resuming from saved state")
            resumeCheck.state
        } else {
            null
        }
    } else {
        null
    }
    
    // Build adjacency map for dependencies
    let dependencies = for i < nodes.length with i = 0, deps = map() {
        deps = deps.set(nodes[i].id, [])
        i = i + 1
    } then deps
    
    let finalDeps = for i < edges.length with i = 0, deps = dependencies {
        let edge = edges[i]
        let currentDeps = deps.get(edge.target)
        deps = deps.set(edge.target, currentDeps + [edge.source])
        i = i + 1
    } then deps
    
    // Channel for node completion notifications
    let completionChan = rendezvous()
    
    // Initialize or restore state
    let completedNodes = if resumeState != null {
        resumeState.completedNodes
    } else {
        for i < nodes.length with i = 0, comp = map() {
            comp = comp.set(nodes[i].id, false)
            i = i + 1
        } then comp
    }
    
    let startedNodes = if resumeState != null {
        resumeState.startedNodes
    } else {
        for i < nodes.length with i = 0, started = map() {
            started = started.set(nodes[i].id, false)
            i = i + 1
        } then started
    }
    
    let initialTotalCompleted = if resumeState != null { resumeState.totalCompleted } else { 0 }
    
    // Find nodes with no dependencies (or ready nodes if resuming) and start them
    for i < nodes.length with i = 0 {
        let node = nodes[i]
        
        if !startedNodes.get(node.id) {
            let deps = finalDeps.get(node.id)
            let allDepsMet = true
            
            for j < deps.length with j = 0 {
                if !completedNodes.get(deps[j]) {
                    allDepsMet = false
                    break allDepsMet
                }
                j = j + 1
            } then allDepsMet
            
            if allDepsMet {
                startedNodes = startedNodes.set(node.id, true)
                let task = & (() -> {
                    log("[DAG]", node.id, "- Starting")
                    let result = executeTask(node, initialContext, config)
                    completionChan.send({
                        nodeId: node.id,
                        success: result.success,
                        data: result.data,
                    })
                })()
            }
        }
        
        i = i + 1
    } then {}
    
    // Collect results
    let initialResults = if resumeState != null {
        resumeState.results
    } else {
        for i < nodes.length with i = 0, res = map() {
            res = res.set(nodes[i].id, null)
            i = i + 1
        } then res
    }
    
    let initialCompleted = completedNodes
    let initialStarted = startedNodes
    
    // Process completions
    let loopResult = for loop = true with loop = true, totalCompleted = initialTotalCompleted, res = initialResults, comp = initialCompleted, started = initialStarted, checkpointCounter = 0 {
        let [msg, done] = completionChan.recv()
        if done { break { results: res, completed: comp, totalCompleted: totalCompleted } }
        
        log("[DAG]", msg.nodeId, "- Completed")
        
        // Mark as completed
        comp = comp.set(msg.nodeId, true)
        res = res.set(msg.nodeId, msg)
        totalCompleted = totalCompleted + 1
        checkpointCounter = checkpointCounter + 1
        
        // Checkpoint state periodically if persistence enabled
        if storage != null {
            if checkpointCounter >= 5 {  // Checkpoint every 5 completions
                log("[WORKFLOW] Creating checkpoint...")
                let currentState = {
                    completedNodes: comp,
                    startedNodes: started,
                    results: res,
                    totalCompleted: totalCompleted,
                    status: "running",
                    nodeCount: nodes.length,
                    edgeCount: edges.length,
                }
                storage.checkpoint(workflowId, currentState)
                checkpointCounter = 0
            }
        }
        
        // Check all nodes to see if any are now ready
        for i < nodes.length with i = 0 {
            let node = nodes[i]
            // If not already completed and dependencies are satisfied, start it
            if !started.get(node.id) {
                let deps = finalDeps.get(node.id)
                let ready = true
                for j < deps.length with j = 0 {
                    if !comp.get(deps[j]) {
                        ready = false
                        break ready
                    }
                    j = j + 1
                } then ready
                
                if ready {
                    log("[DAG]", node.id, "- Dependencies satisfied, starting")
                    // Mark as started immediately to prevent double-start
                    started = started.set(node.id, true)
                    
                    let task = & (() -> {
                        let result = executeTask(node, initialContext, config)
                        completionChan.send({
                            nodeId: node.id,
                            success: result.success,
                            data: result.data,
                        })
                    })()
                }
            }
            i = i + 1
        } then {}
        
        // Exit when all nodes have sent completion messages
        if totalCompleted >= nodes.length {
            break { results: res, completed: comp, totalCompleted: totalCompleted }
        }
    } then { results: res, completed: comp, totalCompleted: totalCompleted }
    
    // Save final state if persistence enabled
    if storage != null {
        log("[WORKFLOW] Saving final state...")
        let finalState = {
            completedNodes: loopResult.completed,
            startedNodes: initialStarted,
            results: loopResult.results,
            totalCompleted: loopResult.totalCompleted,
            status: "completed",
            nodeCount: nodes.length,
            edgeCount: edges.length,
        }
        storage.save(workflowId, finalState)
    }
    
    // Return final workflow result with results converted to array
    {
        success: true,
        workflowId: workflowId,
        results: for i < nodes.length with i = 0, arr = [] {
            let nodeId = nodes[i].id
            let result = loopResult.results.get(nodeId)
            arr += [result]
            i = i + 1
        } then arr,
    }
}

// ----------------------------------------------------------------------------
// SUB-DAG SUPPORT - Reusable Workflow Components
// ----------------------------------------------------------------------------

// Create a sub-DAG node that can be used within a larger DAG
let createSubDAG = (id, name, nodes, edges) -> {
    {
        id: id,
        name: name,
        type: "subdag",
        nodes: nodes,
        edges: edges,
        handler: (context) -> {
            // Execute the sub-DAG
            log("[SUBDAG]", name, "- Starting with", nodes.length, "nodes")
            let result = executeDAG(nodes, edges, context, context.config)
            
            if result.success {
                log("[SUBDAG]", name, "- Completed successfully")
                {
                    success: true,
                    data: {
                        subdagName: name,
                        nodeCount: nodes.length,
                        results: result.results
                    }
                }
            } else {
                log("[SUBDAG]", name, "- Failed")
                { success: false, error: "Sub-DAG execution failed" }
            }
        }
    }
}

// Execute a DAG that may contain sub-DAG nodes
let executeDAGWithSubDAGs = (nodes, edges, initialContext, config) -> {
    log("[WORKFLOW] Starting DAG execution with potential sub-DAGs")
    
    // Enhance context to pass config to sub-DAGs
    let enhancedContext = {
        original: initialContext,
        config: config
    }
    
    // Execute using the standard DAG executor
    // Sub-DAG nodes will execute their embedded DAGs via their handlers
    executeDAG(nodes, edges, enhancedContext, config)
}

// ----------------------------------------------------------------------------
// PIPELINE EXECUTOR - Multi-stage with Worker Pools
// ----------------------------------------------------------------------------

let executePipeline = (stages, items, config) -> {
    log("[PIPELINE] Starting", stages.length, "stage pipeline with", items.length, "items")
    
    // Create channels between stages
    let channels = for i < stages.length with i = 0, chans = [] {
        chans += [rendezvous()]
        i = i + 1
    } then chans
    
    // Producer: feed initial items into first stage
    let producer = & (() -> {
        for i < items.length with i = 0 {
            channels[0].send(items[i])
            i = i + 1
        } then {}
        channels[0].done()
        log("[PIPELINE] Producer finished")
    })()
    
    // Create workers for each stage
    let allWorkers = for stageIdx < stages.length with stageIdx = 0, workers = [] {
        let stage = stages[stageIdx]
        let numWorkers = if stage.workers != null { stage.workers } else { config.defaultWorkers }
        let inputChan = channels[stageIdx]
        let outputChan = if stageIdx < stages.length - 1 { channels[stageIdx + 1] } else { null }
        
        log("[PIPELINE] Stage", stage.name, "- Spawning", numWorkers, "workers")
        
        // Spawn workers for this stage
        let stageWorkers = for wid < numWorkers with wid = 0, workerList = [] {
            let workerId = wid + 1
            
            let worker = & (() -> {
                for processing = true with processing = true, processed = 0 {
                    let [item, done] = inputChan.recv()
                    if done {
                        processing = false
                        break processed
                    }
                    
                    // Process item through stage handler
                    let result = stage.handler(item) ? {
                        { success: false, data: null }
                    }
                    
                    if result.success {
                        processed = processed + 1
                        
                        // Send to next stage if exists
                        if outputChan != null {
                            outputChan.send(result.data)
                        }
                    }
                } then processed
            })()
            
            workerList += [worker]
            wid = wid + 1
        } then workerList
        
        workers += [stageWorkers]
        stageIdx = stageIdx + 1
    } then workers
    
    // Wait for producer
    wait producer
    
    // Wait for each stage to complete and close channels
    for stageIdx < stages.length with stageIdx = 0 {
        let stageWorkers = allWorkers[stageIdx]
        
        // Wait for all workers in this stage
        for wid < stageWorkers.length with wid = 0, totalProcessed = 0 {
            let processed = wait stageWorkers[wid]
            totalProcessed = totalProcessed + processed
            wid = wid + 1
        } then totalProcessed
        
        // Close output channel if exists
        if stageIdx < stages.length - 1 {
            channels[stageIdx + 1].done()
            log("[PIPELINE] Stage", stages[stageIdx].name, "- Completed")
        }
        
        stageIdx = stageIdx + 1
    } then {}
    
    log("[PIPELINE] All stages completed")
    { success: true }
}

// ----------------------------------------------------------------------------
// MAIN WORKFLOW ENGINE
// ----------------------------------------------------------------------------

// Default configuration
let defaultConfig = {
    // Legacy retry settings
    defaultRetries: 2,
    defaultWorkers: 3,
    stopOnError: false,
    
    // Enhanced features
    retryPolicy: null,           // Use RetryModule for advanced retry
    useWorkerPool: false,        // Use ParallelModule for worker pools
    workerCount: 4,              // Number of workers in pool
    queueSize: 100,              // Queue size for worker pool
    batchSize: 10,               // Batch size for processing
    enablePriority: false,       // Enable priority queue
    shutdownTimeout: 30000,      // Shutdown timeout in ms
    enableMetrics: true,         // Collect performance metrics
    enablePersistence: false,    // Enable state persistence
    workflowId: null,            // Unique workflow identifier
    storageConfig: {             // Storage configuration
        storageDir: "./workflow-state",
        enableAutoCheckpoint: false,
    },
}

// Merge user config with defaults to ensure all properties exist
let mergeConfig = (userConfig) -> {
    let hasRetryPolicy = if userConfig.retryPolicy != null { true } else { false }
    {
        defaultRetries: if hasRetryPolicy { 0 } else { if userConfig.defaultRetries != null { userConfig.defaultRetries } else { defaultConfig.defaultRetries } },
        defaultWorkers: if userConfig.defaultWorkers != null { userConfig.defaultWorkers } else { defaultConfig.defaultWorkers },
        stopOnError: if userConfig.stopOnError != null { userConfig.stopOnError } else { defaultConfig.stopOnError },
        retryPolicy: if userConfig.retryPolicy != null { userConfig.retryPolicy } else { null },
        useWorkerPool: if userConfig.useWorkerPool != null { userConfig.useWorkerPool } else { defaultConfig.useWorkerPool },
        workerCount: if userConfig.workerCount != null { userConfig.workerCount } else { defaultConfig.workerCount },
        queueSize: if userConfig.queueSize != null { userConfig.queueSize } else { 100 },
        batchSize: if userConfig.batchSize != null { userConfig.batchSize } else { 10 },
        enablePriority: if userConfig.enablePriority != null { userConfig.enablePriority } else { false },
        shutdownTimeout: if userConfig.shutdownTimeout != null { userConfig.shutdownTimeout } else { 30000 },
        enableMetrics: if userConfig.enableMetrics != null { userConfig.enableMetrics } else { defaultConfig.enableMetrics },
        enablePersistence: if userConfig.enablePersistence != null { userConfig.enablePersistence } else { defaultConfig.enablePersistence },
        workflowId: if userConfig.workflowId != null { userConfig.workflowId } else { null },
        storageConfig: if userConfig.storageConfig != null { userConfig.storageConfig } else { defaultConfig.storageConfig },
    }
}

// Execute workflow based on type
let execute = (workflow, context, userConfig) -> {
    // Merge user config with defaults to ensure all properties exist
    let config = mergeConfig(userConfig)
    
    log("==================================================")
    log("[ENGINE] Executing workflow:", workflow.name)
    log("==================================================")
    
    match workflow.type {
        case "sequential" -> executeSequential(workflow.tasks, context, config)
        case "parallel" -> executeParallel(workflow.tasks, context, config)
        case "dag" -> executeDAG(workflow.nodes, workflow.edges, context, config)
        case "dag-with-subdags" -> executeDAGWithSubDAGs(workflow.nodes, workflow.edges, context, config)
        case "pipeline" -> executePipeline(workflow.stages, workflow.items, config)
        case "timer" -> executeTimerTask(workflow.task, context)
        case "interval" -> executeIntervalTask(workflow.task, context)
        case _ -> { success: false, error: "Unknown workflow type" }
    }
}

// ----------------------------------------------------------------------------
// MODULE EXPORT
// ----------------------------------------------------------------------------

// Export module as a function (Karl module pattern)
let makeModule = () -> {
    {
        // Main execution function
        execute: execute,
        
        // Execution modes
        executeSequential: executeSequential,
        executeParallel: executeParallel,
        executeDAG: executeDAG,
        executeDAGWithSubDAGs: executeDAGWithSubDAGs,
        executePipeline: executePipeline,
        executeTimerTask: executeTimerTask,
        executeIntervalTask: executeIntervalTask,
        
        // Sub-DAG support
        createSubDAG: createSubDAG,
        
        // Configuration
        defaultConfig: defaultConfig,
    }
}

makeModule
