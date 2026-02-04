// ============================================================================
// PARALLEL EXECUTOR - Multi-Core Task Execution with Worker Pools
// ============================================================================
// Provides efficient parallel execution of tasks:
// - Dynamic worker pool management
// - Task queue with priority support
// - Load balancing across workers
// - Graceful shutdown and cancellation
// - Performance metrics and monitoring
// ============================================================================

// ----------------------------------------------------------------------------
// WORKER POOL CONFIGURATION
// ----------------------------------------------------------------------------

let defaultPoolConfig = {
    workerCount: 4,              // Number of concurrent workers
    queueSize: 100,              // Max tasks in queue
    enablePriority: false,       // Priority queue support
    enableMetrics: true,         // Collect performance metrics
    shutdownTimeout: 30000,      // 30 seconds
}

// ----------------------------------------------------------------------------
// TASK QUEUE MANAGEMENT
// ----------------------------------------------------------------------------

// Create a task queue
let createTaskQueue = (config) -> {
    let queue = []
    let queueLock = rendezvous()
    
    {
        // Add task to queue
        enqueue: (task) -> {
            if queue.length >= config.queueSize {
                log("[QUEUE] Queue full, rejecting task:", task.name)
                { success: false, error: "Queue full" }
            } else {
                queue += [task]
                log("[QUEUE] Task enqueued:", task.name, "- Queue size:", queue.length)
                { success: true }
            }
        },
        
        // Remove and return next task
        dequeue: () -> {
            if queue.length == 0 {
                { success: false, error: "Queue empty" }
            } else {
                let task = queue[0]
                
                // Remove first element by creating new array without it
                queue = for i < queue.length with i = 1, newQueue = [] {
                    if i < queue.length {
                        newQueue += [queue[i]]
                    }
                    i = i + 1
                } then newQueue
                
                log("[QUEUE] Task dequeued:", task.name, "- Queue size:", queue.length)
                { success: true, task: task }
            }
        },
        
        // Get queue size
        size: () -> queue.length,
        
        // Check if queue is empty
        isEmpty: () -> queue.length == 0,
        
        // Clear queue
        clear: () -> {
            let oldSize = queue.length
            queue = []
            log("[QUEUE] Cleared", oldSize, "tasks")
            { success: true, cleared: oldSize }
        },
    }
}

// ----------------------------------------------------------------------------
// WORKER IMPLEMENTATION
// ----------------------------------------------------------------------------

// Create a worker that processes tasks from a queue
let createWorker = (workerId, taskChan, resultChan, config) -> {
    let metrics = {
        tasksProcessed: 0,
        tasksSucceeded: 0,
        tasksFailed: 0,
        totalDuration: 0,
    }
    
    let worker = & (() -> {
        log("[WORKER", workerId, "] Started")
        
        for running = true with running = true {
            let [task, done] = taskChan.recv()
            
            if done {
                log("[WORKER", workerId, "] Shutting down")
                break metrics
            }
            
            log("[WORKER", workerId, "] Processing:", task.name)
            
            let startTime = now()
            let result = task.handler(task.context) ? {
                { success: false, error: "Task execution failed" }
            }
            let duration = now() - startTime
            
            // Update metrics
            metrics.tasksProcessed = metrics.tasksProcessed + 1
            metrics.totalDuration = metrics.totalDuration + duration
            
            if result.success {
                metrics.tasksSucceeded = metrics.tasksSucceeded + 1
                log("[WORKER", workerId, "] Success:", task.name, "- Duration:", duration, "ms")
            } else {
                metrics.tasksFailed = metrics.tasksFailed + 1
                log("[WORKER", workerId, "] Failed:", task.name, "- Duration:", duration, "ms")
            }
            
            // Send result back
            resultChan.send({
                workerId: workerId,
                taskName: task.name,
                success: result.success,
                data: result.data,
                error: null,  // TODO: Safely access result.error if it exists
                duration: duration,
            })
        } then metrics
    })()
    
    {
        id: workerId,
        handle: worker,
        metrics: metrics,
    }
}

// ----------------------------------------------------------------------------
// WORKER POOL IMPLEMENTATION
// ----------------------------------------------------------------------------

