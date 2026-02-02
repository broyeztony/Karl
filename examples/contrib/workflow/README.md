# Karl Workflow Engine

A comprehensive workflow orchestration engine demonstrating Karl's powerful concurrency features.

## Quick Start

```bash
# Run the quick start guide
karl run quickstart.k

# Or explore specific examples
karl run examples.k        # Sequential & parallel workflows
karl run timer_tasks.k     # Scheduled tasks
karl run dag_pipeline.k    # Complex DAG processing
karl run csv_pipeline.k    # CSV data processing
karl run file_watcher.k    # Reactive workflows
```

## Overview

The Workflow Engine provides **six execution modes** for orchestrating tasks:

| Mode | Description | Use Case |
|------|-------------|----------|
| **Sequential** | Tasks run one after another | ETL pipelines, step-by-step processing |
| **Parallel** | Tasks run concurrently | API calls, independent operations |
| **DAG** | Dependency-based execution | Complex workflows with prerequisites |
| **Pipeline** | Multi-stage with worker pools | High-throughput data processing |
| **Timer** | Delayed or scheduled execution | Notifications, deferred processing |
| **Interval** | Repeated execution | Health checks, monitoring, polling |

## Features

### ✨ Core Capabilities
- **Dependency Resolution** - Tasks wait for prerequisites to complete
- **Parallel Execution** - Worker pools for concurrent processing  
- **Error Handling** - Configurable retries and graceful degradation
- **Context Passing** - Results flow between tasks
- **Sub-DAGs** - Reusable workflow components

### 🎯 Workflow Patterns
- Fan-out/fan-in for parallel processing
- Multi-stage pipelines with inter-stage channels
- Timer-based coordination
- Hierarchical composition with sub-DAGs

---

## Workflow Types

### 1. Sequential Workflows

Tasks execute one after another, passing results forward.

```karl
let workflow = {
    name: "ETL Pipeline",
    type: "sequential",
    tasks: [
        {
            name: "Extract",
            handler: (ctx) -> {
                success: true,
                data: { records: 1000 }
            },
        },
        {
            name: "Transform",
            handler: (ctx) -> {
                // ctx contains result from Extract
                success: true,
                data: { processed: ctx.data.records * 2 }
            },
        },
        {
            name: "Load",
            handler: (ctx) -> {
                success: true,
                data: ctx.data
            },
        },
    ],
}

engine.execute(workflow, {}, {})
```

**When to use:** Simple step-by-step processes where each step needs the previous result.

---

### 2. Parallel Workflows

All tasks run concurrently and results are collected.

```karl
let workflow = {
    name: "Multi-API Fetch",
    type: "parallel",
    tasks: [
        {
            name: "Fetch Users",
            handler: (ctx) -> fetchAPI("/users"),
        },
        {
            name: "Fetch Products",
            handler: (ctx) -> fetchAPI("/products"),
        },
        {
            name: "Fetch Orders",
            handler: (ctx) -> fetchAPI("/orders"),
        },
    ],
}
```

**When to use:** Independent operations that can run simultaneously.

---

### 3. DAG Workflows (Directed Acyclic Graph)

Tasks execute based on dependencies, with parallel execution where possible.

```karl
let workflow = {
    name: "Data Processing DAG",
    type: "dag",
    nodes: [
        { id: "fetch-users", name: "Fetch Users", handler: (ctx) -> {...} },
        { id: "fetch-orders", name: "Fetch Orders", handler: (ctx) -> {...} },
        { id: "merge", name: "Merge Data", handler: (ctx) -> {...} },
        { id: "analyze", name: "Analyze", handler: (ctx) -> {...} },
    ],
    edges: [
        { source: "fetch-users", target: "merge" },
        { source: "fetch-orders", target: "merge" },
        { source: "merge", target: "analyze" },
    ],
}
// fetch-users and fetch-orders run in parallel
// merge waits for both to complete
// analyze waits for merge
```

**When to use:** Complex workflows where some tasks depend on others.

---

### 4. Timer Tasks

Execute tasks after a delay or at intervals.

