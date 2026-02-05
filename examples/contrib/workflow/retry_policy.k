// ============================================================================
// RETRY POLICY - Resilient Task Execution with Back-off
// ============================================================================
// Provides configurable retry strategies for workflow tasks:
// - Exponential back-off
// - Linear back-off
// - Fixed delay retry
// - Custom retry predicates
// - Jitter support for distributed systems
// ============================================================================

// ----------------------------------------------------------------------------
// RETRY STRATEGY TYPES
// ----------------------------------------------------------------------------

// Fixed delay - same delay between all retries
let RETRY_FIXED = "fixed"

// Linear back-off - delay increases linearly (delay * attempt)
let RETRY_LINEAR = "linear"

// Exponential back-off - delay doubles each attempt (delay * 2^attempt)
let RETRY_EXPONENTIAL = "exponential"

// ----------------------------------------------------------------------------
// RETRY POLICY CONFIGURATION
// ----------------------------------------------------------------------------

let defaultRetryPolicy = {
    maxAttempts: 3,
    strategy: RETRY_EXPONENTIAL,
    initialDelay: 1000,      // 1 second
    maxDelay: 30000,         // 30 seconds
    jitterEnabled: true,
    jitterFactor: 0.1,       // 10% random jitter
    retryableErrors: [],     // Empty = retry all errors
}

// ----------------------------------------------------------------------------
// DELAY CALCULATION
// ----------------------------------------------------------------------------

// Calculate delay for a given attempt based on strategy
let calculateDelay = (attempt, policy) -> {
    let baseDelay = match policy.strategy {
        case "fixed" -> policy.initialDelay
        case "linear" -> policy.initialDelay * attempt
        case "exponential" -> {
            // Calculate 2^(attempt-1) * initialDelay
            let power = attempt - 1
            let multiplier = for i < power with i = 0, result = 1 {
                result = result * 2
                i = i + 1
            } then result
            
            policy.initialDelay * multiplier
        }
        case _ -> policy.initialDelay
    }
    
    // Cap at maxDelay
    let cappedDelay = if baseDelay > policy.maxDelay {
        policy.maxDelay
    } else {
        baseDelay
    }
    
    // Add jitter if enabled
    if policy.jitterEnabled {
        // Use integer arithmetic for jitter (convert factor to percentage)
        let jitterPercent = (policy.jitterFactor * 100)  // 0.1 -> 10
        let jitterRange = (cappedDelay * jitterPercent) / 100  // Integer division
        let randomValue = now() % 1000  // 0-999
        let jitter = ((randomValue * jitterRange * 2) / 1000) - jitterRange
        let finalDelay = cappedDelay + jitter
        
        // Ensure non-negative and integer
        if finalDelay < 0 { 0 } else { finalDelay }
    } else {
        cappedDelay
    }
}

// ----------------------------------------------------------------------------
// ERROR CLASSIFICATION
// ----------------------------------------------------------------------------

// Check if an error is retryable based on policy
let isRetryableError = (error, policy) -> {
    // If no specific retryable errors defined, retry all
    if policy.retryableErrors.length == 0 {
        true
    } else {
        // Check if error matches any retryable error pattern
        for i < policy.retryableErrors.length with i = 0, isRetryable = false {
            let pattern = policy.retryableErrors[i]
            
            // Simple string matching for now
            if error == pattern {
                isRetryable = true
                break isRetryable
            }
            
            i = i + 1
        } then isRetryable
    }
}

// ----------------------------------------------------------------------------
// RETRY EXECUTION
// ----------------------------------------------------------------------------

