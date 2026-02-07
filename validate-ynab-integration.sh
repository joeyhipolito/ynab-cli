#!/usr/bin/env bash

# YNAB Integration Validation Script
# Validates YNAB skill structure, tests, security, and performance

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VIA_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Counters
TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0
HAS_ERRORS=0

# Helper functions
print_error() {
    echo -e "${RED}✗ ERROR${NC}: $1" >&2
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
    HAS_ERRORS=1
}

print_warning() {
    echo -e "${YELLOW}⚠ WARNING${NC}: $1"
    WARNINGS=$((WARNINGS + 1))
}

print_success() {
    echo -e "${GREEN}✓${NC} $1"
    PASSED_CHECKS=$((PASSED_CHECKS + 1))
}

print_info() {
    echo -e "${BLUE}ℹ${NC} $1"
}

print_section() {
    echo ""
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${MAGENTA}  $1${NC}"
    echo -e "${MAGENTA}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
}

increment_total() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

# Validation functions
validate_skill_structure() {
    print_section "1. YNAB Skill Structure Validation"

    # Check if skill directory exists
    increment_total
    if [ -d "$VIA_ROOT/.claude/skills/ynab" ]; then
        print_success "Skill directory exists"
    else
        print_warning "Skill directory not found at .claude/skills/ynab"
    fi

    # Check for SKILL.md (will be created in implementation)
    increment_total
    if [ -f "$VIA_ROOT/.claude/skills/ynab/SKILL.md" ]; then
        print_success "SKILL.md exists"

        # Validate SKILL.md structure
        increment_total
        if grep -q "^# YNAB" "$VIA_ROOT/.claude/skills/ynab/SKILL.md"; then
            print_success "SKILL.md has proper title"
        else
            print_error "SKILL.md missing proper title"
        fi
    else
        print_warning "SKILL.md not found (expected during implementation)"
    fi

    # Check for scripts directory
    increment_total
    if [ -d "$VIA_ROOT/.claude/skills/ynab/scripts" ]; then
        print_success "scripts/ directory exists"
    else
        print_warning "scripts/ directory not found"
    fi

    # Check for references directory
    increment_total
    if [ -d "$VIA_ROOT/.claude/skills/ynab/references" ]; then
        print_success "references/ directory exists"
    else
        print_warning "references/ directory not found"
    fi
}

validate_test_structure() {
    print_section "2. Test Structure Validation"

    # Check for test directory
    increment_total
    if [ -d "$SCRIPT_DIR/test" ]; then
        print_success "test/ directory exists"
    else
        print_error "test/ directory not found"
        return
    fi

    # Check for test fixtures
    increment_total
    if [ -d "$SCRIPT_DIR/test/fixtures" ]; then
        print_success "test/fixtures/ directory exists"

        # Check for API fixtures
        increment_total
        if [ -d "$SCRIPT_DIR/test/fixtures/api" ]; then
            print_success "API fixtures directory exists"

            # Count API fixture files
            api_fixture_count=$(find "$SCRIPT_DIR/test/fixtures/api" -type f -name "*.json" | wc -l | tr -d ' ')
            increment_total
            if [ "$api_fixture_count" -ge 5 ]; then
                print_success "Found $api_fixture_count API fixture files"
            else
                print_warning "Only $api_fixture_count API fixtures (expected 5+)"
            fi
        else
            print_error "API fixtures directory not found"
        fi

        # Check for database fixtures
        increment_total
        if [ -d "$SCRIPT_DIR/test/fixtures/database" ]; then
            print_success "Database fixtures directory exists"
        else
            print_error "Database fixtures directory not found"
        fi

        # Check for config fixtures
        increment_total
        if [ -d "$SCRIPT_DIR/test/fixtures/config" ]; then
            print_success "Config fixtures directory exists"
        else
            print_error "Config fixtures directory not found"
        fi
    else
        print_error "test/fixtures/ directory not found"
    fi

    # Check for test helpers
    increment_total
    if [ -d "$SCRIPT_DIR/test/helpers" ]; then
        print_success "test/helpers/ directory exists"
    else
        print_error "test/helpers/ directory not found"
    fi

    # Check for E2E test files
    increment_total
    if [ -f "$SCRIPT_DIR/e2e_test.go" ]; then
        print_success "e2e_test.go exists"
    else
        print_error "e2e_test.go not found"
    fi

    increment_total
    if [ -f "$SCRIPT_DIR/e2e_cli_test.go" ]; then
        print_success "e2e_cli_test.go exists"
    else
        print_error "e2e_cli_test.go not found"
    fi

    increment_total
    if [ -f "$SCRIPT_DIR/e2e_realtime_test.go" ]; then
        print_success "e2e_realtime_test.go exists"
    else
        print_error "e2e_realtime_test.go not found"
    fi
}

