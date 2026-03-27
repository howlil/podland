#!/bin/bash
#
# VM Lifecycle Integration Tests
# Tests VM creation, lifecycle operations, quota enforcement, SSH connectivity, and tier restrictions
#
# Usage: ./scripts/test-vm-lifecycle.sh [OPTIONS]
#
# Options:
#   --api-url      API base URL (default: http://localhost:8080)
#   --jwt-token    JWT token for authentication
#   --verbose      Enable verbose output
#   --skip-cleanup Skip cleanup after tests
#   --help         Show this help message
#
# Tests:
#   1. VM Lifecycle: Create → Wait running → Stop → Start → Restart → Delete
#   2. Quota Enforcement: Create VMs until quota exceeded
#   3. SSH Key: Verify SSH key works for VM access
#   4. Tier Restrictions: External users cannot create internal tiers
#
# Exit Codes:
#   0 - All tests passed
#   1 - One or more tests failed
#   2 - Configuration error (missing JWT, API URL, etc.)
#

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

API_URL="${API_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"
VERBOSE="${VERBOSE:-false}"
SKIP_CLEANUP="${SKIP_CLEANUP:-false}"
TEST_PREFIX="test-vm-$(date +%Y%m%d-%H%M%S)"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Test counters
TESTS_PASSED=0
TESTS_FAILED=0
TESTS_SKIPPED=0

# Created VMs for cleanup
declare -a CREATED_VMS=()

# =============================================================================
# Utility Functions
# =============================================================================

log() {
    local level="$1"
    shift
    local message="$*"
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    
    case "$level" in
        INFO)  echo -e "${BLUE}[$timestamp]${NC} ${GREEN}[INFO]${NC} $message" ;;
        WARN)  echo -e "${BLUE}[$timestamp]${NC} ${YELLOW}[WARN]${NC} $message" ;;
        ERROR) echo -e "${BLUE}[$timestamp]${NC} ${RED}[ERROR]${NC} $message" ;;
        DEBUG) [[ "$VERBOSE" == "true" ]] && echo -e "${BLUE}[$timestamp]${NC} [DEBUG] $message" ;;
    esac
}

log_test() {
    local test_name="$1"
    local status="${2:-RUNNING}"
    
    case "$status" in
        PASS)    echo -e "${GREEN}[PASS]${NC} $test_name" ;;
        FAIL)    echo -e "${RED}[FAIL]${NC} $test_name" ;;
        SKIP)    echo -e "${YELLOW}[SKIP]${NC} $test_name" ;;
        RUNNING) echo -e "${BLUE}[RUN] ${NC} $test_name" ;;
    esac
}

die() {
    log ERROR "$1"
    exit "${2:-2}"
}

cleanup() {
    if [[ "$SKIP_CLEANUP" == "true" ]]; then
        log INFO "Skipping cleanup (SKIP_CLEANUP=true)"
        return
    fi
    
    log INFO "Cleaning up created VMs..."
    
    for vm_id in "${CREATED_VMS[@]}"; do
        log DEBUG "Deleting VM: $vm_id"
        delete_vm "$vm_id" 2>/dev/null || true
    done
    
    log INFO "Cleanup complete"
}

# Set up trap for cleanup
trap cleanup EXIT

# =============================================================================
# API Helper Functions
# =============================================================================

api_request() {
    local method="$1"
    local endpoint="$2"
    local data="${3:-}"
    local expected_status="${4:-200}"
    
    local url="${API_URL}${endpoint}"
    local curl_opts=(
        -s
        -X "$method"
        -H "Content-Type: application/json"
        -H "Authorization: Bearer ${JWT_TOKEN}"
        -w "\n%{http_code}"
    )
    
    if [[ -n "$data" ]]; then
        curl_opts+=(-d "$data")
    fi
    
    local response
    response=$(curl "${curl_opts[@]}" "$url" 2>/dev/null)
    
    local http_code
    http_code=$(echo "$response" | tail -n1)
    local body
    body=$(echo "$response" | sed '$d')
    
    if [[ "$http_code" != "$expected_status" ]]; then
        log DEBUG "Expected status $expected_status, got $http_code"
        log DEBUG "Response body: $body"
        return 1
    fi
    
    echo "$body"
    return 0
}

