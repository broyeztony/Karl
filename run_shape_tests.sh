#!/bin/bash

# Karl Shape Feature - Test Runner
# Runs all Karl-based shape tests

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         Karl Shape Feature - Integration Test Suite             ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

TESTS_DIR="tests/shapes"
PASSED=0
FAILED=0
TOTAL=0

# Find all test files
TEST_FILES=$(find "$TESTS_DIR" -maxdepth 1 -name "*.k" -type f | sort)

if [ -z "$TEST_FILES" ]; then
    echo "${RED}No test files found in $TESTS_DIR${NC}"
    exit 1
fi

echo "Running tests from: $TESTS_DIR"
echo ""

for test_file in $TEST_FILES; do
    TOTAL=$((TOTAL + 1))
    test_name=$(basename "$test_file")
    
    echo -ne "${BLUE}Running:${NC} $test_name ... "
    
    # Run the test and capture output
    output=$(go run main.go run "$test_file" 2>&1)
    exit_code=$?
    
    # Check if test passed (exit code 0 and output contains PASS messages)
    if [ $exit_code -eq 0 ]; then
        # Count FAIL messages in output
        fail_count=$(echo "$output" | grep -c "❌ FAIL" || true)
        
        if [ $fail_count -eq 0 ]; then
            echo -e "${GREEN}✓ PASSED${NC}"
            PASSED=$((PASSED + 1))
        else
            echo -e "${RED}✗ FAILED${NC}"
            echo "  Failures detected in test output:"
            echo "$output" | grep "❌ FAIL" | sed 's/^/    /'
            FAILED=$((FAILED + 1))
        fi
    else
        echo -e "${RED}✗ FAILED (exit code: $exit_code)${NC}"
        echo "  Error output:"
        echo "$output" | sed 's/^/    /'
        FAILED=$((FAILED + 1))
    fi
    echo ""
done

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                         Test Summary                             ║"
echo "╠══════════════════════════════════════════════════════════════════╣"
echo "  Total:  $TOTAL"
echo -e "  Passed: ${GREEN}$PASSED${NC}"
echo -e "  Failed: ${RED}$FAILED${NC}"
echo "╚══════════════════════════════════════════════════════════════════╝"

if [ $FAILED -eq 0 ]; then
    echo ""
    echo -e "${GREEN}All tests passed! ✓${NC}"
    exit 0
else
    echo ""
    echo -e "${RED}Some tests failed ✗${NC}"
    exit 1
fi
