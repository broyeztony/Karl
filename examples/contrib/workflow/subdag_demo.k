// ============================================================================
// SUB-DAG DEMO - Reusable Workflow Components
// ============================================================================
// Demonstrates sub-DAG features for composable workflows:
// - Creating reusable workflow components
// - Embedding sub-DAGs within larger workflows
// - Hierarchical workflow composition
// - Data flow between parent and child DAGs
// ============================================================================

let makeEngine = import "engine.k"
let WorkflowEngine = makeEngine()

log("================================================================================")
log("SUB-DAG: Reusable and Composable Workflow Components")
log("================================================================================")
log("")

// ============================================================================
// REUSABLE SUB-DAG COMPONENTS
// ============================================================================

// Sub-DAG 1: Data validation pipeline
let createValidationSubDAG = () -> {
    let nodes = [
        {
            id: "validate-format",
            name: "Validate Format",
            handler: (context) -> {
                log("    [VALIDATE] Checking data format...")
                sleep(100)
                {
                    success: true,
                    data: { formatValid: true, records: 150 }
                }
            }
        },
        {
            id: "validate-schema",
            name: "Validate Schema",
            handler: (context) -> {
                log("    [VALIDATE] Checking schema compliance...")
                sleep(120)
                {
                    success: true,
                    data: { schemaValid: true, errors: 0 }
                }
            }
        },
        {
            id: "validate-integrity",
            name: "Validate Data Integrity",
            handler: (context) -> {
                log("    [VALIDATE] Checking data integrity...")
                sleep(90)
                {
                    success: true,
                    data: { integrityOk: true, checksumValid: true }
                }
            }
        }
    ]
    
    let edges = [
        { source: "validate-format", target: "validate-schema" },
        { source: "validate-format", target: "validate-integrity" }
    ]
    
    { nodes: nodes, edges: edges }
}

// Sub-DAG 2: Data transformation pipeline
let createTransformSubDAG = () -> {
    let nodes = [
        {
            id: "normalize",
            name: "Normalize Data",
            handler: (context) -> {
                log("    [TRANSFORM] Normalizing values...")
                sleep(110)
                {
                    success: true,
                    data: { normalized: true, range: "0-1" }
                }
            }
        },
        {
            id: "enrich",
            name: "Enrich Data",
            handler: (context) -> {
                log("    [TRANSFORM] Enriching with metadata...")
                sleep(130)
                {
                    success: true,
                    data: { enriched: true, fieldsAdded: 5 }
                }
            }
        },
        {
            id: "aggregate",
            name: "Aggregate Results",
            handler: (context) -> {
                log("    [TRANSFORM] Computing aggregates...")
                sleep(100)
                {
                    success: true,
                    data: { aggregated: true, summaries: 12 }
                }
            }
        }
    ]
    
    let edges = [
        { source: "normalize", target: "enrich" },
        { source: "normalize", target: "aggregate" }
    ]
    
    { nodes: nodes, edges: edges }
}

// Sub-DAG 3: Quality assurance checks
let createQASubDAG = () -> {
    let nodes = [
        {
            id: "check-quality",
            name: "Quality Metrics",
            handler: (context) -> {
                log("    [QA] Computing quality metrics...")
                sleep(95)
                {
                    success: true,
                    data: { qualityScore: 92, threshold: 80 }
                }
            }
        },
        {
            id: "check-completeness",
            name: "Completeness Check",
            handler: (context) -> {
                log("    [QA] Verifying completeness...")
                sleep(85)
                {
                    success: true,
                    data: { complete: true, coverage: 98 }
                }
            }
        }
    ]
    
    let edges = []  // Parallel execution
    
    { nodes: nodes, edges: edges }
}

// ============================================================================
// EXAMPLE 1: Simple Sub-DAG Integration
// ============================================================================

log("[EXAMPLE 1] Single sub-DAG execution")
log("")

let validation = createValidationSubDAG()
let validationSubDAG = WorkflowEngine.createSubDAG(
    "validation-pipeline",
    "Data Validation Pipeline",
    validation.nodes,
    validation.edges
)

let simpleWorkflow = {
    name: "Simple Validation Workflow",
    type: "dag-with-subdags",
    nodes: [
        {
            id: "load-data",
            name: "Load Data",
            handler: (context) -> {
                log("  [LOAD] Loading dataset...")
                sleep(150)
                {
                    success: true,
                    data: { loaded: true, records: 1000 }
                }
            }
        },
        validationSubDAG,
        {
            id: "report",
            name: "Generate Report",
            handler: (context) -> {
                log("  [REPORT] Creating validation report...")
                sleep(100)
                {
                    success: true,
                    data: { report: "validation-report.pdf", generated: true }
                }
            }
        }
    ],
    edges: [
        { source: "load-data", target: "validation-pipeline" },
        { source: "validation-pipeline", target: "report" }
    ]
}