// Execute a task with retry logic
let executeWithRetry = (task, context, policy) -> {
    log("[RETRY] Starting task:", task.name)
    log("[RETRY] Max attempts:", policy.maxAttempts)
    log("[RETRY] Strategy:", policy.strategy)
    
    for true with attempt = 1, lastResult = null, lastError = null {
        if attempt > policy.maxAttempts {
            log("[RETRY]", task.name, "- Failed after", policy.maxAttempts, "attempts")
            break {
                success: false,
                error: lastError,
                attempts: attempt - 1,
                exhausted: true,
            }
        }
        
        log("[RETRY]", task.name, "- Attempt", attempt, "/", policy.maxAttempts)
        
        // Execute task
        let result = task.handler(context) ? {
            { success: false, error: "Task execution failed" }
        }
        
        if result.success {
            log("[RETRY]", task.name, "- Success on attempt", attempt)
            break {
                success: true,
                data: result.data,
                attempts: attempt,
            }
        } else {
            lastResult = result
            lastError = if result.error != null { result.error } else { "Unknown error" }
            
            // Check if error is retryable
            if !isRetryableError(lastError, policy) {
                log("[RETRY]", task.name, "- Non-retryable error:", lastError)
                break {
                    success: false,
                    error: lastError,
                    attempts: attempt,
                    nonRetryable: true,
                }
            }
            
            // Calculate delay for next attempt
            if attempt < policy.maxAttempts {
                let delay = calculateDelay(attempt, policy)
                log("[RETRY]", task.name, "- Waiting", delay, "ms before retry")
                sleep(delay)
            }
        }
        
        attempt = attempt + 1
    } then lastResult
}

// ----------------------------------------------------------------------------
// RETRY POLICY BUILDERS
// ----------------------------------------------------------------------------

// Create a fixed delay retry policy
let createFixedRetryPolicy = (maxAttempts, delay) -> {
    {
        maxAttempts: maxAttempts,
        strategy: RETRY_FIXED,
        initialDelay: delay,
        maxDelay: delay,
        jitterEnabled: false,
        jitterFactor: 0,
        retryableErrors: [],
    }
}

// Create a linear back-off retry policy
let createLinearRetryPolicy = (maxAttempts, initialDelay, maxDelay) -> {
    {
        maxAttempts: maxAttempts,
        strategy: RETRY_LINEAR,
        initialDelay: initialDelay,
        maxDelay: maxDelay,
        jitterEnabled: false,
        jitterFactor: 0.1,
        retryableErrors: [],
    }
}

// Create an exponential back-off retry policy
let createExponentialRetryPolicy = (maxAttempts, initialDelay, maxDelay) -> {
    {
        maxAttempts: maxAttempts,
        strategy: RETRY_EXPONENTIAL,
        initialDelay: initialDelay,
        maxDelay: maxDelay,
        jitterEnabled: false,
        jitterFactor: 0.1,
        retryableErrors: [],
    }
}

// Create a custom retry policy
let createCustomRetryPolicy = (config) -> {
    {
        maxAttempts: if config.maxAttempts != null { config.maxAttempts } else { defaultRetryPolicy.maxAttempts },
        strategy: if config.strategy != null { config.strategy } else { defaultRetryPolicy.strategy },
        initialDelay: if config.initialDelay != null { config.initialDelay } else { defaultRetryPolicy.initialDelay },
        maxDelay: if config.maxDelay != null { config.maxDelay } else { defaultRetryPolicy.maxDelay },
        jitterEnabled: if config.jitterEnabled != null { config.jitterEnabled } else { defaultRetryPolicy.jitterEnabled },
        jitterFactor: if config.jitterFactor != null { config.jitterFactor } else { defaultRetryPolicy.jitterFactor },
        retryableErrors: if config.retryableErrors != null { config.retryableErrors } else { defaultRetryPolicy.retryableErrors },
    }
}

// ----------------------------------------------------------------------------
// CIRCUIT BREAKER PATTERN
// ----------------------------------------------------------------------------

// Circuit breaker states
let CIRCUIT_CLOSED = "closed"       // Normal operation
let CIRCUIT_OPEN = "open"           // Failing, reject requests
let CIRCUIT_HALF_OPEN = "half_open" // Testing if service recovered

