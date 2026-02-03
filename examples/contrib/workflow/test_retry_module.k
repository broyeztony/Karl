// Minimal test to isolate the retry policy issue

let makeRetry = import "examples/contrib/workflow/retry_policy.k"
let Retry = makeRetry()

log("Creating retry engine...")

let policy = {
    maxAttempts: 3,
    strategy: Retry.RETRY_EXPONENTIAL,
    initialDelay: 100,
    maxDelay: 1000,
    jitterEnabled: false,
    jitterFactor: 0.1,
    retryableErrors: [],
}

let retryEngine = Retry.createRetryEngine(policy)

log("Retry engine created successfully!")

let attempts = 0
let task = {
    name: "TestTask",
    handler: (ctx) -> {
        attempts = attempts + 1
        log("Attempt", attempts)
        if attempts < 2 {
            { success: false, error: "Simulated failure" }
        } else {
            { success: true, data: "Success!" }
        }
    },
}

log("Executing task with retry...")
let result = retryEngine.execute(task, {})

log("Result:", result)
