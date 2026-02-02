// ============================================================================
// WORKFLOW ENGINE - Task-based Execution with DAG Support
// ============================================================================
// A comprehensive workflow orchestration engine for Karl demonstrating:
// - Task-based workflow definitions
// - DAG (Directed Acyclic Graph) execution
// - Parallel and sequential task execution
// - Worker pools for pipeline stages
// - Error handling and retries
// - Status tracking and result aggregation
// ============================================================================

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

// ----------------------------------------------------------------------------
// DAG EXECUTOR - Directed Acyclic Graph with Dependencies
// ----------------------------------------------------------------------------

let executeDAG = (nodes, edges, initialContext, config) -> {
    log("[WORKFLOW] Starting DAG execution with", nodes.length, "nodes")
    
    // Build adjacency map for dependencies
    let dependencies = for i < nodes.length with i = 0, deps = {} {
        deps[nodes[i].id] = []
        i = i + 1
    } then deps
    
    let finalDeps = for i < edges.length with i = 0, deps = dependencies {
        let edge = edges[i]
        let currentDeps = deps[edge.target]
        deps[edge.target] = currentDeps + [edge.source]
        i = i + 1
    } then deps
    
    // Channel for node completion notifications
    let completionChan = rendezvous()
    
    // Track started/completed nodes with boolean flags keyed by node id.
    let completedNodes = for i < nodes.length with i = 0, comp = {} {
        comp[nodes[i].id] = false
        i = i + 1
    } then comp
    
    let startedNodes = for i < nodes.length with i = 0, started = {} {
        started[nodes[i].id] = false
        i = i + 1
    } then started
    
    // Find nodes with no dependencies and start them
    for i < nodes.length with i = 0 {
        let node = nodes[i]
        let deps = finalDeps[node.id]
        if deps.length == 0 {
            startedNodes[node.id] = true
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
        i = i + 1
    } then {}
    
    // Collect results
    let initialResults = for i < nodes.length with i = 0, res = {} {
        res[nodes[i].id] = null
        i = i + 1
    } then res
    let initialCompleted = completedNodes
    let initialStarted = startedNodes
    
    // Process completions
    let loopResult = for loop = true with loop = true, totalCompleted = 0, res = initialResults, comp = initialCompleted, started = initialStarted {
        let [msg, done] = completionChan.recv()
        if done { break { results: res, completed: comp } }
        
        log("[DAG]", msg.nodeId, "- Completed")
        
        // Mark as completed
        comp[msg.nodeId] = true
        res[msg.nodeId] = msg
        totalCompleted = totalCompleted + 1
        
        // Check all nodes to see if any are now ready
        for i < nodes.length with i = 0 {
            let node = nodes[i]
            // If not already completed and dependencies are satisfied, start it
            if !started[node.id] {
                let deps = finalDeps[node.id]
                let ready = true
                for j < deps.length with j = 0 {
                    if !comp[deps[j]] {
                        ready = false
                        break ready
                    }
                    j = j + 1
                } then ready
                
                if ready {
                    log("[DAG]", node.id, "- Dependencies satisfied, starting")
                    // Mark as started immediately to prevent double-start
                    started[node.id] = true
                    
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
            break { results: res, completed: comp }
        }
    } then { results: res, completed: comp }
    
    // Return final workflow result with results converted to array
    {
        success: true,
        results: for i < nodes.length with i = 0, arr = [] {
            let nodeId = nodes[i].id
            let result = loopResult.results[nodeId]
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
        let numWorkers = if stage.workers { stage.workers } else { config.defaultWorkers }
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
    defaultRetries: 2,
    defaultWorkers: 3,
    stopOnError: false,
}

// Execute workflow based on type
let execute = (workflow, context, config) -> {
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

log("Workflow Engine Module Loaded!")
