// ============================================================================
// WORKFLOW STORAGE - Persisted DAG State Management
// ============================================================================
// Provides state persistence for workflow execution, allowing:
// - Save/Load workflow state to disk
// - Resume interrupted workflows
// - Track execution history
// - Checkpoint-based recovery
// ============================================================================

// ----------------------------------------------------------------------------
// STORAGE CONFIGURATION
// ----------------------------------------------------------------------------

let defaultStorageConfig = {
    storageDir: "./workflow-state",
    checkpointInterval: 5000,  // Auto-checkpoint every 5 seconds
    enableAutoCheckpoint: false,
    compressionEnabled: false,
}

// Safe optional field access for user-provided objects/config.
let getOr = (obj, key, fallback) -> obj[key] ? { fallback }

// ----------------------------------------------------------------------------
// STATE SERIALIZATION
// ----------------------------------------------------------------------------

// Serialize workflow state to JSON-compatible structure
let serializeWorkflowState = (workflowId, state) -> {
    {
        workflowId: workflowId,
        timestamp: now(),
        version: "1.0",
        state: {
            completedNodes: state.completedNodes,
            startedNodes: state.startedNodes,
            results: state.results,
            totalCompleted: state.totalCompleted,
            status: state.status,
        },
        metadata: {
            nodeCount: state.nodeCount,
            edgeCount: state.edgeCount,
        }
    }
}

// Deserialize workflow state from JSON structure
let deserializeWorkflowState = (serialized) -> {
    if serialized.state == null {
        { success: false, error: "Invalid state format" }
    } else {
        {
            success: true,
            state: {
                completedNodes: serialized.state.completedNodes,
                startedNodes: serialized.state.startedNodes,
                results: serialized.state.results,
                totalCompleted: serialized.state.totalCompleted,
                status: serialized.state.status,
                nodeCount: serialized.metadata.nodeCount,
                edgeCount: serialized.metadata.edgeCount,
            }
        }
    }
}

// ----------------------------------------------------------------------------
// FILE OPERATIONS
// ----------------------------------------------------------------------------

// Generate state file path for a workflow
let getStateFilePath = (workflowId, config) -> {
    config.storageDir + "/" + workflowId + ".json"
}

// Save workflow state to disk
let saveState = (workflowId, state, config) -> {
    let serialized = serializeWorkflowState(workflowId, state)
    let json = encodeJson(serialized)
    let filePath = getStateFilePath(workflowId, config)
    
    log("[STORAGE] Saving state for workflow:", workflowId)
    
    let writeResult = writeFile(filePath, json) ? {
        { success: false, error: "Failed to write state file" }
    }
    
    if writeResult.success {
        log("[STORAGE] State saved successfully to:", filePath)
        { success: true, path: filePath }
    } else {
        log("[STORAGE] Failed to save state:", writeResult.error)
        writeResult
    }
}

// Load workflow state from disk
let loadState = (workflowId, config) -> {
    let filePath = getStateFilePath(workflowId, config)
    
    log("[STORAGE] Loading state for workflow:", workflowId)
    
    let readResult = readFile(filePath) ? {
        { success: false, error: "Failed to read state file" }
    }
    
    if !readResult.success {
        log("[STORAGE] No saved state found for:", workflowId)
        { success: false, error: "No saved state" }
    } else {
        let parseResult = decodeJson(readResult.data) ? {
            { success: false, error: "Failed to parse state JSON" }
        }
        
        if parseResult.success {
            let deserialized = deserializeWorkflowState(parseResult.data)
            if deserialized.success {
                log("[STORAGE] State loaded successfully from:", filePath)
                { success: true, state: deserialized.state }
            } else {
                log("[STORAGE] Failed to deserialize state:", deserialized.error)
                deserialized
            }
        } else {
            log("[STORAGE] Failed to parse JSON:", parseResult.error)
            parseResult
        }
    }
}

// Delete workflow state from disk
let deleteState = (workflowId, config) -> {
    let filePath = getStateFilePath(workflowId, config)
    
    log("[STORAGE] Deleting state for workflow:", workflowId)
    
    let deleteResult = deleteFile(filePath) ? {
        { success: false, error: "Failed to delete state file" }
    }
    
    if deleteResult.success {
        log("[STORAGE] State deleted successfully")
        { success: true }
    } else {
        log("[STORAGE] Failed to delete state:", deleteResult.error)
        deleteResult
    }
}

// ----------------------------------------------------------------------------
// CHECKPOINT MANAGEMENT
// ----------------------------------------------------------------------------

// Create a checkpoint of current workflow state
let checkpoint = (workflowId, state, config) -> {
    log("[CHECKPOINT] Creating checkpoint for:", workflowId)
    saveState(workflowId, state, config)
}

