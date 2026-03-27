#!/bin/bash
#
# Integration Test Script for Phase 3 Week 1
# Tests all VM endpoints and verifies backward compatibility
#
# Usage: ./scripts/test-phase3-week1.sh
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BACKEND_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)/apps/backend"
BASE_URL="${BASE_URL:-http://localhost:8080}"
TEST_TIMEOUT="${TEST_TIMEOUT:-30}"

# Counters
TESTS_PASSED=0
TESTS_FAILED=0

# Helper functions
log_info() {
    echo -e "${YELLOW}[INFO]${NC} $1"
}

log_pass() {
    echo -e "${GREEN}[PASS]${NC} $1"
    ((TESTS_PASSED++))
}

log_fail() {
    echo -e "${RED}[FAIL]${NC} $1"
    ((TESTS_FAILED++))
}

log_section() {
    echo ""
    echo "========================================"
    echo "$1"
    echo "========================================"
}

# Check if backend is running
check_backend() {
    log_info "Checking if backend is running at $BASE_URL..."
    if curl -s --max-time "$TEST_TIMEOUT" "$BASE_URL/api/health" > /dev/null 2>&1; then
        log_pass "Backend is running"
        return 0
    else
        log_fail "Backend is not running at $BASE_URL"
        return 1
    fi
}

# Test health endpoint
test_health() {
    log_info "Testing GET /api/health..."
    RESPONSE=$(curl -s --max-time "$TEST_TIMEOUT" "$BASE_URL/api/health")
    if echo "$RESPONSE" | grep -q "healthy\|ok\|status"; then
        log_pass "Health endpoint returned valid response"
    else
        log_fail "Health endpoint returned unexpected response: $RESPONSE"
    fi
}

# Test VM endpoints (requires authentication)
# Note: These tests assume you have a valid auth token
test_vm_endpoints() {
    log_section "VM Endpoint Tests"
    
    # Get auth token from environment or skip
    AUTH_TOKEN="${AUTH_TOKEN:-}"
    if [ -z "$AUTH_TOKEN" ]; then
        log_info "AUTH_TOKEN not set, skipping authenticated VM tests"
        log_info "To run VM tests, set AUTH_TOKEN environment variable"
        return 0
    fi

    HEADERS="Authorization: Bearer $AUTH_TOKEN"

    # Test GET /api/vms (list VMs)
    log_info "Testing GET /api/vms..."
    RESPONSE=$(curl -s --max-time "$TEST_TIMEOUT" -H "$HEADERS" "$BASE_URL/api/vms")
    HTTP_CODE=$(curl -s --max-time "$TEST_TIMEOUT" -o /dev/null -w "%{http_code}" -H "$HEADERS" "$BASE_URL/api/vms")
    if [ "$HTTP_CODE" -eq 200 ]; then
        log_pass "GET /api/vms returned 200"
    else
        log_fail "GET /api/vms returned $HTTP_CODE (expected 200)"
    fi

    # Test POST /api/vms (create VM) - optional, requires valid request body
    log_info "Testing POST /api/vms (create VM)..."
    RESPONSE=$(curl -s --max-time "$TEST_TIMEOUT" \
        -X POST \
        -H "$HEADERS" \
        -H "Content-Type: application/json" \
        -d '{"name":"test-vm-integration","os":"ubuntu-2204","tier":"micro"}' \
        "$BASE_URL/api/vms")
    HTTP_CODE=$(curl -s --max-time "$TEST_TIMEOUT" -o /dev/null -w "%{http_code}" \
        -X POST \
        -H "$HEADERS" \
        -H "Content-Type: application/json" \
        -d '{"name":"test-vm-integration","os":"ubuntu-2204","tier":"micro"}' \
        "$BASE_URL/api/vms")
    if [ "$HTTP_CODE" -eq 202 ] || [ "$HTTP_CODE" -eq 400 ] || [ "$HTTP_CODE" -eq 403 ]; then
        log_pass "POST /api/vms returned expected status $HTTP_CODE"
    else
        log_fail "POST /api/vms returned unexpected status $HTTP_CODE"
    fi
}