// Create a worker pool
let createWorkerPool = (config) -> {
    let finalConfig = if config != null {
        {
            workerCount: if config.workerCount != null { config.workerCount } else { defaultPoolConfig.workerCount },
            queueSize: if config.queueSize != null { config.queueSize } else { defaultPoolConfig.queueSize },
            enablePriority: if config.enablePriority != null { config.enablePriority } else { defaultPoolConfig.enablePriority },
            enableMetrics: if config.enableMetrics != null { config.enableMetrics } else { defaultPoolConfig.enableMetrics },
            shutdownTimeout: if config.shutdownTimeout != null { config.shutdownTimeout } else { defaultPoolConfig.shutdownTimeout },
        }
    } else {
        defaultPoolConfig
    }
    
    // Use buffered channels to prevent deadlock
    // Task channel needs to hold all tasks that might be submitted
    let taskChan = buffered(finalConfig.queueSize)
    // Result channel needs to hold all results
    let resultChan = buffered(finalConfig.queueSize)
    let workers = []
    let isRunning = false
    
    {
        config: finalConfig,
        
        // Start the worker pool
        start: () -> {
            if isRunning {
                log("[POOL] Already running")
                { success: false, error: "Pool already running" }
            } else {
                log("[POOL] Starting with", finalConfig.workerCount, "workers")
                
                // Create workers
                workers = for i < finalConfig.workerCount with i = 0, workerList = [] {
                    let worker = createWorker(i + 1, taskChan, resultChan, finalConfig)
                    workerList += [worker]
                    i = i + 1
                } then workerList
                
                isRunning = true
                log("[POOL] Started successfully")
                { success: true, workerCount: finalConfig.workerCount }
            }
        },
        
        // Submit a task to the pool
        submit: (task) -> {
            if !isRunning {
                log("[POOL] Pool not running")
                { success: false, error: "Pool not running" }
            } else {
                taskChan.send(task)
                { success: true }
            }
        },
        
        // Wait for a result
        waitForResult: () -> {
            let [result, done] = resultChan.recv()
            if done {
                { success: false, error: "Result channel closed" }
            } else {
                { success: true, result: result }
            }
        },
        
        // Shutdown the pool
        shutdown: () -> {
            if !isRunning {
                log("[POOL] Not running")
                { success: false, error: "Pool not running" }
            } else {
                log("[POOL] Shutting down...")
                
                // Close task channel to signal workers to stop
                taskChan.done()
                
                // Wait for all workers to finish
                let allMetrics = for i < workers.length with i = 0, metricsList = [] {
                    let worker = workers[i]
                    let workerMetrics = wait worker.handle
                    metricsList += [{
                        workerId: worker.id,
                        metrics: workerMetrics,
                    }]
                    i = i + 1
                } then metricsList
                
                // Close result channel
                resultChan.done()
                
                isRunning = false
                log("[POOL] Shutdown complete")
                
                { success: true, metrics: allMetrics }
            }
        },
        
        // Get pool status
        getStatus: () -> {
            {
                running: isRunning,
                workerCount: workers.length,
            }
        },
    }
}

// ----------------------------------------------------------------------------
// PARALLEL EXECUTION STRATEGIES
// ----------------------------------------------------------------------------

// Execute tasks in parallel using a worker pool
let executeParallelWithPool = (tasks, context, config) -> {
    log("[PARALLEL] Starting execution of", tasks.length, "tasks")
    log("[PARALLEL] Worker count:", config.workerCount)
    
    let pool = createWorkerPool(config)
    
    // Start the pool
    let startResult = pool.start()
    if !startResult.success {
        log("[PARALLEL] Failed to start pool:", startResult.error)
        { success: false, error: startResult.error }
    } else {
        // Submit all tasks
        for i < tasks.length with i = 0 {
            let task = tasks[i]
            pool.submit({
                name: task.name,
                handler: task.handler,
                context: context,
            })
            i = i + 1
        } then {}
        
        // Collect results
        let results = for i < tasks.length with i = 0, resultList = [] {
            let resultWrapper = pool.waitForResult()
            if resultWrapper.success {
                resultList += [resultWrapper.result]
            }
            i = i + 1
        } then resultList
        
        // Shutdown pool
        let shutdownResult = pool.shutdown()
        
        log("[PARALLEL] Execution complete")
        log("[PARALLEL] Total tasks:", tasks.length)
        log("[PARALLEL] Results collected:", results.length)
        
        {
            success: true,
            results: results,
            metrics: if shutdownResult.success { shutdownResult.metrics } else { [] },
        }
    }
}

