#!/bin/bash

# ============================================================================
# Karl Workflow Engine - Comprehensive Test Runner
# ============================================================================
# This script runs all tests and demos for the workflow engine
# ============================================================================

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Counters
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Test results array
declare -a FAILED_TEST_NAMES

echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}Karl Workflow Engine - Test Suite${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""

# for Mac users
function timeout() { perl -e 'alarm shift; exec @ARGV' "$@"; }

# Function to run a test
run_test() {
    local test_file=$1
    local test_name=$2
    local timeout_seconds=${3:-60}
    
    TOTAL_TESTS=$((TOTAL_TESTS + 1))
    
    echo -e "${YELLOW}[TEST $TOTAL_TESTS] Running: $test_name${NC}"
    echo "  File: $test_file"
    echo "  Timeout: ${timeout_seconds}s"
    echo ""
    
    if timeout ${timeout_seconds} karl run "$test_file" > /tmp/karl_test_output.log 2>&1; then
        echo -e "${GREEN}✅ PASSED${NC}"
        PASSED_TESTS=$((PASSED_TESTS + 1))
    else
        echo -e "${RED}❌ FAILED${NC}"
        FAILED_TESTS=$((FAILED_TESTS + 1))
        FAILED_TEST_NAMES+=("$test_name")
        echo "  Error output (last 20 lines):"
        tail -20 /tmp/karl_test_output.log | sed 's/^/    /'
    fi
    echo ""
    echo "---"
    echo ""
}

# ============================================================================
# UNIT TESTS
# ============================================================================

echo -e "${BLUE}=== UNIT TESTS ===${NC}"
echo ""

run_test "test_retry_module.k" "Retry Policy Module Tests" 30
run_test "test_retry.k" "Retry Strategy Tests" 30
run_test "test_sequential.k" "Sequential Execution Tests" 30
run_test "test_pipeline.k" "Pipeline Execution Tests" 30
run_test "test_dag.k" "DAG Execution Tests" 60

# ============================================================================
# INTEGRATION TESTS
# ============================================================================

echo -e "${BLUE}=== INTEGRATION TESTS ===${NC}"
echo ""

run_test "test_integrated_features.k" "Integrated Features Tests" 90

# ============================================================================
# DEMONSTRATION SCRIPTS
# ============================================================================

echo -e "${BLUE}=== DEMONSTRATION SCRIPTS ===${NC}"
echo ""

run_test "quickstart.k" "Quickstart Demo" 30
run_test "examples.k" "Basic Examples" 60
run_test "enhanced_demo.k" "Enhanced Features Demo" 120
run_test "subdag_demo.k" "Sub-DAG Demo" 90
run_test "timer_tasks.k" "Timer Tasks Demo" 60

# ============================================================================
# PIPELINE EXAMPLES
# ============================================================================

echo -e "${BLUE}=== PIPELINE EXAMPLES ===${NC}"
echo ""

run_test "dag_pipeline.k" "DAG Pipeline Example" 60
run_test "csv_pipeline.k" "CSV Pipeline Example" 60

# Note: file_watcher.k is skipped as it's an interactive/long-running demo

# ============================================================================
# SUMMARY
# ============================================================================

echo ""
echo -e "${BLUE}============================================================================${NC}"
echo -e "${BLUE}TEST SUMMARY${NC}"
echo -e "${BLUE}============================================================================${NC}"
echo ""
echo "Total Tests:  $TOTAL_TESTS"
echo -e "${GREEN}Passed:       $PASSED_TESTS${NC}"
if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed:       $FAILED_TESTS${NC}"
else
    echo "Failed:       $FAILED_TESTS"
fi
echo ""

if [ $FAILED_TESTS -gt 0 ]; then
    echo -e "${RED}Failed Tests:${NC}"
    for test_name in "${FAILED_TEST_NAMES[@]}"; do
        echo "  - $test_name"
    done
    echo ""
    exit 1
else
    echo -e "${GREEN}🎉 All tests passed!${NC}"
    echo ""
    exit 0
fi
