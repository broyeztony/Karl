#!/bin/bash

# Karl Shape Feature - Complete Test Suite Runner
# Runs both Go unit tests and Karl integration tests

set -e  # Exit on first failure

echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║         Karl Shape Feature - Complete Test Suite                ║"
echo "╚══════════════════════════════════════════════════════════════════╝"
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Track results
GO_PASSED=false
KARL_PASSED=false

echo -e "${BLUE}[1/2] Running Go Unit Tests...${NC}"
echo ""

# Run Go tests
if go test ./interpreter/... ./shape/... -v; then
    GO_PASSED=true
    echo ""
    echo -e "${GREEN}✓ Go unit tests passed${NC}"
else
    echo ""
    echo -e "${RED}✗ Go unit tests failed${NC}"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""

echo -e "${BLUE}[2/2] Running Karl Integration Tests...${NC}"
echo ""

# Run Karl tests
if ./run_shape_tests.sh; then
    KARL_PASSED=true
else
    echo -e "${RED}✗ Karl integration tests failed${NC}"
    exit 1
fi

echo ""
echo "╔══════════════════════════════════════════════════════════════════╗"
echo "║                     Final Test Summary                          ║"
echo "╠══════════════════════════════════════════════════════════════════╣"

if [ "$GO_PASSED" = true ] && [ "$KARL_PASSED" = true ]; then
    echo -e "║  Go Unit Tests (72):              ${GREEN}✓ PASSED${NC}                     ║"
    echo -e "║  Karl Integration Tests (7):      ${GREEN}✓ PASSED${NC}                     ║"
    echo "║                                                                  ║"
    echo -e "║  ${GREEN}🎉 ALL 79+ TESTS PASSED!${NC}                                      ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    exit 0
else
    echo -e "║  ${RED}✗ SOME TESTS FAILED${NC}                                          ║"
    echo "╚══════════════════════════════════════════════════════════════════╝"
    exit 1
fi