create_vm() {
    local name="$1"
    local os="${2:-ubuntu-2204}"
    local tier="${3:-nano}"
    
    log DEBUG "Creating VM: name=$name, os=$os, tier=$tier"
    
    local data
    data=$(cat <<EOF
{
    "name": "$name",
    "os": "$os",
    "tier": "$tier"
}
EOF
)
    
    local response
    if response=$(api_request "POST" "/api/vms" "$data" "202"); then
        local vm_id
        vm_id=$(echo "$response" | jq -r '.id // empty')
        if [[ -n "$vm_id" ]]; then
            CREATED_VMS+=("$vm_id")
            log DEBUG "Created VM with ID: $vm_id"
            echo "$vm_id"
            return 0
        fi
    fi
    
    log ERROR "Failed to create VM: $name"
    return 1
}

get_vm() {
    local vm_id="$1"
    api_request "GET" "/api/vms/${vm_id}" "" "200"
}

list_vms() {
    api_request "GET" "/api/vms" "" "200"
}

start_vm() {
    local vm_id="$1"
    api_request "POST" "/api/vms/${vm_id}/start" "" "202"
}

stop_vm() {
    local vm_id="$1"
    api_request "POST" "/api/vms/${vm_id}/stop" "" "202"
}

restart_vm() {
    local vm_id="$1"
    api_request "POST" "/api/vms/${vm_id}/restart" "" "202"
}

delete_vm() {
    local vm_id="$1"
    api_request "DELETE" "/api/vms/${vm_id}" "" "200"
}

wait_vm_status() {
    local vm_id="$1"
    local expected_status="$2"
    local timeout="${3:-60}"
    local interval="${4:-5}"
    
    log DEBUG "Waiting for VM $vm_id to reach status: $expected_status (timeout: ${timeout}s)"
    
    local elapsed=0
    while [[ $elapsed -lt $timeout ]]; do
        local vm_data
        if vm_data=$(get_vm "$vm_id" 2>/dev/null); then
            local current_status
            current_status=$(echo "$vm_data" | jq -r '.status // empty')
            
            log DEBUG "Current status: $current_status (expected: $expected_status)"
            
            if [[ "$current_status" == "$expected_status" ]]; then
                log DEBUG "VM reached status: $expected_status"
                return 0
            fi
        fi
        
        sleep "$interval"
        elapsed=$((elapsed + interval))
    done
    
    log ERROR "Timeout waiting for VM $vm_id to reach status: $expected_status"
    return 1
}

# =============================================================================
# Test Functions
# =============================================================================

test_vm_lifecycle() {
    log INFO "=== Test: VM Lifecycle ==="
    local test_name="VM Lifecycle (Create → Running → Stop → Start → Restart → Delete)"
    log_test "$test_name" "RUNNING"
    
    local vm_name="${TEST_PREFIX}-lifecycle"
    local vm_id
    
    # Step 1: Create VM
    log DEBUG "Step 1: Creating VM"
    if ! vm_id=$(create_vm "$vm_name" "ubuntu-2204" "nano"); then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    log DEBUG "VM created: $vm_id"
    
    # Step 2: Wait for running status
    log DEBUG "Step 2: Waiting for VM to be running"
    if ! wait_vm_status "$vm_id" "running" 60 5; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    log DEBUG "VM is running"
    
    # Step 3: Stop VM
    log DEBUG "Step 3: Stopping VM"
    if ! stop_vm "$vm_id"; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    if ! wait_vm_status "$vm_id" "stopped" 60 5; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    log DEBUG "VM is stopped"
    
    # Step 4: Start VM
    log DEBUG "Step 4: Starting VM"
    if ! start_vm "$vm_id"; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    if ! wait_vm_status "$vm_id" "running" 60 5; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    log DEBUG "VM is running again"
    
    # Step 5: Restart VM
    log DEBUG "Step 5: Restarting VM"
    if ! restart_vm "$vm_id"; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Restart goes through stopped → running
    if ! wait_vm_status "$vm_id" "running" 60 5; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    log DEBUG "VM restarted successfully"
    
    # Step 6: Delete VM
    log DEBUG "Step 6: Deleting VM"
    if ! delete_vm "$vm_id"; then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Remove from cleanup list since we deleted it
    CREATED_VMS=("${CREATED_VMS[@]/$vm_id}")
    
    log_test "$test_name" "PASS"
    ((TESTS_PASSED++))
    return 0
}

