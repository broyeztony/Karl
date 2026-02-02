// ============================================================================
// QUICK START: Timer Tasks & Sub-DAGs
// ============================================================================
// Minimal examples to get started with the new workflow features
// ============================================================================

let makeEngine = import "examples/contrib/workflow/engine.k"
let WorkflowEngine = makeEngine()

log("================================================================================")
log("QUICK START: Timer Tasks & Sub-DAGs")
log("================================================================================")
log("")

// ============================================================================
// TIMER TASK QUICK START
// ============================================================================

log("Example 1: Simple Timer Task (Delayed Execution)")
log("─────────────────────────────────────────────────")

let timerDemo = {
    name: "Welcome Email",
    type: "timer",
    task: {
        name: "Send Welcome",
        delay: 500,
        handler: (ctx) -> {
            log("  📧 Sending welcome email...")
            { success: true, data: "Email sent!" }
        }
    }
}

let result1 = WorkflowEngine.execute(timerDemo, {}, WorkflowEngine.defaultConfig)
log("Status:", if result1.success { "✓" } else { "✗" })
log("")

// ============================================================================
// INTERVAL TASK QUICK START
// ============================================================================

log("Example 2: Interval Task (Repeated Execution)")
log("──────────────────────────────────────────────")

let intervalDemo = {
    name: "Heartbeat Monitor",
    type: "interval",
    task: {
        name: "Check Status",
        interval: 200,
        maxRepetitions: 3,
        handler: (ctx) -> {
            log("  💓 Heartbeat #", ctx.iteration)
            { success: true, data: { beat: ctx.iteration } }
        }
    }
}

let result2 = WorkflowEngine.execute(intervalDemo, {}, WorkflowEngine.defaultConfig)
log("Status:", if result2.success { "✓" } else { "✗" })
log("Total iterations:", result2.iterations)
log("")

// ============================================================================
// SUB-DAG QUICK START
// ============================================================================

log("Example 3: Reusable Sub-DAG")
log("───────────────────────────")

// Define a reusable validation component
let createValidation = () -> {
    let nodes = [
        {
            id: "check-1",
            name: "Check Format",
            handler: (ctx) -> {
                log("    ✓ Format check passed")
                { success: true, data: { valid: true } }
            }
        },
        {
            id: "check-2",
            name: "Check Schema",
            handler: (ctx) -> {
                log("    ✓ Schema check passed")
                { success: true, data: { valid: true } }
            }
        }
    ]
    { nodes: nodes, edges: [] }  // Parallel execution
}

let validation = createValidation()
let validationSubDAG = WorkflowEngine.createSubDAG(
    "validation",
    "Validation Sub-DAG",
    validation.nodes,
    validation.edges
)

// Use the sub-DAG in a workflow
let workflow = {
    name: "Data Pipeline",
    type: "dag-with-subdags",
    nodes: [
        {
            id: "load",
            name: "Load Data",
            handler: (ctx) -> {
                log("  📥 Loading data...")
                { success: true, data: { records: 100 } }
            }
        },
        validationSubDAG,  // <- Reusable component
        {
            id: "save",
            name: "Save Data",
            handler: (ctx) -> {
                log("  💾 Saving data...")
                { success: true, data: { saved: true } }
            }
        }
    ],
    edges: [
        { source: "load", target: "validation" },
        { source: "validation", target: "save" }
    ]
}

let result3 = WorkflowEngine.execute(workflow, {}, WorkflowEngine.defaultConfig)
log("Status:", if result3.success { "✓" } else { "✗" })
log("")

// ============================================================================
// COMBINED EXAMPLE
// ============================================================================

log("Example 4: Timer + Sub-DAG Combined")
log("────────────────────────────────────")

// Delayed workflow that uses a sub-DAG
let processingSubDAG = WorkflowEngine.createSubDAG(
    "process",
    "Processing Steps",
    [
        {
            id: "step-1",
            name: "Transform",
            handler: (ctx) -> {
                log("    🔄 Transforming...")
                { success: true, data: {} }
            }
        },
        {
            id: "step-2",
            name: "Enrich",
            handler: (ctx) -> {
                log("    ➕ Enriching...")
                { success: true, data: {} }
            }
        }
    ],
    [{ source: "step-1", target: "step-2" }]
)

// Delayed workflow
let delayedWorkflow = & (() -> {
    log("  ⏰ Waiting before processing...")
    sleep(300)
    
    let wf = {
        name: "Delayed Processing",
        type: "dag-with-subdags",
        nodes: [processingSubDAG],
        edges: []
    }
    
    WorkflowEngine.execute(wf, {}, WorkflowEngine.defaultConfig)
})()

let result4 = wait delayedWorkflow
log("Status:", if result4.success { "✓" } else { "✗" })
log("")

// ============================================================================
// SUMMARY
// ============================================================================

log("================================================================================")
log("✓ Quick Start Complete!")
log("================================================================================")
log("")
log("You've learned:")
log("  1. Timer tasks - Delayed execution with configurable delays")
log("  2. Interval tasks - Repeated execution at regular intervals")
log("  3. Sub-DAGs - Reusable workflow components")
log("  4. Combining features - Timer + Sub-DAG integration")
log("")
log("Next steps:")
log("  • Check timer_tasks.k for advanced timer examples")
log("  • Check subdag_demo.k for complex sub-DAG patterns")
log("  • Read README.md for complete documentation")
log("")