validate_integration_tests() {
    print_section "3. Integration Test Validation"

    # Check for storage integration tests
    increment_total
    if [ -f "$SCRIPT_DIR/internal/storage/integration_test.go" ]; then
        print_success "Storage integration tests exist"

        # Check for key test functions
        increment_total
        if grep -q "TestYNABStoreConcurrentWrites" "$SCRIPT_DIR/internal/storage/integration_test.go"; then
            print_success "Concurrent writes test found"
        else
            print_warning "TestYNABStoreConcurrentWrites not found"
        fi
    else
        print_error "Storage integration tests not found"
    fi

    # Check for event integration tests
    increment_total
    if [ -f "$SCRIPT_DIR/internal/events/ynab_events_test.go" ]; then
        print_success "Event integration tests exist"
    else
        print_error "Event integration tests not found"
    fi

    # Check for security integration tests
    increment_total
    if [ -f "$SCRIPT_DIR/internal/security/ynab_security_test.go" ]; then
        print_success "Security integration tests exist"
    else
        print_error "Security integration tests not found"
    fi
}

validate_security() {
    print_section "4. Security Audit"

    # Check for sensitive data in fixtures
    increment_total
    print_info "Scanning for hardcoded tokens in fixtures..."
    if grep -r "personal_access_token" "$SCRIPT_DIR/test/fixtures/" 2>/dev/null | grep -v "test-token\|mock\|fake\|example" > /dev/null; then
        print_error "Potential real tokens found in fixtures"
    else
        print_success "No real tokens found in fixtures"
    fi

    # Check for security manager usage in tests
    increment_total
    if grep -r "security.NewManager\|secMgr" "$SCRIPT_DIR"/*.go 2>/dev/null > /dev/null; then
        print_success "Security manager usage found in tests"
    else
        print_warning "No security manager usage found in tests"
    fi

    # Check for encryption verification in tests
    increment_total
    if grep -r "Token stored in plaintext" "$SCRIPT_DIR/e2e_test.go" 2>/dev/null > /dev/null; then
        print_success "Encryption verification found in tests"
    else
        print_warning "No encryption verification found in tests"
    fi

    # Check for sensitive data in Go source files
    increment_total
    print_info "Scanning for hardcoded credentials in source..."
    if grep -r "ynab.*token.*=.*\"[a-f0-9]\{20,\}\"" "$SCRIPT_DIR" 2>/dev/null > /dev/null; then
        print_error "Potential hardcoded credentials found"
    else
        print_success "No hardcoded credentials found"
    fi
}

validate_test_coverage() {
    print_section "5. Test Coverage Analysis"

    print_info "Checking if Go tests can compile..."

    increment_total
    cd "$SCRIPT_DIR"
    if go test -run=^$ ./... 2>&1 | grep -q "no test files\|PASS\|FAIL"; then
        print_success "Go tests compile successfully"

        # Try to get coverage if tests exist
        print_info "Attempting to measure test coverage..."
        coverage_output=$(go test -cover ./... 2>&1 || true)

        if echo "$coverage_output" | grep -q "coverage:"; then
            coverage_pct=$(echo "$coverage_output" | grep -oP 'coverage: \K[0-9.]+' | head -1 || echo "0")
            increment_total
            if [ -n "$coverage_pct" ]; then
                print_info "Current test coverage: ${coverage_pct}%"
                if (( $(echo "$coverage_pct >= 50" | bc -l 2>/dev/null || echo 0) )); then
                    print_success "Coverage meets minimum threshold (50%)"
                else
                    print_warning "Coverage below recommended threshold: ${coverage_pct}% (target: 50%+)"
                fi
            fi
        else
            print_info "Coverage measurement skipped (tests may not be implemented yet)"
        fi
    else
        print_error "Go tests do not compile"
    fi
}

validate_documentation() {
    print_section "6. Documentation Validation"

    # Check for README in test directory
    increment_total
    if [ -f "$SCRIPT_DIR/test/README.md" ]; then
        print_success "Test README.md exists"
    else
        print_error "Test README.md not found"
    fi

    # Check for E2E test documentation
    increment_total
    if [ -f "$SCRIPT_DIR/E2E_TESTS.md" ]; then
        print_success "E2E_TESTS.md exists"
    else
        print_error "E2E_TESTS.md not found"
    fi

    # Check for test runner script
    increment_total
    if [ -f "$SCRIPT_DIR/run-e2e-tests.sh" ] && [ -x "$SCRIPT_DIR/run-e2e-tests.sh" ]; then
        print_success "run-e2e-tests.sh exists and is executable"
    else
        print_warning "run-e2e-tests.sh not found or not executable"
    fi
}

validate_performance_benchmarks() {
    print_section "7. Performance Benchmark Validation"

    # Check for performance test functions
    increment_total
    if grep -r "Performance\|Benchmark" "$SCRIPT_DIR"/*.go 2>/dev/null > /dev/null; then
        print_success "Performance tests found"

        # List performance tests
        print_info "Performance tests found:"
        grep -h "func.*Performance\|func.*Benchmark" "$SCRIPT_DIR"/*.go 2>/dev/null | sed 's/func /  - /' | sed 's/(.*$//' || true
    else
        print_warning "No performance tests found"
    fi

    # Check for benchmark command in test runner
    increment_total
    if [ -f "$SCRIPT_DIR/run-e2e-tests.sh" ]; then
        if grep -q "benchmark" "$SCRIPT_DIR/run-e2e-tests.sh"; then
            print_success "Benchmark command found in test runner"
        else
            print_warning "No benchmark command in test runner"
        fi
    fi
}

validate_via_integration() {
    print_section "8. Via Feature Integration Validation"

    # Check for events integration
    increment_total
    if grep -r "via/features/events" "$SCRIPT_DIR"/*.go 2>/dev/null > /dev/null; then
        print_success "Events feature integration found"
    else
        print_warning "Events feature integration not found"
    fi

    # Check for security integration
    increment_total
    if grep -r "via/features/security" "$SCRIPT_DIR"/*.go 2>/dev/null > /dev/null; then
        print_success "Security feature integration found"
    else
        print_warning "Security feature integration not found"
    fi

    # Check for storage integration
    increment_total
    if grep -r "via/features/storage" "$SCRIPT_DIR"/*.go 2>/dev/null > /dev/null; then
        print_success "Storage feature integration found"
    else
        print_info "Storage feature integration not found (using internal storage)"
    fi

    # Check for gateway integration (future)
    increment_total
    if grep -r "via/features/gateway" "$SCRIPT_DIR"/*.go 2>/dev/null > /dev/null; then
        print_success "Gateway feature integration found"
    else
        print_info "Gateway feature integration not yet implemented"
    fi
}

print_summary() {
    print_section "Validation Summary"

    echo ""
    echo -e "Total Checks:   ${BLUE}$TOTAL_CHECKS${NC}"
    echo -e "Passed:         ${GREEN}$PASSED_CHECKS${NC}"
    echo -e "Failed:         ${RED}$FAILED_CHECKS${NC}"
    echo -e "Warnings:       ${YELLOW}$WARNINGS${NC}"
    echo ""

    if [ $HAS_ERRORS -eq 0 ]; then
        if [ $WARNINGS -eq 0 ]; then
            echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
            echo -e "${GREEN}  ✓ All validations passed!${NC}"
            echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        else
            echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
            echo -e "${YELLOW}  ✓ Validation passed with warnings${NC}"
            echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        fi
        return 0
    else
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}  ✗ Validation failed${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 1
    fi
}

# Main execution
main() {
    echo -e "${MAGENTA}"
    cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║        YNAB Integration Validation Suite                     ║
║        Via Personal Intelligence OS                          ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"

    validate_skill_structure
    validate_test_structure
    validate_integration_tests
    validate_security
    validate_test_coverage
    validate_documentation
    validate_performance_benchmarks
    validate_via_integration
    print_summary
}

# Run main
main "$@"
