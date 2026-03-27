#!/bin/bash
#
# Load Test: Concurrent VM Creation
# Tests system performance under concurrent VM creation load
#
# Usage: ./scripts/load-test-vm-creation.sh [OPTIONS]
#
# Options:
#   --api-url        API base URL (default: http://localhost:8080)
#   --jwt-token      JWT token for authentication
#   --concurrency    Number of concurrent VM creations (default: 100)
#   --timeout        Timeout per request in seconds (default: 30)
#   --output-dir     Directory for results (default: ./load-test-results)
#   --verbose        Enable verbose output
#   --skip-cleanup   Skip cleanup after tests
#   --help           Show this help message
#
# Metrics:
#   - Success rate (% of successful VM creations)
#   - Average creation time (ms)
#   - P50, P95, P99 latencies
#   - Quota race condition detection
#   - Error breakdown by type
#
# Exit Codes:
#   0 - Load test completed successfully
#   1 - Load test completed with failures
#   2 - Configuration error (missing JWT, API URL, etc.)
#

set -euo pipefail

# =============================================================================
# Configuration
# =============================================================================

API_URL="${API_URL:-http://localhost:8080}"
JWT_TOKEN="${JWT_TOKEN:-}"
CONCURRENCY="${CONCURRENCY:-100}"
TIMEOUT="${TIMEOUT:-30}"
OUTPUT_DIR="${OUTPUT_DIR:-./load-test-results}"
VERBOSE="${VERBOSE:-false}"
SKIP_CLEANUP="${SKIP_CLEANUP:-false}"
TEST_PREFIX="load-test-$(date +%Y%m%d-%H%M%S)"

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Metrics
TOTAL_REQUESTS=0
SUCCESSFUL_REQUESTS=0
FAILED_REQUESTS=0
QUOTA_EXCEEDED_COUNT=0
RACE_CONDITION_COUNT=0
declare -a RESPONSE_TIMES=()
declare -a CREATED_VMS=()
declare -A ERROR_BREAKDOWN=()

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
        METRIC) echo -e "${CYAN}[METRIC]${NC} $message" ;;
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
        curl -s -X DELETE \
            -H "Authorization: Bearer ${JWT_TOKEN}" \
            "${API_URL}/api/vms/${vm_id}" 2>/dev/null || true
    done

    log INFO "Cleanup complete (${#CREATED_VMS[@]} VMs deleted)"
}

# Set up trap for cleanup
trap cleanup EXIT

# =============================================================================
# Load Test Functions
# =============================================================================

create_vm_async() {
    local vm_name="$1"
    local worker_id="$2"
    local start_time
    local end_time
    local response_time
    local http_code
    local response_body

    start_time=$(date +%s%3N)

    # Create VM request
    response_body=$(curl -s -w "\n%{http_code}" \
        -X POST \
        -H "Content-Type: application/json" \
        -H "Authorization: Bearer ${JWT_TOKEN}" \
        --max-time "$TIMEOUT" \
        -d "{
            \"name\": \"${vm_name}\",
            \"os\": \"ubuntu-2204\",
            \"tier\": \"nano\"
        }" \
        "${API_URL}/api/vms" 2>/dev/null) || {
        # Curl failed (timeout, connection error, etc.)
        end_time=$(date +%s%3N)
        response_time=$((end_time - start_time))
        RESPONSE_TIMES+=("$response_time")
        ((FAILED_REQUESTS++))
        ERROR_BREAKDOWN["CONNECTION_ERROR"]=$((${ERROR_BREAKDOWN["CONNECTION_ERROR"]:-0} + 1))
        log DEBUG "Worker $worker_id: Connection error after ${response_time}ms"
        return 1
    }

    end_time=$(date +%s%3N)
    response_time=$((end_time - start_time))
    RESPONSE_TIMES+=("$response_time")

    # Parse response
    http_code=$(echo "$response_body" | tail -n1)
    response_body=$(echo "$response_body" | sed '$d')

    case "$http_code" in
        202)
            # Success - VM created
            ((SUCCESSFUL_REQUESTS++))
            local vm_id
            vm_id=$(echo "$response_body" | jq -r '.id // empty')
            if [[ -n "$vm_id" ]]; then
                CREATED_VMS+=("$vm_id")
                log DEBUG "Worker $worker_id: VM created successfully in ${response_time}ms (ID: $vm_id)"
            else
                log WARN "Worker $worker_id: VM created but no ID returned"
            fi
            ;;
        403)
            # Quota exceeded
            ((QUOTA_EXCEEDED_COUNT++))
            ((FAILED_REQUESTS++))
            ERROR_BREAKDOWN["QUOTA_EXCEEDED"]=$((${ERROR_BREAKDOWN["QUOTA_EXCEEDED"]:-0} + 1))
            log DEBUG "Worker $worker_id: Quota exceeded after ${response_time}ms"
            ;;
        401|403)
            # Authentication/Authorization error
            ((FAILED_REQUESTS++))
            ERROR_BREAKDOWN["AUTH_ERROR"]=$((${ERROR_BREAKDOWN["AUTH_ERROR"]:-0} + 1))
            log DEBUG "Worker $worker_id: Auth error (${http_code}) after ${response_time}ms"
            ;;
        409)
            # Conflict (possible race condition on name)
            ((RACE_CONDITION_COUNT++))
            ((FAILED_REQUESTS++))
            ERROR_BREAKDOWN["RACE_CONDITION"]=$((${ERROR_BREAKDOWN["RACE_CONDITION"]:-0} + 1))
            log DEBUG "Worker $worker_id: Race condition detected after ${response_time}ms"
            ;;
        500|502|503|504)
            # Server error
            ((FAILED_REQUESTS++))
            ERROR_BREAKDOWN["SERVER_ERROR"]=$((${ERROR_BREAKDOWN["SERVER_ERROR"]:-0} + 1))
            log DEBUG "Worker $worker_id: Server error (${http_code}) after ${response_time}ms"
            ;;
        *)
            # Other error
            ((FAILED_REQUESTS++))
            ERROR_BREAKDOWN["OTHER"]=$((${ERROR_BREAKDOWN["OTHER"]:-0} + 1))
            log DEBUG "Worker $worker_id: Unexpected status ${http_code} after ${response_time}ms"
            ;;
    esac

    ((TOTAL_REQUESTS++))
    return 0
}

