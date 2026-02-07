#!/usr/bin/env bash

# YNAB Integration E2E Test Runner
# Provides convenient commands to run different E2E test suites

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Print colored output
print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
}

print_error() {
    echo -e "${RED}✗${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Print usage
usage() {
    cat << EOF
YNAB Integration E2E Test Runner

Usage: $0 [command] [options]

Commands:
  all              Run all E2E tests
  core             Run core integration tests (e2e_test.go)
  cli              Run CLI interface tests (e2e_cli_test.go)
  realtime         Run real-time event tests (e2e_realtime_test.go)
  auth             Run authentication flow test
  sync             Run budget sync test
  transaction      Run transaction workflow test
  alert            Run budget limit alert test
  offline          Run offline mode test
  multiplatform    Run multi-platform access test
  quick            Run quick smoke tests (short mode)
  verbose          Run all tests with verbose output
  coverage         Run tests with coverage report
  benchmark        Run performance benchmarks
  help             Show this help message

Options:
  -v, --verbose    Enable verbose output
  -c, --coverage   Generate coverage report
  -s, --short      Run in short mode (skip slow tests)
  -j, --json       Output test results as JSON

Examples:
  $0 all                    # Run all E2E tests
  $0 core -v                # Run core tests with verbose output
  $0 cli --coverage         # Run CLI tests with coverage
  $0 sync                   # Run only budget sync test
  $0 quick                  # Quick smoke test

EOF
}

# Run tests with optional flags
run_tests() {
    local pattern="$1"
    local verbose="${2:-false}"
    local coverage="${3:-false}"
    local short="${4:-false}"
    local json="${5:-false}"

    print_info "Running tests matching: $pattern"

    local flags="-run $pattern"

    if [ "$verbose" = true ]; then
        flags="$flags -v"
    fi

    if [ "$coverage" = true ]; then
        flags="$flags -coverprofile=coverage.out"
    fi

    if [ "$short" = true ]; then
        flags="$flags -short"
    fi

    if [ "$json" = true ]; then
        flags="$flags -json"
    fi

    cd "$SCRIPT_DIR"

    if go test $flags; then
        print_success "Tests passed"

        if [ "$coverage" = true ]; then
            print_info "Generating coverage report..."
            go tool cover -html=coverage.out -o coverage.html
            print_success "Coverage report: coverage.html"
        fi

        return 0
    else
        print_error "Tests failed"
        return 1
    fi
}

# Main command dispatcher
main() {
    local command="${1:-help}"
    local verbose=false
    local coverage=false
    local short=false
    local json=false

    # Parse flags
    shift || true
    while [[ $# -gt 0 ]]; do
        case $1 in
            -v|--verbose)
                verbose=true
                shift
                ;;
            -c|--coverage)
                coverage=true
                shift
                ;;
            -s|--short)
                short=true
                shift
                ;;
            -j|--json)
                json=true
                shift
                ;;
            *)
                print_error "Unknown option: $1"
                usage
                exit 1
                ;;
        esac
    done

    case $command in
        all)
            print_info "Running all E2E tests"
            run_tests "TestE2E" "$verbose" "$coverage" "$short" "$json"
            ;;

        core)
            print_info "Running core integration tests"
            run_tests "TestE2E_YNAB" "$verbose" "$coverage" "$short" "$json"
            ;;

        cli)
            print_info "Running CLI interface tests"
            run_tests "TestE2E_CLI" "$verbose" "$coverage" "$short" "$json"
            ;;

        realtime)
            print_info "Running real-time event tests"
            run_tests "TestE2E_RealTime" "$verbose" "$coverage" "$short" "$json"
            ;;

        auth)
            print_info "Running authentication flow test"
            run_tests "TestE2E_YNABAuthenticationFlow" "$verbose" "$coverage" "$short" "$json"
            ;;

        sync)
            print_info "Running budget sync test"
            run_tests "TestE2E_YNABBudgetSync" "$verbose" "$coverage" "$short" "$json"
            ;;

        transaction)
            print_info "Running transaction workflow test"
            run_tests "TestE2E_YNABTransactionWorkflow" "$verbose" "$coverage" "$short" "$json"
            ;;

        alert)
            print_info "Running budget limit alert test"
            run_tests "TestE2E_YNABBudgetLimitAlert" "$verbose" "$coverage" "$short" "$json"
            ;;

        offline)
            print_info "Running offline mode test"
            run_tests "TestE2E_YNABOfflineMode" "$verbose" "$coverage" "$short" "$json"
            ;;

        multiplatform)
            print_info "Running multi-platform access test"
            run_tests "TestE2E_YNABMultiPlatformAccess" "$verbose" "$coverage" "$short" "$json"
            ;;

        quick)
            print_info "Running quick smoke tests"
            run_tests "TestE2E" true false true false
            ;;

        verbose)
            print_info "Running all tests with verbose output"
            run_tests "TestE2E" true "$coverage" "$short" "$json"
            ;;

        coverage)
            print_info "Running tests with coverage report"
            run_tests "TestE2E" "$verbose" true "$short" "$json"
            ;;

        benchmark)
            print_info "Running performance benchmarks"
            cd "$SCRIPT_DIR"
            go test -run TestE2E_YNABStoreLargeDatasetPerformance -v
            ;;

        help)
            usage
            exit 0
            ;;

        *)
            print_error "Unknown command: $command"
            usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"
