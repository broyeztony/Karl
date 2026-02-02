// Test simple DAG execution
let makeEngine = import "examples/contrib/workflow/engine.k"
let WorkflowEngine = makeEngine()

log("Testing simple DAG workflow...")

let simpleDAG = {
    name: "Simple DAG Test",
    type: "dag",  // Use normal DAG, not dag-with-subdags
    nodes: [
        {
            id: "task-1",
            name: "Task 1",
            handler: (ctx) -> {
                log("  Task 1 executing")
                { success: true, data: { result: "task1" } }
            }
        },
        {
            id: "task-2",
            name: "Task 2",
            handler: (ctx) -> {
                log("  Task 2 executing")
                { success: true, data: { result: "task2" } }
            }
        }
    ],
    edges: [
        { source: "task-1", target: "task-2" }
    ]
}

let result = WorkflowEngine.execute(simpleDAG, {}, WorkflowEngine.defaultConfig)
log("Result:", if result.success { "✓ Success" } else { "✗ Failed" })