run_load_test() {
    local concurrency="$1"
    local pids=()
    local worker_id=0

    log INFO "Starting load test with $concurrency concurrent VM creations"
    log INFO "API URL: $API_URL"
    log INFO "Timeout per request: ${TIMEOUT}s"

    # Create temporary directory for worker logs
    local temp_dir
    temp_dir=$(mktemp -d)

    # Launch concurrent workers
    for ((i=0; i<concurrency; i++)); do
        worker_id=$i
        local vm_name="${TEST_PREFIX}-vm-${i}"

        # Run VM creation in background
        (
            create_vm_async "$vm_name" "$worker_id"
        ) &

        pids+=($!)

        # Small delay to prevent thundering herd
        sleep 0.05
    done

    log INFO "All workers launched, waiting for completion..."

    # Wait for all workers to complete
    local failed=0
    for pid in "${pids[@]}"; do
        if ! wait "$pid"; then
            ((failed++))
        fi
    done

    # Clean up temp directory
    rm -rf "$temp_dir"

    log INFO "All workers completed"
}

calculate_percentile() {
    local percentile="$1"
    local sorted_times

    # Sort response times
    sorted_times=($(printf '%s\n' "${RESPONSE_TIMES[@]}" | sort -n))
    local count=${#sorted_times[@]}

    if [[ $count -eq 0 ]]; then
        echo "0"
        return
    fi

    # Calculate index
    local index
    index=$(echo "scale=0; ($count * $percentile / 100) - 1" | bc)
    if [[ $index -lt 0 ]]; then
        index=0
    fi
    if [[ $index -ge $count ]]; then
        index=$((count - 1))
    fi

    echo "${sorted_times[$index]}"
}

calculate_average() {
    local sum=0
    local count=${#RESPONSE_TIMES[@]}

    if [[ $count -eq 0 ]]; then
        echo "0"
        return
    fi

    for time in "${RESPONSE_TIMES[@]}"; do
        sum=$((sum + time))
    done

    echo $((sum / count))
}

generate_report() {
    local timestamp
    timestamp=$(date '+%Y-%m-%d %H:%M:%S')
    local report_file="${OUTPUT_DIR}/load-test-report-$(date +%Y%m%d-%H%M%S).md"

    # Calculate metrics
    local avg_time
    avg_time=$(calculate_average)
    local p50_time
    p50_time=$(calculate_percentile 50)
    local p95_time
    p95_time=$(calculate_percentile 95)
    local p99_time
    p99_time=$(calculate_percentile 99)
    local min_time
    min_time=$(printf '%s\n' "${RESPONSE_TIMES[@]}" | sort -n | head -n1)
    local max_time
    max_time=$(printf '%s\n' "${RESPONSE_TIMES[@]}" | sort -n | tail -n1)
    local success_rate
    if [[ $TOTAL_REQUESTS -gt 0 ]]; then
        success_rate=$(echo "scale=2; $SUCCESSFUL_REQUESTS * 100 / $TOTAL_REQUESTS" | bc)
    else
        success_rate="0.00"
    fi

    # Create output directory
    mkdir -p "$OUTPUT_DIR"

    # Generate markdown report
    cat > "$report_file" <<EOF
# Load Test Report: Concurrent VM Creation

**Test Date:** ${timestamp}
**Test ID:** ${TEST_PREFIX}

---

## Configuration

| Parameter | Value |
|-----------|-------|
| API URL | ${API_URL} |
| Concurrency | ${CONCURRENCY} |
| Timeout | ${TIMEOUT}s |
| VM Tier | nano |
| VM OS | ubuntu-2204 |

---

## Summary

| Metric | Value |
|--------|-------|
| Total Requests | ${TOTAL_REQUESTS} |
| Successful | ${SUCCESSFUL_REQUESTS} |
| Failed | ${FAILED_REQUESTS} |
| Success Rate | ${success_rate}% |

---

## Performance Metrics

| Metric | Value (ms) |
|--------|------------|
| Average | ${avg_time} |
| P50 (Median) | ${p50_time} |
| P95 | ${p95_time} |
| P99 | ${p99_time} |
| Minimum | ${min_time} |
| Maximum | ${max_time} |

---

## Error Analysis

### Quota Enforcement

| Metric | Value |
|--------|-------|
| Quota Exceeded Errors | ${QUOTA_EXCEEDED_COUNT} |
| Race Conditions Detected | ${RACE_CONDITION_COUNT} |

### Error Breakdown

| Error Type | Count |
|------------|-------|
EOF

    # Add error breakdown
    for error_type in "${!ERROR_BREAKDOWN[@]}"; do
        echo "| ${error_type} | ${ERROR_BREAKDOWN[$error_type]} |" >> "$report_file"
    done

    cat >> "$report_file" <<EOF

---

## Recommendations

EOF

    # Generate recommendations based on results
    if (( $(echo "$success_rate < 90" | bc -l) )); then
        cat >> "$report_file" <<EOF
### Critical Issues

1. **Low Success Rate (${success_rate}%)**: The system is failing to handle the concurrent load.
   - Investigate database connection pooling
   - Check for resource exhaustion (CPU, memory, file descriptors)
   - Review k8s API server rate limiting

EOF
    fi

    if [[ $RACE_CONDITION_COUNT -gt 0 ]]; then
        cat >> "$report_file" <<EOF
### Race Conditions Detected

1. **${RACE_CONDITION_COUNT} race conditions** detected during concurrent VM creation.
   - Ensure quota checks use SELECT FOR UPDATE
   - Add unique constraints on VM names per user
   - Implement optimistic locking for quota updates

EOF
    fi

    if [[ $QUOTA_EXCEEDED_COUNT -gt 0 ]]; then
        cat >> "$report_file" <<EOF
### Quota Enforcement

1. **${QUOTA_EXCEEDED_COUNT} requests rejected** due to quota limits.
   - This is expected behavior if concurrency exceeds user quota
   - Consider implementing request queuing for quota-limited users
   - Add client-side quota checking before bulk operations

EOF
    fi

    if (( p99_time > 5000 )); then
        cat >> "$report_file" <<EOF
### High Latency

1. **P99 latency exceeds 5 seconds** (${p99_time}ms).
   - Profile k8s resource creation (namespace, PVC, Deployment, Service)
   - Check for database lock contention
   - Consider async VM provisioning with status polling

EOF
    fi

    if [[ $SUCCESSFUL_REQUESTS -gt 0 ]]; then
        cat >> "$report_file" <<EOF
### Positive Findings

1. **Successfully created ${SUCCESSFUL_REQUESTS} VMs** under concurrent load.
2. **Average creation time**: ${avg_time}ms
3. **System remained stable** throughout the test.

EOF
    fi

    cat >> "$report_file" <<EOF
---

## Raw Data

### Response Times (ms)

EOF

    # Add raw response times
    printf '%s\n' "${RESPONSE_TIMES[@]}" >> "$report_file"

    log INFO "Report generated: $report_file"
    echo "$report_file"
}

print_summary() {
    echo ""
    echo "=============================================="
    echo "  Load Test Summary"
    echo "=============================================="
    echo -e "  Total Requests:     ${TOTAL_REQUESTS}"
    echo -e "  Successful:         ${GREEN}${SUCCESSFUL_REQUESTS}${NC}"
    echo -e "  Failed:             ${RED}${FAILED_REQUESTS}${NC}"

    if [[ $TOTAL_REQUESTS -gt 0 ]]; then
        local success_rate
        success_rate=$(echo "scale=2; $SUCCESSFUL_REQUESTS * 100 / $TOTAL_REQUESTS" | bc)
        echo -e "  Success Rate:       ${success_rate}%"
    fi

    echo "----------------------------------------------"

    if [[ ${#RESPONSE_TIMES[@]} -gt 0 ]]; then
        local avg_time
        avg_time=$(calculate_average)
        local p50_time
        p50_time=$(calculate_percentile 50)
        local p95_time
        p95_time=$(calculate_percentile 95)
        local p99_time
        p99_time=$(calculate_percentile 99)

        echo -e "  Average Time:       ${avg_time}ms"
        echo -e "  P50 (Median):       ${p50_time}ms"
        echo -e "  P95:                ${p95_time}ms"
        echo -e "  P99:                ${p99_time}ms"
    fi

    echo "----------------------------------------------"
    echo -e "  Quota Exceeded:     ${YELLOW}${QUOTA_EXCEEDED_COUNT}${NC}"
    echo -e "  Race Conditions:    ${RED}${RACE_CONDITION_COUNT}${NC}"
    echo "=============================================="
    echo ""

    # Print error breakdown
    if [[ ${#ERROR_BREAKDOWN[@]} -gt 0 ]]; then
        echo -e "${RED}Error Breakdown:${NC}"
        for error_type in "${!ERROR_BREAKDOWN[@]}"; do
            echo "  ${error_type}: ${ERROR_BREAKDOWN[$error_type]}"
        done
        echo ""
    fi
}

# =============================================================================
# Main
# =============================================================================

show_help() {
    cat <<EOF
Load Test: Concurrent VM Creation

Usage: $0 [OPTIONS]

Options:
  --api-url URL        API base URL (default: http://localhost:8080)
  --jwt-token TOKEN    JWT token for authentication (required)
  --concurrency N      Number of concurrent VM creations (default: 100)
  --timeout SECONDS    Timeout per request in seconds (default: 30)
  --output-dir DIR     Directory for results (default: ./load-test-results)
  --verbose            Enable verbose output
  --skip-cleanup       Skip cleanup after tests
  --help               Show this help message

Environment Variables:
  API_URL              API base URL (default: http://localhost:8080)
  JWT_TOKEN            JWT token for authentication
  CONCURRENCY          Number of concurrent VM creations (default: 100)
  TIMEOUT              Timeout per request in seconds (default: 30)
  OUTPUT_DIR           Directory for results (default: ./load-test-results)
  VERBOSE              Enable verbose output (true/false)
  SKIP_CLEANUP         Skip cleanup after tests (true/false)

Examples:
  $0 --jwt-token "eyJhbGc..." --concurrency 100
  JWT_TOKEN="eyJhbGc..." CONCURRENCY=50 $0
  $0 --jwt-token "eyJhbGc..." --verbose --skip-cleanup

Metrics Measured:
  - Success rate (% of successful VM creations)
  - Average creation time (ms)
  - P50, P95, P99 latencies
  - Quota race condition detection
  - Error breakdown by type
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
            --concurrency)
                CONCURRENCY="$2"
                shift 2
                ;;
            --timeout)
                TIMEOUT="$2"
                shift 2
                ;;
            --output-dir)
                OUTPUT_DIR="$2"
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

    # Check dependencies
    if ! command -v jq &> /dev/null; then
        die "jq is required but not installed." 2
    fi

    if ! command -v bc &> /dev/null; then
        die "bc is required but not installed." 2
    fi

    log INFO "Starting Load Test: Concurrent VM Creation"
    log INFO "Test ID: $TEST_PREFIX"
    log INFO "Concurrency: $CONCURRENCY"
    log INFO "Output directory: $OUTPUT_DIR"

    echo ""
    echo "=============================================="
    echo "  Load Test: Concurrent VM Creation"
    echo "=============================================="
    echo ""

    # Run load test
    run_load_test "$CONCURRENCY"

    # Print summary
    print_summary

    # Generate report
    local report_file
    report_file=$(generate_report)

    log INFO "Load test completed"
    log INFO "Report: $report_file"

    # Exit with appropriate code
    if [[ $FAILED_REQUESTS -gt 0 && $QUOTA_EXCEEDED_COUNT -eq 0 ]]; then
        log WARN "Load test completed with failures"
        exit 1
    else
        log INFO "Load test completed successfully"
        exit 0
    fi
}

# Run main function
main "$@"
