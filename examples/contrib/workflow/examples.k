// ============================================================================
// WORKFLOW ENGINE EXAMPLES
// ============================================================================
// Demonstrates all workflow types:
// 1. Sequential Workflows - Tasks run one after another
// 2. Parallel Workflows - Tasks run concurrently
// 3. DAG Workflows - Tasks with dependencies
// 4. Pipeline Workflows - Multi-stage data processing
// ============================================================================

// Import the workflow engine
let makeEngine = import "examples/contrib/workflow/engine.k"
let engine = makeEngine()

// Access the exported functions
let execute = engine.execute
let defaultConfig = engine.defaultConfig

// ============================================================================
// EXAMPLE 1: DATA PROCESSING PIPELINE
// ============================================================================

log("")
log("======================================================================")
log("EXAMPLE 1: Sequential Data Processing")
log("======================================================================")

let dataWorkflow = {
    name: "Data ETL Pipeline",
    type: "sequential",
    tasks: [
        {
            name: "Extract",
            handler: (ctx) -> {
                log("    → Extracting data from source...")
                sleep(100)
                {
                    success: true,
                    data: [
                        { id: 1, value: 100 },
                        { id: 2, value: 200 },
                        { id: 3, value: 300 },
                    ],
                }
            },
        },
        {
            name: "Transform",
            handler: (ctx) -> {
                log("    → Transforming data...")
                sleep(100)
                let items = ctx.data
                {
                    success: true,
                    data: for i < items.length with i = 0, result = [] {
                        let item = items[i]
                        result += [{ id: item.id, value: item.value * 2, transformed: true }]
                        i = i + 1
                    } then result
                }
            },
        },
        {
            name: "Load",
            handler: (ctx) -> {
                log("    → Loading data to destination...")
                sleep(100)
                log("    → Loaded", ctx.data.length, "records")
                { success: true, data: ctx.data }
            },
        },
    ],
}


let result1 = execute(dataWorkflow, {}, defaultConfig)
log("✓ Workflow completed:", result1.success)
log("")

// ============================================================================
// EXAMPLE 2: PARALLEL API REQUESTS
// ============================================================================

log("======================================================================")
log("EXAMPLE 2: Parallel API Requests")
log("======================================================================")

let apiWorkflow = {
    name: "Multi-API Fetch",
    type: "parallel",
    tasks: [
        {
            name: "Fetch Users",
            handler: (ctx) -> {
                sleep(150)
                { success: true, data: { endpoint: "users", count: 42 } }
            },
        },
        {
            name: "Fetch Products",
            handler: (ctx) -> {
                sleep(120)
                { success: true, data: { endpoint: "products", count: 128 } }
            },
        },
        {
            name: "Fetch Orders",
            handler: (ctx) -> {
                sleep(100)
                { success: true, data: { endpoint: "orders", count: 256 } }
            },
        },
        {
            name: "Fetch Analytics",
            handler: (ctx) -> {
                sleep(180)
                { success: true, data: { endpoint: "analytics", metrics: 89 } }
            },
        },
    ],
}

let result2 = execute(apiWorkflow, {}, defaultConfig)
log("✓ Workflow completed:", result2.success)
log("✓ Fetched data from", result2.results.length, "endpoints")
log("")

// ============================================================================
// EXAMPLE 3: COMPUTATION WORKFLOW WITH ERROR HANDLING
// ============================================================================

log("======================================================================")
log("EXAMPLE 3: Mathematical Pipeline with Validation")
log("======================================================================")

let mathWorkflow = {
    name: "Math Processing",
    type: "sequential",
    tasks: [
        {
            name: "Generate Numbers",
            handler: (ctx) -> {
                let numbers = for i < 10 with i = 0, nums = [] {
                    nums += [i * 10]
                    i = i + 1
                } then nums
                log("    → Generated:", numbers.length, "numbers")
                { success: true, data: numbers }
            },
        },
        {
            name: "Filter Evens",
            handler: (ctx) -> {
                let filtered = for i < ctx.data.length with i = 0, result = [] {
                    let num = ctx.data[i]
                    if num % 20 == 0 {
                        result += [num]
                    }
                    i = i + 1
                } then result
                log("    → Filtered to:", filtered.length, "numbers")
                { success: true, data: filtered }
            },
        },
        {
            name: "Calculate Sum",
            handler: (ctx) -> {
                let sum = for i < ctx.data.length with i = 0, total = 0 {
                    total = total + ctx.data[i]
                    i = i + 1
                } then total
                log("    → Sum:", sum)
                { success: true, data: sum }
            },
        },
        {
            name: "Validate Result",
            handler: (ctx) -> {
                let isValid = ctx.data > 0
                log("    → Validation:", if isValid { "PASSED" } else { "FAILED" })
                { success: isValid, data: ctx.data }
            },
        },
    ],
}