let result1 = WorkflowEngine.execute(simpleWorkflow, {}, WorkflowEngine.defaultConfig)
log("Result:", if result1.success { "✓ Success" } else { "✗ Failed" })
log("")

// ============================================================================
// EXAMPLE 2: Multiple Sub-DAGs in Sequence
// ============================================================================

log("[EXAMPLE 2] Sequential sub-DAG pipeline")
log("")

let validationDAG = createValidationSubDAG()
let transformDAG = createTransformSubDAG()
let qaDAG = createQASubDAG()

let sequentialWorkflow = {
    name: "ETL Pipeline with Sub-DAGs",
    type: "dag-with-subdags",
    nodes: [
        {
            id: "extract",
            name: "Extract Data",
            handler: (context) -> {
                log("  [EXTRACT] Fetching data from sources...")
                sleep(120)
                {
                    success: true,
                    data: { sources: 3, totalRecords: 5000 }
                }
            }
        },
        WorkflowEngine.createSubDAG(
            "validation",
            "Validation Sub-DAG",
            validationDAG.nodes,
            validationDAG.edges
        ),
        WorkflowEngine.createSubDAG(
            "transform",
            "Transform Sub-DAG",
            transformDAG.nodes,
            transformDAG.edges
        ),
        WorkflowEngine.createSubDAG(
            "qa",
            "Quality Assurance Sub-DAG",
            qaDAG.nodes,
            qaDAG.edges
        ),
        {
            id: "load",
            name: "Load to Warehouse",
            handler: (context) -> {
                log("  [LOAD] Writing to data warehouse...")
                sleep(200)
                {
                    success: true,
                    data: { loaded: true, warehouse: "production", rowsWritten: 4850 }
                }
            }
        }
    ],
    edges: [
        { source: "extract", target: "validation" },
        { source: "validation", target: "transform" },
        { source: "transform", target: "qa" },
        { source: "qa", target: "load" }
    ]
}

let result2 = WorkflowEngine.execute(sequentialWorkflow, {}, WorkflowEngine.defaultConfig)
log("Result:", if result2.success { "✓ Success" } else { "✗ Failed" })
log("Nodes executed:", result2.results.length)
log("")

// ============================================================================
// EXAMPLE 3: Parallel Sub-DAGs with Fan-Out/Fan-In
// ============================================================================

log("[EXAMPLE 3] Parallel sub-DAG execution with fan-out/fan-in")
log("")

// Create processing sub-DAG for different data types
let createProcessingSubDAG = (dataType) -> {
    let nodes = [
        {
            id: dataType + "-clean",
            name: "Clean " + dataType + " Data",
            handler: (context) -> {
                log("    [PROCESS] Cleaning", dataType, "data...")
                sleep(100)
                {
                    success: true,
                    data: { cleaned: true, type: dataType }
                }
            }
        },
        {
            id: dataType + "-analyze",
            name: "Analyze " + dataType + " Data",
            handler: (context) -> {
                log("    [PROCESS] Analyzing", dataType, "data...")
                sleep(120)
                {
                    success: true,
                    data: { analyzed: true, type: dataType, insights: 8 }
                }
            }
        }
    ]
    
    let edges = [
        { source: dataType + "-clean", target: dataType + "-analyze" }
    ]
    
    { nodes: nodes, edges: edges }
}

let userProcessing = createProcessingSubDAG("user")
let productProcessing = createProcessingSubDAG("product")
let orderProcessing = createProcessingSubDAG("order")

let parallelWorkflow = {
    name: "Multi-Source Processing Pipeline",
    type: "dag-with-subdags",
    nodes: [
        {
            id: "ingest",
            name: "Ingest All Sources",
            handler: (context) -> {
                log("  [INGEST] Loading data from all sources...")
                sleep(150)
                {
                    success: true,
                    data: { sources: ["users", "products", "orders"], ready: true }
                }
            }
        },
        WorkflowEngine.createSubDAG(
            "process-users",
            "User Data Processing",
            userProcessing.nodes,
            userProcessing.edges
        ),
        WorkflowEngine.createSubDAG(
            "process-products",
            "Product Data Processing",
            productProcessing.nodes,
            productProcessing.edges
        ),
        WorkflowEngine.createSubDAG(
            "process-orders",
            "Order Data Processing",
            orderProcessing.nodes,
            orderProcessing.edges
        ),
        {
            id: "merge",
            name: "Merge Results",
            handler: (context) -> {
                log("  [MERGE] Combining all processed data...")
                sleep(180)
                {
                    success: true,
                    data: { merged: true, totalRecords: 15000, datasets: 3 }
                }
            }
        }
    ],
    edges: [
        // Fan-out: ingest to all processing sub-DAGs
        { source: "ingest", target: "process-users" },
        { source: "ingest", target: "process-products" },
        { source: "ingest", target: "process-orders" },
        // Fan-in: all processing to merge
        { source: "process-users", target: "merge" },
        { source: "process-products", target: "merge" },
        { source: "process-orders", target: "merge" }
    ]
}