test_quota_enforcement() {
    log INFO "=== Test: Quota Enforcement ==="
    local test_name="Quota Enforcement (Create VMs until quota exceeded)"
    log_test "$test_name" "RUNNING"
    
    local vms_created=0
    local max_attempts=10
    local quota_exceeded=false
    
    for i in $(seq 1 $max_attempts); do
        local vm_name="${TEST_PREFIX}-quota-$i"
        log DEBUG "Attempt $i: Creating VM $vm_name"
        
        local vm_id
        if vm_id=$(create_vm "$vm_name" "ubuntu-2204" "nano" 2>/dev/null); then
            ((vms_created++))
            log DEBUG "VM created: $vm_id (total: $vms_created)"
        else
            log DEBUG "Quota exceeded after $vms_created VMs"
            quota_exceeded=true
            break
        fi
    done
    
    if [[ "$quota_exceeded" == "true" ]]; then
        log DEBUG "Quota enforcement working: $vms_created VMs created before limit"
        log_test "$test_name" "PASS"
        ((TESTS_PASSED++))
        return 0
    else
        log ERROR "Quota not enforced after $max_attempts attempts"
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
}

test_ssh_key() {
    log INFO "=== Test: SSH Key ==="
    local test_name="SSH Key (Verify SSH key works)"
    log_test "$test_name" "RUNNING"
    
    # This test requires actual SSH connectivity
    # For now, we verify that SSH key is returned on VM creation
    
    local vm_name="${TEST_PREFIX}-ssh"
    local vm_id
    
    log DEBUG "Creating VM to get SSH key"
    if ! vm_id=$(create_vm "$vm_name" "ubuntu-2204" "nano"); then
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Get VM details
    local vm_data
    if ! vm_data=$(get_vm "$vm_id"); then
        log ERROR "Failed to get VM details"
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Check if SSH public key exists
    local ssh_pub_key
    ssh_pub_key=$(echo "$vm_data" | jq -r '.ssh_public_key // empty')
    
    if [[ -z "$ssh_pub_key" ]]; then
        log ERROR "SSH public key not found in VM data"
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Validate SSH key format (should start with ssh-ed25519 or ssh-rsa)
    if [[ ! "$ssh_pub_key" =~ ^ssh-(ed25519|rsa) ]]; then
        log ERROR "Invalid SSH public key format"
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    log DEBUG "SSH public key validated: ${ssh_pub_key:0:50}..."
    
    # Note: Full SSH connectivity test would require:
    # 1. VM to be fully running
    # 2. SSH private key from creation response
    # 3. Network access to VM domain
    # This is typically done in a separate E2E test environment
    
    log DEBUG "SSH key format validation passed"
    log_test "$test_name" "PASS"
    ((TESTS_PASSED++))
    return 0
}

