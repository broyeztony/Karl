// ============================================================================
// FILE WATCHER WORKFLOW
// ============================================================================
// Demonstrates a reactive workflow pattern that monitors for new files
// and automatically processes them through a pipeline.
//
// Pattern: File System Watcher + Event-Driven Processing
// Use Cases:
//   - ETL pipelines triggered by file uploads
//   - Log file processing as they arrive
//   - Document processing workflows
//   - Data ingestion systems
// ============================================================================

// Import the workflow engine
let makeEngine = import "./engine.k"
let engine = makeEngine()
let execute = engine.execute
let defaultConfig = engine.defaultConfig

log("================================================================================")
log("FILE WATCHER: Event-Driven File Processing")
log("================================================================================")
log("")

// ============================================================================
// SIMULATED FILE SYSTEM
// ============================================================================

// Simulate a file system that receives files over time
let simulateIncomingFiles = () -> {
    log("[SIMULATOR] Simulating file system with incoming files...")
    log("")
    
    // Simulate files arriving at different times
    let fileSystem = {
        iteration: 0,
        files: [],
    }
    
    fileSystem
}

// Simulate new files being detected
let pollForNewFiles = (fs, lastSeen) -> {
    // Simulate files appearing over time
    let currentIteration = fs.iteration
    
    let newFiles = if currentIteration == 0 {
        // First check: some initial files
        [
            { name: "data_001.csv", size: 1024, timestamp: 100 },
            { name: "data_002.csv", size: 2048, timestamp: 150 },
        ]
    } else { if currentIteration == 1 {
        // Second check: more files arrived
        [
            { name: "data_001.csv", size: 1024, timestamp: 100 },
            { name: "data_002.csv", size: 2048, timestamp: 150 },
            { name: "data_003.csv", size: 1536, timestamp: 200 },
            { name: "report.txt", size: 512, timestamp: 210 },
        ]
    } else { if currentIteration == 2 {
        // Third check: even more files
        [
            { name: "data_001.csv", size: 1024, timestamp: 100 },
            { name: "data_002.csv", size: 2048, timestamp: 150 },
            { name: "data_003.csv", size: 1536, timestamp: 200 },
            { name: "report.txt", size: 512, timestamp: 210 },
            { name: "data_004.csv", size: 3072, timestamp: 300 },
            { name: "summary.json", size: 768, timestamp: 320 },
        ]
    } else {
        // No new files after iteration 3
        [
            { name: "data_001.csv", size: 1024, timestamp: 100 },
            { name: "data_002.csv", size: 2048, timestamp: 150 },
            { name: "data_003.csv", size: 1536, timestamp: 200 },
            { name: "report.txt", size: 512, timestamp: 210 },
            { name: "data_004.csv", size: 3072, timestamp: 300 },
            { name: "summary.json", size: 768, timestamp: 320 },
        ]
    } } }
    
    // Find files that weren't in lastSeen and return result
    {
        detected: for i < newFiles.length with i = 0, result = [] {
            let file = newFiles[i]
            
            // Check if this file was already seen
            let wasAlreadySeen = for j < lastSeen.length with j = 0, found = false {
                if lastSeen[j].name == file.name {
                    found = true
                }
                j = j + 1
            } then found
            
            // Add to result if it's new
            if !wasAlreadySeen {
                result += [file]
            }
            i = i + 1
        } then result,
        all: newFiles,
    }
}

// ============================================================================
// FILE PROCESSING WORKFLOW
// ============================================================================

let fileProcessingWorkflow = {
    name: "File Processing Pipeline",
    type: "sequential",
    tasks: [
        {
            name: "Validate File",
            handler: (ctx) -> {
                let file = ctx.file
                log("    → Validating file:", file.name)
                
                // Check file extension
                let isCsv = file.name == "data_001.csv" || file.name == "data_002.csv" || file.name == "data_003.csv" || file.name == "data_004.csv"
                let isValid = file.size > 0 && isCsv
                
                if isValid {
                    log("      ✓ Valid CSV file")
                } else {
                    log("      ⚠ Skipping non-CSV file")
                }
                
                { success: isValid, data: { file: file, valid: isValid } }
            },
        },
        {
            name: "Process File Content",
            handler: (ctx) -> {
                let file = ctx.data.file
                log("    → Processing file content...")
                sleep(50)
                
                // Simulate parsing CSV data
                let recordCount = file.size / 64  // Simulate ~64 bytes per record
                log("      • Parsed", recordCount, "records")
                
                { success: true, data: { file: file, records: recordCount } }
            },
        },
        {
            name: "Transform Data",
            handler: (ctx) -> {
                log("    → Applying transformations...")
                sleep(30)
                
                // Simulate data transformation
                let enrichedRecords = ctx.data.records
                log("      • Enriched", enrichedRecords, "records")
                
                { success: true, data: { file: ctx.data.file, processed: enrichedRecords } }
            },
        },
        {
            name: "Store Results",
            handler: (ctx) -> {
                log("    → Storing processed data...")
                sleep(20)
                
                log("      ✓ Saved to database:", ctx.data.processed, "records")
                
                { success: true, data: { file: ctx.data.file, stored: ctx.data.processed } }
            },
        },
    ],
}