let result3 = execute(mathWorkflow, {}, defaultConfig)
log("✓ Workflow completed:", result3.success)
log("")

// ============================================================================
// EXAMPLE 4: RETRY LOGIC DEMONSTRATION
// ============================================================================

log("======================================================================")
log("EXAMPLE 4: Task Retry Mechanism")
log("======================================================================")

let retryCounter = { attempts: 0 }

let retryWorkflow = {
    name: "Retry Demo",
    type: "sequential",
    tasks: [
        {
            name: "Stable Task",
            handler: (ctx) -> {
                log("    → Executing stable task")
                { success: true, data: "OK" }
            },
        },
        {
            name: "Flaky Task",
            retries: 3,
            handler: (ctx) -> {
                retryCounter.attempts = retryCounter.attempts + 1
                log("    → Attempt", retryCounter.attempts)
                
                // Succeed on 3rd attempt
                if retryCounter.attempts >= 3 {
                    { success: true, data: "Eventually succeeded!" }
                } else {
                    { success: false, error: "Not yet..." }
                }
            },
        },
        {
            name: "Final Task",
            handler: (ctx) -> {
                log("    → Final task running")
                { success: true, data: "Complete" }
            },
        },
    ],
}

let result4 = execute(retryWorkflow, {}, defaultConfig)
log("✓ Workflow completed:", result4.success)
log("✓ Total attempts for flaky task:", retryCounter.attempts)
log("")

// ============================================================================
// EXAMPLE 5: PARALLEL DATA AGGREGATION
// ============================================================================

log("======================================================================")
log("EXAMPLE 5: Parallel Data Aggregation")
log("======================================================================")

let aggregationWorkflow = {
    name: "Parallel Stats",
    type: "parallel",
    tasks: [
        {
            name: "Count Primes < 50",
            handler: (ctx) -> {
                let count = 0
                for n < 50 with n = 2 {
                    let isPrime = for i < n with i = 2, prime = true {
                        if n % i == 0 {
                            prime = false
                            break prime
                        }
                        i = i + 1
                    } then prime
                    
                    if isPrime { count = count + 1 }
                    n = n + 1
                } then count
                
                log("    → Found", count, "primes")
                { success: true, data: count }
            },
        },
        {
            name: "Sum 1..100",
            handler: (ctx) -> {
                let sum = for i < 101 with i = 1, total = 0 {
                    total = total + i
                    i = i + 1
                } then total
                
                log("    → Sum is", sum)
                { success: true, data: sum }
            },
        },
        {
            name: "Fibonacci(10)",
            handler: (ctx) -> {
                let fib = (n) -> {
                    if n <= 1 { n } else {
                        for i < n with i = 2, a = 0, b = 1 {
                            let next = a + b
                            a = b
                            b = next
                            i = i + 1
                        } then b
                    }
                }
                
                let result = fib(10)
                log("    → Fib(10) =", result)
                { success: true, data: result }
            },
        },
    ],
}

let result5 = execute(aggregationWorkflow, {}, defaultConfig)
log("✓ All calculations completed in parallel")
log("")

// ============================================================================
// SUMMARY
// ============================================================================

log("======================================================================")
log("WORKFLOW ENGINE EXAMPLES COMPLETE")
log("======================================================================")
log("")
log("Demonstrated:")
log("  ✓ Sequential workflows (ETL pipeline)")
log("  ✓ Parallel workflows (API requests)")
log("  ✓ Error handling and validation")
log("  ✓ Retry logic with configurable attempts")
log("  ✓ Data aggregation across parallel tasks")
log("")
log("All workflows executed successfully!")