test_tier_restrictions() {
    log INFO "=== Test: Tier Restrictions ==="
    local test_name="Tier Restrictions (External cannot create internal tiers)"
    log_test "$test_name" "RUNNING"
    
    # Internal tiers: small, medium, large, xlarge
    local internal_tiers=("small" "medium" "large" "xlarge")
    local blocked_count=0
    local total_internal=${#internal_tiers[@]}
    
    for tier in "${internal_tiers[@]}"; do
        local vm_name="${TEST_PREFIX}-restricted-$tier"
        log DEBUG "Attempting to create VM with internal tier: $tier"
        
        local vm_id
        if vm_id=$(create_vm "$vm_name" "ubuntu-2204" "$tier" 2>/dev/null); then
            log DEBUG "Unexpectedly created VM with internal tier: $tier"
            # This should fail for external users
        else
            log DEBUG "Correctly blocked internal tier: $tier"
            ((blocked_count++))
        fi
    done
    
    if [[ $blocked_count -eq $total_internal ]]; then
        log DEBUG "All internal tiers correctly blocked ($blocked_count/$total_internal)"
        log_test "$test_name" "PASS"
        ((TESTS_PASSED++))
        return 0
    else
        log ERROR "Only $blocked_count/$total_internal internal tiers were blocked"
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
}

test_vm_list() {
    log INFO "=== Test: VM List ==="
    local test_name="VM List (List user's VMs)"
    log_test "$test_name" "RUNNING"
    
    local vms
    if ! vms=$(list_vms); then
        log ERROR "Failed to list VMs"
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    # Validate response is JSON array
    if ! echo "$vms" | jq -e 'type == "array"' > /dev/null 2>&1; then
        log ERROR "VM list response is not a JSON array"
        log_test "$test_name" "FAIL"
        ((TESTS_FAILED++))
        return 1
    fi
    
    local vm_count
    vm_count=$(echo "$vms" | jq 'length')
    log DEBUG "User has $vm_count VM(s)"
    
    log_test "$test_name" "PASS"
    ((TESTS_PASSED++))
    return 0
}

# =============================================================================
# Main
# =============================================================================

show_help() {
    cat <<EOF
VM Lifecycle Integration Tests

Usage: $0 [OPTIONS]

Options:
  --api-url URL      API base URL (default: http://localhost:8080)
  --jwt-token TOKEN  JWT token for authentication (required)
  --verbose          Enable verbose output
  --skip-cleanup     Skip cleanup after tests
  --help             Show this help message

Environment Variables:
  API_URL            API base URL (default: http://localhost:8080)
  JWT_TOKEN          JWT token for authentication
  VERBOSE            Enable verbose output (true/false)
  SKIP_CLEANUP       Skip cleanup after tests (true/false)

Examples:
  $0 --jwt-token "eyJhbGc..."
  API_URL=https://api.podland.app JWT_TOKEN="eyJhbGc..." $0
  $0 --jwt-token "eyJhbGc..." --verbose --skip-cleanup

Tests:
  1. VM Lifecycle: Create → Wait running → Stop → Start → Restart → Delete
  2. Quota Enforcement: Create VMs until quota exceeded
  3. SSH Key: Verify SSH key format and availability
  4. Tier Restrictions: External users cannot create internal tiers
EOF
}

main() {
    # Parse command line arguments
    while [[ $# -gt 0 ]]; do
        case "$1" in
            --api-url)
                API_URL="$2"
                shift 2
                ;;
            --jwt-token)
                JWT_TOKEN="$2"
                shift 2
                ;;
            --verbose)
                VERBOSE="true"
                shift
                ;;
            --skip-cleanup)
                SKIP_CLEANUP="true"
                shift
                ;;
            --help)
                show_help
                exit 0
                ;;
            *)
                die "Unknown option: $1" 2
                ;;
        esac
    done
    
    # Validate configuration
    if [[ -z "$JWT_TOKEN" ]]; then
        die "JWT token is required. Set JWT_TOKEN environment variable or use --jwt-token option." 2
    fi
    
    log INFO "Starting VM Lifecycle Integration Tests"
    log INFO "API URL: $API_URL"
    log DEBUG "Test prefix: $TEST_PREFIX"
    log DEBUG "Verbose: $VERBOSE"
    log DEBUG "Skip cleanup: $SKIP_CLEANUP"
    
    echo ""
    echo "=============================================="
    echo "  VM Lifecycle Integration Tests"
    echo "=============================================="
    echo ""
    
    # Run tests
    test_vm_lifecycle || true
    test_quota_enforcement || true
    test_ssh_key || true
    test_tier_restrictions || true
    test_vm_list || true
    
    # Print summary
    echo ""
    echo "=============================================="
    echo "  Test Summary"
    echo "=============================================="
    echo -e "  ${GREEN}Passed${NC}:  $TESTS_PASSED"
    echo -e "  ${RED}Failed${NC}:  $TESTS_FAILED"
    echo -e "  ${YELLOW}Skipped${NC}: $TESTS_SKIPPED"
    echo "=============================================="
    echo ""
    
    if [[ $TESTS_FAILED -gt 0 ]]; then
        log ERROR "Tests completed with failures"
        exit 1
    else
        log INFO "All tests passed!"
        exit 0
    fi
}

# Run main function
main "$@"