// Create a circuit breaker
let createCircuitBreaker = (config) -> {
    let state = {
        status: CIRCUIT_CLOSED,
        failureCount: 0,
        successCount: 0,
        lastFailureTime: 0,
        threshold: if config.threshold != null { config.threshold } else { 5 },
        timeout: if config.timeout != null { config.timeout } else { 60000 },  // 1 minute
        halfOpenAttempts: if config.halfOpenAttempts != null { config.halfOpenAttempts } else { 3 },
    }
    
    {
        state: state,
        
        // Execute task through circuit breaker
        execute: (task, context) -> {
            let currentTime = now()
            
            // Check circuit state
            if state.status == CIRCUIT_OPEN {
                // Check if timeout has passed
                if currentTime - state.lastFailureTime > state.timeout {
                    log("[CIRCUIT] Moving to HALF_OPEN state")
                    state.status = CIRCUIT_HALF_OPEN
                    state.successCount = 0
                } else {
                    log("[CIRCUIT] Circuit is OPEN, rejecting request")
                    { success: false, error: "Circuit breaker is open", circuitOpen: true }
                }
            }
            
            // Execute task
            let result = task.handler(context) ? {
                { success: false, error: "Task execution failed" }
            }
            
            if result.success {
                state.successCount = state.successCount + 1
                state.failureCount = 0
                
                // If in half-open, check if we can close the circuit
                if state.status == CIRCUIT_HALF_OPEN {
                    if state.successCount >= state.halfOpenAttempts {
                        log("[CIRCUIT] Moving to CLOSED state")
                        state.status = CIRCUIT_CLOSED
                    }
                }
                
                result
            } else {
                state.failureCount = state.failureCount + 1
                state.lastFailureTime = currentTime
                state.successCount = 0
                
                // Check if we should open the circuit
                if state.failureCount >= state.threshold {
                    log("[CIRCUIT] Threshold reached, moving to OPEN state")
                    state.status = CIRCUIT_OPEN
                }
                
                result
            }
        },
        
        // Reset circuit breaker
        reset: () -> {
            log("[CIRCUIT] Resetting circuit breaker")
            state.status = CIRCUIT_CLOSED
            state.failureCount = 0
            state.successCount = 0
            state.lastFailureTime = 0
        },
        
        // Get current state
        getState: () -> state,
    }
}

// ----------------------------------------------------------------------------
// RETRY POLICY API
// ----------------------------------------------------------------------------

// Create a retry engine instance
let createRetryEngine = (policy) -> {
    let finalPolicy = if policy != null {
        createCustomRetryPolicy(policy)
    } else {
        defaultRetryPolicy
    }
    
    {
        policy: finalPolicy,
        execute: (task, context) -> executeWithRetry(task, context, finalPolicy),
        calculateDelay: (attempt) -> calculateDelay(attempt, finalPolicy),
        isRetryable: (error) -> isRetryableError(error, finalPolicy),
        
        // Policy builders
        fixed: (maxAttempts, delay) -> createRetryEngine(createFixedRetryPolicy(maxAttempts, delay)),
        linear: (maxAttempts, initialDelay, maxDelay) -> createRetryEngine(createLinearRetryPolicy(maxAttempts, initialDelay, maxDelay)),
        exponential: (maxAttempts, initialDelay, maxDelay) -> createRetryEngine(createExponentialRetryPolicy(maxAttempts, initialDelay, maxDelay)),
    }
}

// ----------------------------------------------------------------------------
// MODULE EXPORT
// ----------------------------------------------------------------------------

// Export module as a function (Karl module pattern)
let makeModule = () -> {
    {
        // Constants
        RETRY_FIXED: RETRY_FIXED,
        RETRY_LINEAR: RETRY_LINEAR,
        RETRY_EXPONENTIAL: RETRY_EXPONENTIAL,
        CIRCUIT_CLOSED: CIRCUIT_CLOSED,
        CIRCUIT_OPEN: CIRCUIT_OPEN,
        CIRCUIT_HALF_OPEN: CIRCUIT_HALF_OPEN,
        
        // Functions
        createRetryEngine: createRetryEngine,
        createCircuitBreaker: createCircuitBreaker,
        executeWithRetry: executeWithRetry,
        calculateDelay: calculateDelay,
        isRetryableError: isRetryableError,
        
        // Policy builders
        createFixedRetryPolicy: createFixedRetryPolicy,
        createLinearRetryPolicy: createLinearRetryPolicy,
        createExponentialRetryPolicy: createExponentialRetryPolicy,
        createCustomRetryPolicy: createCustomRetryPolicy,
    }
}

makeModule