// Execute tasks with dynamic load balancing
let executeWithLoadBalancing = (tasks, context, config) -> {
    log("[LOAD_BALANCE] Starting with", tasks.length, "tasks")
    
    // Sort tasks by estimated complexity (if available)
    let sortedTasks = tasks  // TODO: Implement sorting by complexity
    
    // Execute using worker pool
    executeParallelWithPool(sortedTasks, context, config)
}

// Execute tasks with batching
let executeWithBatching = (tasks, context, config) -> {
    let batchSize = if config.batchSize != null { config.batchSize } else { 10 }
    
    log("[BATCH] Processing", tasks.length, "tasks in batches of", batchSize)
    
    let batchCount = (tasks.length + batchSize - 1) / batchSize  // Ceiling division
    
    for batchIdx < batchCount with batchIdx = 0, allResults = [] {
        let startIdx = batchIdx * batchSize
        let endIdx = if startIdx + batchSize > tasks.length {
            tasks.length
        } else {
            startIdx + batchSize
        }
        
        log("[BATCH] Processing batch", batchIdx + 1, "/", batchCount)
        
        // Extract batch
        let batch = for i < endIdx - startIdx with i = 0, batchTasks = [] {
            batchTasks += [tasks[startIdx + i]]
            i = i + 1
        } then batchTasks
        
        // Execute batch
        let batchResult = executeParallelWithPool(batch, context, config)
        
        if batchResult.success {
            allResults += batchResult.results
        }
        
        batchIdx = batchIdx + 1
    } then {
        success: true,
        results: allResults,
    }
}

// ----------------------------------------------------------------------------
// PARALLEL EXECUTOR API
// ----------------------------------------------------------------------------

// Create a parallel executor instance
let createParallelExecutor = (config) -> {
    let finalConfig = if config != null {
        {
            workerCount: if config.workerCount != null { config.workerCount } else { defaultPoolConfig.workerCount },
            queueSize: if config.queueSize != null { config.queueSize } else { defaultPoolConfig.queueSize },
            enablePriority: if config.enablePriority != null { config.enablePriority } else { defaultPoolConfig.enablePriority },
            shutdownTimeout: if config.shutdownTimeout != null { config.shutdownTimeout } else { defaultPoolConfig.shutdownTimeout },
            enableMetrics: if config.enableMetrics != null { config.enableMetrics } else { defaultPoolConfig.enableMetrics },
            batchSize: if config.batchSize != null { config.batchSize } else { 10 },
        }
    } else {
        defaultPoolConfig
    }
    
    {
        config: finalConfig,
        
        // Execute tasks in parallel
        execute: (tasks, context) -> executeParallelWithPool(tasks, context, finalConfig),
        
        // Execute with load balancing
        executeBalanced: (tasks, context) -> executeWithLoadBalancing(tasks, context, finalConfig),
        
        // Execute in batches
        executeBatched: (tasks, context) -> executeWithBatching(tasks, context, finalConfig),
        
        // Create worker pool
        createPool: () -> createWorkerPool(finalConfig),
    }
}

// ----------------------------------------------------------------------------
// MODULE EXPORT
// ----------------------------------------------------------------------------

// Export module as a function (Karl module pattern)
let makeModule = () -> {
    {
        // Functions
        createWorkerPool: createWorkerPool,
        createParallelExecutor: createParallelExecutor,
        executeParallelWithPool: executeParallelWithPool,
        executeWithLoadBalancing: executeWithLoadBalancing,
        executeWithBatching: executeWithBatching,
        
        // Configuration
        defaultPoolConfig: defaultPoolConfig,
    }
}

makeModule