let result3 = WorkflowEngine.execute(parallelWorkflow, {}, WorkflowEngine.defaultConfig)
log("Result:", if result3.success { "✓ Success" } else { "✗ Failed" })
log("Total nodes (including sub-DAG nodes):", result3.results.length)
log("")

// ============================================================================
// EXAMPLE 4: Nested Sub-DAGs (Sub-DAG within Sub-DAG)
// ============================================================================

log("[EXAMPLE 4] Nested sub-DAG composition")
log("")

// Inner sub-DAG: Data cleansing
let cleansingSubDAG = {
    nodes: [
        {
            id: "remove-duplicates",
            name: "Remove Duplicates",
            handler: (context) -> {
                log("      [CLEANSE] Removing duplicates...")
                sleep(80)
                {
                    success: true,
                    data: { duplicatesRemoved: 42 }
                }
            }
        },
        {
            id: "fix-nulls",
            name: "Fix Null Values",
            handler: (context) -> {
                log("      [CLEANSE] Handling null values...")
                sleep(70)
                {
                    success: true,
                    data: { nullsFixed: 18 }
                }
            }
        }
    ],
    edges: []  // Parallel
}

// Middle sub-DAG: Preprocessing (contains cleansing sub-DAG)
let preprocessingSubDAG = {
    nodes: [
        WorkflowEngine.createSubDAG(
            "cleansing",
            "Data Cleansing",
            cleansingSubDAG.nodes,
            cleansingSubDAG.edges
        ),
        {
            id: "convert-types",
            name: "Convert Data Types",
            handler: (context) -> {
                log("    [PREPROCESS] Converting data types...")
                sleep(90)
                {
                    success: true,
                    data: { conversions: 25 }
                }
            }
        }
    ],
    edges: [
        { source: "cleansing", target: "convert-types" }
    ]
}

// Outer workflow: ML Pipeline
let mlPipeline = {
    name: "Machine Learning Pipeline",
    type: "dag-with-subdags",
    nodes: [
        {
            id: "load-raw",
            name: "Load Raw Data",
            handler: (context) -> {
                log("  [LOAD] Loading raw dataset...")
                sleep(100)
                {
                    success: true,
                    data: { records: 10000, features: 50 }
                }
            }
        },
        WorkflowEngine.createSubDAG(
            "preprocessing",
            "Data Preprocessing Pipeline",
            preprocessingSubDAG.nodes,
            preprocessingSubDAG.edges
        ),
        {
            id: "train-model",
            name: "Train Model",
            handler: (context) -> {
                log("  [TRAIN] Training ML model...")
                sleep(300)
                {
                    success: true,
                    data: { modelId: "model-v1", accuracy: 0.94 }
                }
            }
        },
        {
            id: "evaluate",
            name: "Evaluate Model",
            handler: (context) -> {
                log("  [EVALUATE] Computing evaluation metrics...")
                sleep(150)
                {
                    success: true,
                    data: { precision: 0.92, recall: 0.91, f1: 0.91 }
                }
            }
        }
    ],
    edges: [
        { source: "load-raw", target: "preprocessing" },
        { source: "preprocessing", target: "train-model" },
        { source: "train-model", target: "evaluate" }
    ]
}

let result4 = WorkflowEngine.execute(mlPipeline, {}, WorkflowEngine.defaultConfig)
log("Result:", if result4.success { "✓ Success" } else { "✗ Failed" })
log("")

// ============================================================================
// SUMMARY
// ============================================================================

log("================================================================================")
log("✓ Sub-DAG Demo Complete")
log("================================================================================")
log("")
log("Features demonstrated:")
log("  ✓ Reusable workflow components")
log("  ✓ Sub-DAG integration in larger workflows")
log("  ✓ Sequential sub-DAG pipelines")
log("  ✓ Parallel sub-DAG execution with fan-out/fan-in")
log("  ✓ Nested sub-DAGs (hierarchical composition)")
log("")
log("Benefits of sub-DAGs:")
log("  • Modularity: Create once, reuse anywhere")
log("  • Maintainability: Update component logic in one place")
log("  • Clarity: Express complex workflows hierarchically")
log("  • Testing: Test sub-DAGs independently")
log("  • Scalability: Compose large workflows from small pieces")
log("")
log("Use cases:")
log("  • ETL pipelines with reusable validation/transformation steps")
log("  • ML workflows with shared preprocessing components")
log("  • Multi-tenant processing with per-tenant sub-DAGs")
log("  • Microservice orchestration with service-specific sub-DAGs")
log("  • Data quality pipelines with pluggable check modules")
log("")
