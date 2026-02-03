# Workflow Engine Tests

This directory contains test files for the Karl workflow engine.

## Test Files

### Core Functionality Tests

- **test_sequential.k** - Tests basic sequential task execution
  - Simple workflow with one task
  - Validates basic engine functionality
  - ✅ PASSING

- **test_retry.k** - Tests retry policy through the workflow engine
  - Sequential workflow with retry policy
  - Exponential backoff strategy
  - ✅ PASSING

- **test_dag.k** - Tests DAG (Directed Acyclic Graph) execution
  - Two tasks with dependency (task-a → task-b)
  - Validates dependency resolution
  - ✅ PASSING

- **test_retry_module.k** - Tests retry module directly
  - Direct retry engine usage (not through workflow engine)
  - Simulates flaky task with failures
  - ✅ PASSING

### Integration Tests

- **test_integrated_features.k** - Comprehensive integration test
  - Tests retry policy integration
  - Tests worker pool integration (⚠️ KNOWN DEADLOCK)
  - Tests DAG with persistence
  - Tests complex workflows
  - ⚠️ PARTIAL - Worker pool has known deadlock issue

## Running Tests

Run individual tests:
```bash
karl run examples/contrib/workflow/test_sequential.k
karl run examples/contrib/workflow/test_retry.k
karl run examples/contrib/workflow/test_dag.k
karl run examples/contrib/workflow/test_retry_module.k
```

Run all basic tests:
```bash
for test in test_sequential test_retry test_dag test_retry_module; do
  echo "Running $test..."
  karl run examples/contrib/workflow/$test.k
done
```

## Known Issues

### Worker Pool Deadlock
The parallel executor (worker pool) has a known deadlock issue where goroutines get stuck in channel send operations. This affects:
- `test_integrated_features.k` (TEST 2: Worker Pool Integration)
- Any workflow using `useWorkerPool: true`

**Status**: Tracked separately, fix coming soon

**Workaround**: Use sequential execution or legacy parallel mode (without worker pool)

## Test Status Summary

| Test | Status | Notes |
|------|--------|-------|
| test_sequential.k | ✅ PASSING | Basic sequential execution |
| test_retry.k | ✅ PASSING | Retry through workflow engine |
| test_dag.k | ✅ PASSING | DAG dependency resolution |
| test_retry_module.k | ✅ PASSING | Direct retry module usage |
| test_integrated_features.k | ⚠️ PARTIAL | Worker pool deadlock in TEST 2 |

## Recent Updates

- **2026-02-03**: Upgraded all tests to use truthy/falsy semantics
- **2026-02-03**: Upgraded DAG execution to use `map()` for dynamic storage
- **2026-02-03**: Cleaned up duplicate test files
- **2026-02-03**: Renamed tests for clarity (removed "simple" and "minimal" prefixes)