// ============================================================================
// FILE WATCHER MAIN LOOP
// ============================================================================

log("[WATCHER] Starting file system monitor...")
log("")

let fileSystem = simulateIncomingFiles()
let processedCount = 0
let skippedCount = 0

// Watcher loop - polls for new files
let watcherLoop = for iteration < 4 with iteration = 0, lastSeen = [], totalProcessed = 0, totalSkipped = 0 {
    log("================================================================================")
    log("[POLL]", iteration + 1, "- Checking for new files...")
    log("================================================================================")
    log("")
    
    // Update file system state
    fileSystem = { iteration: iteration, files: [] }
    
    // Poll for new files
    let pollResult = pollForNewFiles(fileSystem, lastSeen)
    
    if pollResult.detected.length > 0 {
        log("🔔 Detected", pollResult.detected.length, "new file(s)!")
        log("")
        
        // Process each new file
        let results = for i < pollResult.detected.length with i = 0, processed = totalProcessed, skipped = totalSkipped {
            let file = pollResult.detected[i]
            log("──────────────────────────────────────────────────────────────────────")
            log("📄 New File Detected:", file.name)
            log("   Size:", file.size, "bytes | Timestamp:", file.timestamp)
            log("──────────────────────────────────────────────────────────────────────")
            
            // Execute processing workflow (fail fast on validation errors)
            let quickFailConfig = { defaultRetries: 0, defaultWorkers: 3, stopOnError: true }
            let result = execute(fileProcessingWorkflow, { file: file }, quickFailConfig)
            
            // Check if workflow succeeded
            if result.success {
                log("  ✅ Processing completed successfully")
                processed = processed + 1
            } else {
                log("  ⚠️  File skipped (validation failed)")
                skipped = skipped + 1
            }
            log("")
            
            i = i + 1
        } then { processed: processed, skipped: skipped }
        
        totalProcessed = results.processed
        totalSkipped = results.skipped
    } else {
        log("ℹ️  No new files detected")
        log("")
    }
    
    // Update lastSeen to current state
    lastSeen = pollResult.all
    
    // Simulate waiting before next poll
    if iteration < 3 {
        log("⏱️  Waiting for next poll cycle...")
        sleep(100)
        log("")
    }
    
    iteration = iteration + 1
} then { processed: totalProcessed, skipped: totalSkipped }

// ============================================================================
// FINAL SUMMARY
// ============================================================================

log("================================================================================")
log("FILE WATCHER: MONITORING SESSION COMPLETE")
log("================================================================================")
log("")

log("📊 SUMMARY")
log("────────────────────────────────────────────────────────────────────────────")
log("Total Files Processed:", watcherLoop.processed)
log("Total Files Skipped:", watcherLoop.skipped)
log("Total Files Detected:", watcherLoop.processed + watcherLoop.skipped)
log("")

log("✅ WORKFLOW PATTERN DEMONSTRATED")
log("────────────────────────────────────────────────────────────────────────────")
log("Pattern: Event-Driven File Processing")
log("")
log("Key Features:")
log("  • Continuous polling for new files")
log("  • State tracking (remembers processed files)")
log("  • Automatic workflow execution on detection")
log("  • File validation and filtering")
log("  • Multi-stage processing pipeline")
log("  • Comprehensive error handling")
log("")

log("Real-World Applications:")
log("  • ETL pipelines triggered by file uploads")
log("  • Log aggregation and analysis systems")
log("  • Document processing workflows")
log("  • Data lake ingestion")
log("  • Backup and archival systems")
log("")

log("================================================================================")
log("")
log("💡 TIP: In production, this pattern would use:")
log("   - File system watchers (inotify, FSEvents, etc.)")
log("   - Message queues for event distribution")
log("   - Distributed task processing")
log("   - Persistent state storage")
log("================================================================================")
