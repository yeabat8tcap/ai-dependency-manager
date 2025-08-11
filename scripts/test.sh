#!/bin/bash

# AI Dependency Manager Test Runner
# This script runs comprehensive tests for the AI Dependency Manager

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
COVERAGE_THRESHOLD=80
TEST_TIMEOUT=10m
PARALLEL_JOBS=4

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check Go version
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"
    
    # Check required tools
    local tools=("npm" "python3" "pip3")
    for tool in "${tools[@]}"; do
        if command -v "$tool" &> /dev/null; then
            log_info "$tool is available"
        else
            log_warning "$tool is not available - some tests may be skipped"
        fi
    done
}

setup_test_environment() {
    log_info "Setting up test environment..."
    
    # Create test directories
    mkdir -p test/tmp
    mkdir -p test/reports
    mkdir -p test/coverage
    
    # Set environment variables
    export AI_DEP_MANAGER_DATA_DIR="$(pwd)/test/tmp"
    export AI_DEP_MANAGER_LOG_LEVEL="error"
    export AI_DEP_MANAGER_TEST_MODE="true"
    
    log_success "Test environment ready"
}

cleanup_test_environment() {
    log_info "Cleaning up test environment..."
    rm -rf test/tmp/*
    rm -rf test/reports/*
}

run_unit_tests() {
    log_info "Running unit tests..."
    
    go test -v -short -timeout="$TEST_TIMEOUT" -parallel="$PARALLEL_JOBS" \
        -coverprofile=test/coverage/unit.out \
        -covermode=atomic \
        ./internal/... | tee test/reports/unit.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Unit tests passed"
    else
        log_error "Unit tests failed"
        return 1
    fi
}

run_integration_tests() {
    log_info "Running integration tests..."
    
    go test -v -run Integration -timeout="$TEST_TIMEOUT" -parallel="$PARALLEL_JOBS" \
        -coverprofile=test/coverage/integration.out \
        -covermode=atomic \
        ./test/integration/... | tee test/reports/integration.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Integration tests passed"
    else
        log_error "Integration tests failed"
        return 1
    fi
}

run_e2e_tests() {
    log_info "Running end-to-end tests..."
    
    # Build binary for e2e tests
    make build
    
    go test -v -timeout="$TEST_TIMEOUT" -parallel=1 \
        ./test/e2e/... | tee test/reports/e2e.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "End-to-end tests passed"
    else
        log_error "End-to-end tests failed"
        return 1
    fi
}

run_security_tests() {
    log_info "Running security tests..."
    
    go test -v -run Security -timeout="$TEST_TIMEOUT" \
        ./... | tee test/reports/security.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Security tests passed"
    else
        log_error "Security tests failed"
        return 1
    fi
}

run_race_tests() {
    log_info "Running race detection tests..."
    
    go test -v -race -timeout="$TEST_TIMEOUT" \
        ./internal/... | tee test/reports/race.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Race detection tests passed"
    else
        log_error "Race detection tests failed"
        return 1
    fi
}

run_benchmark_tests() {
    log_info "Running benchmark tests..."
    
    go test -v -bench=. -benchmem -timeout="$TEST_TIMEOUT" \
        ./... | tee test/reports/benchmark.log
    
    if [ ${PIPESTATUS[0]} -eq 0 ]; then
        log_success "Benchmark tests completed"
    else
        log_warning "Some benchmark tests failed"
    fi
}

generate_coverage_report() {
    log_info "Generating coverage report..."
    
    # Merge coverage files
    if [ -f test/coverage/unit.out ] && [ -f test/coverage/integration.out ]; then
        echo "mode: atomic" > test/coverage/merged.out
        tail -n +2 test/coverage/unit.out >> test/coverage/merged.out
        tail -n +2 test/coverage/integration.out >> test/coverage/merged.out
    elif [ -f test/coverage/unit.out ]; then
        cp test/coverage/unit.out test/coverage/merged.out
    elif [ -f test/coverage/integration.out ]; then
        cp test/coverage/integration.out test/coverage/merged.out
    else
        log_warning "No coverage files found"
        return 0
    fi
    
    # Generate HTML report
    go tool cover -html=test/coverage/merged.out -o test/coverage/coverage.html
    
    # Calculate coverage percentage
    COVERAGE=$(go tool cover -func=test/coverage/merged.out | grep total | awk '{print $3}' | sed 's/%//')
    
    log_info "Total coverage: ${COVERAGE}%"
    
    if (( $(echo "$COVERAGE >= $COVERAGE_THRESHOLD" | bc -l) )); then
        log_success "Coverage meets threshold (${COVERAGE_THRESHOLD}%)"
    else
        log_warning "Coverage below threshold: ${COVERAGE}% < ${COVERAGE_THRESHOLD}%"
    fi
}

run_linting() {
    log_info "Running linting..."
    
    if command -v golangci-lint &> /dev/null; then
        golangci-lint run --timeout="$TEST_TIMEOUT" | tee test/reports/lint.log
        
        if [ ${PIPESTATUS[0]} -eq 0 ]; then
            log_success "Linting passed"
        else
            log_error "Linting failed"
            return 1
        fi
    else
        log_warning "golangci-lint not installed, skipping linting"
    fi
}

generate_test_report() {
    log_info "Generating test report..."
    
    cat > test/reports/summary.md << EOF
# Test Report

Generated: $(date)

## Test Results

EOF
    
    # Add results for each test type
    local test_types=("unit" "integration" "e2e" "security" "race")
    
    for test_type in "${test_types[@]}"; do
        if [ -f "test/reports/${test_type}.log" ]; then
            if grep -q "PASS" "test/reports/${test_type}.log"; then
                echo "- ✅ ${test_type} tests: PASSED" >> test/reports/summary.md
            else
                echo "- ❌ ${test_type} tests: FAILED" >> test/reports/summary.md
            fi
        else
            echo "- ⏭️ ${test_type} tests: SKIPPED" >> test/reports/summary.md
        fi
    done
    
    if [ -f test/coverage/merged.out ]; then
        COVERAGE=$(go tool cover -func=test/coverage/merged.out | grep total | awk '{print $3}')
        echo "" >> test/reports/summary.md
        echo "## Coverage" >> test/reports/summary.md
        echo "Total coverage: ${COVERAGE}" >> test/reports/summary.md
    fi
    
    log_success "Test report generated: test/reports/summary.md"
}

# Main execution
main() {
    local test_type="${1:-all}"
    local exit_code=0
    
    log_info "Starting AI Dependency Manager test suite"
    log_info "Test type: $test_type"
    
    check_dependencies
    setup_test_environment
    
    case "$test_type" in
        "unit")
            run_unit_tests || exit_code=1
            ;;
        "integration")
            run_integration_tests || exit_code=1
            ;;
        "e2e")
            run_e2e_tests || exit_code=1
            ;;
        "security")
            run_security_tests || exit_code=1
            ;;
        "race")
            run_race_tests || exit_code=1
            ;;
        "benchmark")
            run_benchmark_tests
            ;;
        "lint")
            run_linting || exit_code=1
            ;;
        "coverage")
            run_unit_tests || exit_code=1
            run_integration_tests || exit_code=1
            generate_coverage_report
            ;;
        "all")
            run_linting || exit_code=1
            run_unit_tests || exit_code=1
            run_integration_tests || exit_code=1
            run_e2e_tests || exit_code=1
            run_security_tests || exit_code=1
            run_race_tests || exit_code=1
            run_benchmark_tests
            generate_coverage_report
            ;;
        *)
            log_error "Unknown test type: $test_type"
            echo "Usage: $0 [unit|integration|e2e|security|race|benchmark|lint|coverage|all]"
            exit 1
            ;;
    esac
    
    generate_test_report
    cleanup_test_environment
    
    if [ $exit_code -eq 0 ]; then
        log_success "All tests completed successfully"
    else
        log_error "Some tests failed"
    fi
    
    exit $exit_code
}

# Run main function with all arguments
main "$@"