// Auto-checkpoint worker - runs in background
let startAutoCheckpoint = (workflowId, stateRef, config) -> {
    if !config.enableAutoCheckpoint {
        log("[CHECKPOINT] Auto-checkpoint disabled")
        { success: false, error: "Auto-checkpoint disabled" }
    } else {
        log("[CHECKPOINT] Starting auto-checkpoint worker")
        log("[CHECKPOINT] Interval:", config.checkpointInterval, "ms")
        
        let worker = & (() -> {
            for running = true with running = true, checkpointCount = 0 {
                sleep(config.checkpointInterval)
                
                // Get current state from reference
                let currentState = stateRef.get()
                
                if currentState.status == "completed" || currentState.status == "failed" {
                    log("[CHECKPOINT] Workflow finished, stopping auto-checkpoint")
                    break checkpointCount
                }
                
                checkpoint(workflowId, currentState, config)
                checkpointCount = checkpointCount + 1
            } then checkpointCount
        })()
        
        { success: true, worker: worker }
    }
}

// ----------------------------------------------------------------------------
// RESUME SUPPORT
// ----------------------------------------------------------------------------

// Check if a workflow can be resumed
let canResume = (workflowId, config) -> {
    let loadResult = loadState(workflowId, config)
    
    if loadResult.success {
        let state = loadResult.state
        if state.status == "running" || state.status == "paused" {
            log("[RESUME] Workflow can be resumed")
            { success: true, resumable: true, state: state }
        } else {
            log("[RESUME] Workflow already completed or failed")
            { success: true, resumable: false, state: state }
        }
    } else {
        log("[RESUME] No state found, cannot resume")
        { success: false, resumable: false }
    }
}

// Get list of nodes that still need to be executed
let getIncompleteNodes = (nodes, state) -> {
    for i < nodes.length with i = 0, incomplete = [] {
        let node = nodes[i]
        let isCompleted = state.completedNodes[node.id]
        
        if !isCompleted {
            incomplete += [node]
        }
        
        i = i + 1
    } then incomplete
}

// Get list of nodes that are ready to execute (dependencies met)
let getReadyNodes = (nodes, edges, state) -> {
    // Build dependency map
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
    
    // Find nodes that are not started and have all dependencies completed
    for i < nodes.length with i = 0, ready = [] {
        let node = nodes[i]
        let isStarted = state.startedNodes[node.id]
        
        if !isStarted {
            let deps = finalDeps[node.id]
            let allDepsMet = true
            
            for j < deps.length with j = 0 {
                if !state.completedNodes[deps[j]] {
                    allDepsMet = false
                    break allDepsMet
                }
                j = j + 1
            } then allDepsMet
            
            if allDepsMet {
                ready += [node]
            }
        }
        
        i = i + 1
    } then ready
}

// ----------------------------------------------------------------------------
// STATE INITIALIZATION
// ----------------------------------------------------------------------------

// Create initial state for a new workflow
let createInitialState = (nodes, edges) -> {
    {
        completedNodes: for i < nodes.length with i = 0, comp = {} {
            comp[nodes[i].id] = false
            i = i + 1
        } then comp,
        startedNodes: for i < nodes.length with i = 0, started = {} {
            started[nodes[i].id] = false
            i = i + 1
        } then started,
        results: for i < nodes.length with i = 0, res = {} {
            res[nodes[i].id] = null
            i = i + 1
        } then res,
        totalCompleted: 0,
        status: "initialized",
        nodeCount: nodes.length,
        edgeCount: edges.length,
    }
}

// ----------------------------------------------------------------------------
// STORAGE ENGINE API
// ----------------------------------------------------------------------------

// Create a storage engine instance
let createStorageEngine = (config) -> {
    let finalConfig = if config {
        {
            storageDir: getOr(config, "storageDir", defaultStorageConfig.storageDir),
            checkpointInterval: getOr(config, "checkpointInterval", defaultStorageConfig.checkpointInterval),
            enableAutoCheckpoint: getOr(config, "enableAutoCheckpoint", defaultStorageConfig.enableAutoCheckpoint),
            compressionEnabled: getOr(config, "compressionEnabled", defaultStorageConfig.compressionEnabled),
        }
    } else {
        defaultStorageConfig
    }
    
    {
        config: finalConfig,
        save: (workflowId, state) -> saveState(workflowId, state, finalConfig),
        load: (workflowId) -> loadState(workflowId, finalConfig),
        delete: (workflowId) -> deleteState(workflowId, finalConfig),
        checkpoint: (workflowId, state) -> checkpoint(workflowId, state, finalConfig),
        canResume: (workflowId) -> canResume(workflowId, finalConfig),
        createInitialState: (nodes, edges) -> createInitialState(nodes, edges),
        getIncompleteNodes: (nodes, state) -> getIncompleteNodes(nodes, state),
        getReadyNodes: (nodes, edges, state) -> getReadyNodes(nodes, edges, state),
    }
}

// ----------------------------------------------------------------------------
// MODULE EXPORT
// ----------------------------------------------------------------------------

// Export module as a function (Karl module pattern)
let makeModule = () -> {
    {
        // Functions
        createStorageEngine: createStorageEngine,
        serializeState: serializeState,
        deserializeState: deserializeState,
        saveState: saveState,
        loadState: loadState,
        deleteState: deleteState,
        checkpoint: checkpoint,
        canResume: canResume,
        getIncompleteNodes: getIncompleteNodes,
        getReadyNodes: getReadyNodes,
        createInitialState: createInitialState,
    }
}

makeModule