```karl
// Delayed execution
let delayedWorkflow = {
    name: "Delayed Notification",
    type: "timer",
    task: {
        name: "Send Email",
        delay: 5000,  // Wait 5 seconds
        handler: (ctx) -> {
            sendEmail()
            { success: true, data: "Email sent" }
        },
    },
}

// Interval execution (repeated)
let intervalWorkflow = {
    name: "Health Monitor",
    type: "interval",
    task: {
        name: "Check Health",
        interval: 1000,        // Every 1 second
        maxRepetitions: 10,    // Run 10 times
        handler: (ctx) -> {
            log("Health check #", ctx.iteration)
            checkSystemHealth()
            { success: true, data: { iteration: ctx.iteration } }
        },
    },
}
```

**When to use:**
- **Timer:** Scheduled notifications, deferred processing, rate limiting
- **Interval:** Health checks, monitoring, polling, periodic updates

---

### 5. Sub-DAGs (Reusable Components)

Create workflow components once and reuse them anywhere.

```karl
// Define reusable validation component
let validationSubDAG = createSubDAG(
    "validation",
    "Data Validation",
    [
        { id: "check-format", name: "Format", handler: validateFormat },
        { id: "check-schema", name: "Schema", handler: validateSchema },
        { id: "check-integrity", name: "Integrity", handler: checkIntegrity },
    ],
    [
        { source: "check-format", target: "check-schema" },
        { source: "check-schema", target: "check-integrity" },
    ]
)

// Use in a larger workflow
let workflow = {
    name: "ETL with Validation",
    type: "dag-with-subdags",
    nodes: [
        { id: "extract", name: "Extract", handler: extractData },
        validationSubDAG,  // <- Reusable component!
        { id: "load", name: "Load", handler: loadData },
    ],
    edges: [
        { source: "extract", target: "validation" },
        { source: "validation", target: "load" },
    ],
}
```

**When to use:** Modular workflows, shared validation/transformation logic, testing.

---

## Examples

### Basic Examples (`examples.k`)
Demonstrates core workflow patterns:
- ✅ Sequential ETL pipeline
- ✅ Parallel API requests
- ✅ Mathematical pipeline with validation
- ✅ Retry mechanism
- ✅ Parallel data aggregation

```bash
karl run examples.k
```

---

### Timer Tasks (`timer_tasks.k`)
Comprehensive timer demonstrations:
- Delayed task execution
- Interval-based health monitoring  
- Coordinated batch processing
- Multi-timer coordination
- Periodic status updates

```bash
karl run timer_tasks.k
```

**Output shows:**
- 5-second delayed notification
- Health checks running every 300ms for 5 iterations
- Batch processing with timed coordination
- Multiple timers working together

---

### DAG Pipeline (`dag_pipeline.k`)
Advanced multi-stage data processing:
- **Stage 1:** Parallel data fetching (4 sources)
- **Stage 2:** Fan-out to 2 processing paths
- **Stage 3:** Worker pool transformation (4 workers)
- **Stage 4:** Parallel aggregation (4 metrics)
- **Stage 5:** Report generation

```bash
karl run dag_pipeline.k
```

**Output shows:**
- 425 records fetched from 4 APIs
- 388 records processed through pipeline
- Detailed metrics and worker distribution

---

### CSV Pipeline (`csv_pipeline.k`)
CSV data processing workflow:
- File reading and parsing
- Data validation
- Transformation pipeline
- Result aggregation

```bash
karl run csv_pipeline.k
```

---

### File Watcher (`file_watcher.k`)
Reactive file monitoring:
- Event-driven processing
- File system monitoring
- Automated workflows

```bash
karl run file_watcher.k
```

---

### Sub-DAG Demo (`subdag_demo.k`)
Reusable workflow components:
- Single sub-DAG execution
- Sequential sub-DAG pipelines
- Parallel sub-DAG execution
- Nested sub-DAGs (3 levels deep)
- ETL and ML pipeline patterns

```bash
karl run subdag_demo.k
```

**Output shows:**
- Validation pipeline sub-DAG
- ETL pipeline with reusable components
- 3-source parallel processing
- ML workflow with nested preprocessing

---

### Quick Start (`quickstart.k`)
Perfect for beginners:
1. Simple delayed task
2. Interval-based execution  
3. Basic sub-DAG usage
4. Combined timer + sub-DAG

```bash
karl run quickstart.k
```

---

## Configuration

### Engine Configuration

```karl
let config = {
    defaultRetries: 2,        // Retry failed tasks
    defaultWorkers: 3,        // Worker pool size
    stopOnError: false,       // Continue on failures
}

engine.execute(workflow, initialContext, config)
```