# Test backward compatibility
test_backward_compatibility() {
    log_section "Backward Compatibility Tests"
    
    log_info "Checking API response structure..."
    
    # Check that responses contain expected fields
    RESPONSE=$(curl -s --max-time "$TEST_TIMEOUT" "$BASE_URL/api/health")
    
    # Health endpoint should return JSON
    if echo "$RESPONSE" | python3 -c "import sys,json; json.load(sys.stdin)" 2>/dev/null; then
        log_pass "Health endpoint returns valid JSON"
    else
        log_fail "Health endpoint does not return valid JSON"
    fi
}

# Run Go unit tests
run_unit_tests() {
    log_section "Unit Tests"
    
    log_info "Running Go unit tests in $BACKEND_DIR..."
    cd "$BACKEND_DIR"
    
    if go test ./internal/usecase/... -v -count=1 2>&1 | tee /tmp/test-output.txt; then
        log_pass "All unit tests passed"
    else
        log_fail "Some unit tests failed"
        cat /tmp/test-output.txt
    fi
    
    cd - > /dev/null
}

# Verify build
verify_build() {
    log_section "Build Verification"
    
    log_info "Verifying Go build..."
    cd "$BACKEND_DIR"
    
    if go build ./... 2>&1; then
        log_pass "Build successful (go build ./...)"
    else
        log_fail "Build failed"
        return 1
    fi
    
    cd - > /dev/null
}

# Check code quality metrics
check_code_quality() {
    log_section "Code Quality Checks"
    
    log_info "Checking handler line counts..."
    
    HANDLER_FILE="$BACKEND_DIR/handler/vm_handler.go"
    if [ -f "$HANDLER_FILE" ]; then
        LINE_COUNT=$(wc -l < "$HANDLER_FILE")
        if [ "$LINE_COUNT" -lt 100 ]; then
            log_pass "Handler file has $LINE_COUNT lines (< 100)"
        else
            log_fail "Handler file has $LINE_COUNT lines (should be < 100)"
        fi
    else
        log_fail "Handler file not found: $HANDLER_FILE"
    fi
    
    log_info "Checking for global db variable..."
    if grep -r "^var db " "$BACKEND_DIR/cmd/" 2>/dev/null; then
        log_fail "Global db variable found in cmd/"
    else
        log_pass "No global db variable in cmd/"
    fi
    
    log_info "Checking usecase layer exists..."
    if [ -d "$BACKEND_DIR/internal/usecase" ]; then
        USECASE_FILES=$(ls "$BACKEND_DIR/internal/usecase"/*.go 2>/dev/null | wc -l)
        if [ "$USECASE_FILES" -gt 0 ]; then
            log_pass "Usecase layer exists with $USECASE_FILES files"
        else
            log_fail "No files in usecase layer"
        fi
    else
        log_fail "Usecase layer directory not found"
    fi
    
    log_info "Checking repository layer exists..."
    if [ -d "$BACKEND_DIR/internal/repository" ]; then
        REPO_FILES=$(ls "$BACKEND_DIR/internal/repository"/*.go 2>/dev/null | wc -l)
        if [ "$REPO_FILES" -gt 0 ]; then
            log_pass "Repository layer exists with $REPO_FILES files"
        else
            log_fail "No files in repository layer"
        fi
    else
        log_fail "Repository layer directory not found"
    fi
}

# Print summary
print_summary() {
    log_section "Test Summary"
    echo ""
    echo -e "Tests Passed: ${GREEN}$TESTS_PASSED${NC}"
    echo -e "Tests Failed: ${RED}$TESTS_FAILED${NC}"
    echo ""
    
    if [ "$TESTS_FAILED" -eq 0 ]; then
        echo -e "${GREEN}All tests passed!${NC}"
        return 0
    else
        echo -e "${RED}Some tests failed.${NC}"
        return 1
    fi
}

# Main execution
main() {
    log_section "Phase 3 Week 1 Integration Tests"
    echo "Backend Directory: $BACKEND_DIR"
    echo "Base URL: $BASE_URL"
    echo "Test Timeout: ${TEST_TIMEOUT}s"
    
    # Run verification tests
    verify_build
    run_unit_tests
    check_code_quality
    
    # Run integration tests if backend is available
    if check_backend; then
        test_health
        test_vm_endpoints
        test_backward_compatibility
    else
        log_info "Skipping integration tests (backend not running)"
    fi
    
    print_summary
    exit $?
}

main "$@"