### Task Configuration

```karl
let task = {
    name: "Task Name",
    retries: 3,               // Override default
    handler: (ctx) -> {
        // Task logic
        { success: true, data: result }
    },
}
```

---

## Karl Language Features

### Concurrency
```karl
& taskFunction()          // Spawn async task
wait task                 // Wait for completion
rendezvous()             // Create channel
channel.send(value)      // Send to channel
channel.recv()           // Receive from channel
channel.done()           // Close channel
```

### Control Flow
```karl
for i < 10 with i = 0 {
    // Loop body
    i = i + 1
} then result                // Loop returns value

break value                  // Early exit with result

match workflow.type {
    "sequential" => ...,
    "parallel" => ...,
    default => ...
}
```

### Object Methods (New!)
```karl
let obj = {}
obj = obj.set("key", value)  // Set property
let val = obj.get("key")     // Get property
let has = obj.has("key")     // Check existence
```

---

## Architecture Patterns

### Worker Pool
```karl
let workChan = rendezvous()

// Spawn workers
for i < numWorkers with i = 0 {
    & (() -> {
        for processing = true with processing = true {
            let [item, done] = workChan.recv()
            if done { break {} }
            processItem(item)
        }
    })()
    i = i + 1
} then {}

// Distribute work
for i < items.length with i = 0 {
    workChan.send(items[i])
    i = i + 1
} then {}

workChan.done()
```

### Fan-out/Fan-in
```karl
// Fan-out: spawn parallel tasks
let tasks = for i < items.length with i = 0, workers = [] {
    workers += [& processItem(items[i])]
    i = i + 1
} then workers

// Fan-in: collect results
let results = for i < tasks.length with i = 0, collected = [] {
    collected += [wait tasks[i]]
    i = i + 1
} then collected
```

---

## Use Cases

### ETL Pipelines
Sequential data extraction, transformation, and loading with validation steps.

### API Orchestration
Parallel requests to multiple endpoints with result aggregation.

### Data Processing
Multi-stage pipelines with worker pools for high-throughput processing.

### Scheduled Tasks
Delayed notifications, periodic health checks, batch processing.

### Modular Workflows
Reusable components for validation, transformation, and quality assurance.

---

## Error Handling

The engine provides multiple strategies:

1. **Task-level retries** - Automatic retry with configurable attempts
2. **Graceful degradation** - Continue workflow on non-critical failures
3. **Error propagation** - Failed tasks return error information
4. **Stop-on-error** - Optional immediate halt on first failure

---

## File Overview

| File | Lines | Purpose | Status |
|------|-------|---------|--------|
| `engine.k` | ~500 | Core workflow engine | ✅ |
| `examples.k` | ~200 | Basic workflow demos | ✅ |
| `timer_tasks.k` | ~300 | Timer demonstrations | ✅ |
| `dag_pipeline.k` | ~430 | Advanced DAG demo | ✅ |
| `csv_pipeline.k` | ~300 | CSV processing | ✅ |
| `file_watcher.k` | ~300 | File monitoring | ✅ |
| `subdag_demo.k` | ~550 | Sub-DAG patterns | ✅ |
| `quickstart.k` | ~150 | Beginner guide | ✅ |

**Total:** ~2,700 lines of workflow orchestration code

---

## Recent Updates

### 2026-01-30: DAG Executor Fix ✨
- **Fixed deadlock** in DAG executor
- **Added Object methods** to Karl language (`.get()`, `.set()`, `.has()`)
- **Simplified DAG logic** with counting-based approach
- **All DAG tests passing** ✅

### 2026-01-30: Timer & Sub-DAG Features ✨
- **Timer tasks** with delayed and interval execution
- **Sub-DAGs** for reusable workflow components
- **New examples** demonstrating advanced patterns

---

## Future Enhancements

Potential additions:
- Task-level execution timeouts
- Priority queues for task scheduling
- Dynamic routing based on results
- State persistence (save/resume workflows)
- Built-in metrics collection
- Circuit breakers for failure recovery
- Distributed execution across multiple nodes
- Workflow versioning

---

**Author:** Nico  
**Language:** Karl  
**Created:** 2026-01-29  
**Updated:** 2026-01-30  
**Files:** 10  
**Status:** Production Ready ✅
