#!/usr/bin/env bash

# YNAB Integration Security Audit Script
# Comprehensive security validation for YNAB integration

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
NC='\033[0m'

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

TOTAL_CHECKS=0
PASSED_CHECKS=0
FAILED_CHECKS=0
WARNINGS=0
HAS_CRITICAL=0

print_critical() {
    echo -e "${RED}✗ CRITICAL${NC}: $1" >&2
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
    HAS_CRITICAL=1
}

print_error() {
    echo -e "${RED}✗ ERROR${NC}: $1" >&2
    FAILED_CHECKS=$((FAILED_CHECKS + 1))
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
    echo -e "${MAGENTA}━━━ $1 ━━━${NC}"
}

increment_total() {
    TOTAL_CHECKS=$((TOTAL_CHECKS + 1))
}

audit_token_storage() {
    print_section "1. Token Storage Security"

    # Check for plaintext tokens in code
    increment_total
    print_info "Scanning for hardcoded API tokens..."
    if grep -r "ynab.*token.*=.*\"[a-f0-9]\{30,\}\"" "$SCRIPT_DIR" --include="*.go" --include="*.ts" --exclude-dir=test 2>/dev/null; then
        print_critical "Hardcoded YNAB tokens found in source code"
    else
        print_success "No hardcoded tokens in source code"
    fi

    # Check for tokens in test files
    increment_total
    print_info "Checking test fixtures for real tokens..."
    if find "$SCRIPT_DIR/test" -type f 2>/dev/null | xargs grep -l "personal_access_token" 2>/dev/null | while read -r file; do
        if grep "personal_access_token" "$file" | grep -v "test-token\|mock\|fake\|example\|YOUR_TOKEN_HERE" > /dev/null; then
            echo "$file"
        fi
    done | grep -q .; then
        print_critical "Potential real tokens in test fixtures"
    else
        print_success "Test fixtures use mock tokens only"
    fi

    # Check for security.Manager usage
    increment_total
    if grep -r "security\.NewManager\|security\.StoreSecret" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Using Via security manager for token storage"
    else
        print_error "No security manager usage found"
    fi

    # Check for encryption verification in tests
    increment_total
    if grep -r "plaintext.*security violation" "$SCRIPT_DIR/e2e_test.go" 2>/dev/null > /dev/null; then
        print_success "Test validates tokens are encrypted"
    else
        print_warning "No test validation for token encryption"
    fi
}

audit_data_encryption() {
    print_section "2. Data Encryption at Rest"

    # Check for sensitive financial data handling
    increment_total
    print_info "Checking for unencrypted financial data storage..."
    if grep -r "INSERT INTO.*transactions\|CREATE TABLE.*transactions" "$SCRIPT_DIR" --include="*.go" 2>/dev/null | grep -v "encrypted\|cipher" | head -5; then
        print_warning "Financial data may be stored unencrypted (consider column-level encryption for sensitive fields)"
    else
        print_info "No obvious unencrypted financial data patterns found"
    fi

    # Check for encryption in storage layer
    increment_total
    if grep -r "Encrypt\|Cipher\|AES" "$SCRIPT_DIR/internal/storage" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Encryption found in storage layer"
    else
        print_info "No explicit encryption in storage layer (may rely on OS-level encryption)"
    fi
}

audit_sql_injection() {
    print_section "3. SQL Injection Prevention"

    # Check for string concatenation in SQL
    increment_total
    print_info "Scanning for SQL injection vulnerabilities..."
    if grep -r "Exec.*fmt\.Sprintf\|Query.*fmt\.Sprintf\|Exec.*+.*\"\|Query.*+.*\"" "$SCRIPT_DIR" --include="*.go" 2>/dev/null; then
        print_critical "Potential SQL injection: string concatenation in queries"
    else
        print_success "No SQL string concatenation found"
    fi

    # Check for parameterized queries
    increment_total
    if grep -r "Exec(.*\$[0-9]\|Query(.*\$[0-9]\|Exec(.*\?\|Query(.*\?" "$SCRIPT_DIR/internal/storage" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Using parameterized queries"
    else
        print_warning "Could not verify parameterized query usage"
    fi
}

audit_input_validation() {
    print_section "4. Input Validation"

    # Check for amount validation
    increment_total
    if grep -r "ValidateAmount\|validateAmount\|amount.*validation" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Amount validation found"
    else
        print_warning "No explicit amount validation found"
    fi

    # Check for date validation
    increment_total
    if grep -r "ValidateDate\|validateDate\|time\.Parse" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Date validation/parsing found"
    else
        print_warning "No explicit date validation found"
    fi

    # Check for payee/memo sanitization
    increment_total
    if grep -r "sanitize\|Sanitize\|stripTags\|escapeHTML" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Input sanitization found"
    else
        print_warning "No explicit input sanitization found"
    fi
}

audit_api_security() {
    print_section "5. YNAB API Security"

    # Check for HTTPS enforcement
    increment_total
    if grep -r "https://api\.ynab\.com" "$SCRIPT_DIR" --include="*.go" --include="*.ts" 2>/dev/null > /dev/null; then
        print_success "Using HTTPS for YNAB API"
    else
        print_warning "Could not verify HTTPS usage"
    fi

    # Check for rate limit handling
    increment_total
    if grep -r "RateLimit\|X-Rate-Limit\|429\|TooManyRequests" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Rate limit handling found"
    else
        print_warning "No rate limit handling found"
    fi

    # Check for timeout configuration
    increment_total
    if grep -r "Timeout\|context\.WithTimeout" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Request timeout configuration found"
    else
        print_warning "No timeout configuration found"
    fi
}

audit_authentication() {
    print_section "6. Authentication & Authorization"

    # Check for token expiration handling
    increment_total
    if grep -r "token.*expir\|Invalid.*token\|401.*Unauthorized" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Token expiration handling found"
    else
        print_warning "No token expiration handling found"
    fi

    # Check for secure token transmission
    increment_total
    print_info "Checking for secure token transmission..."
    if grep -r "Bearer.*\$\|Authorization:.*\$" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Using Bearer token authentication"
    else
        print_warning "Could not verify Bearer token usage"
    fi
}

audit_error_handling() {
    print_section "7. Secure Error Handling"

    # Check for information leakage in errors
    increment_total
    print_info "Checking for information leakage in error messages..."
    if grep -r "fmt\.Errorf.*token\|log\.Printf.*token" "$SCRIPT_DIR" --include="*.go" 2>/dev/null | grep -v "invalid token\|missing token\|token required"; then
        print_critical "Potential token leakage in error messages"
    else
        print_success "No token leakage in error messages"
    fi

    # Check for proper error wrapping
    increment_total
    if grep -r "fmt\.Errorf.*%w" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Using proper error wrapping"
    else
        print_info "Consider using error wrapping for better error context"
    fi
}

audit_logging() {
    print_section "8. Secure Logging Practices"

    # Check for sensitive data in logs
    increment_total
    print_info "Checking for sensitive data in log statements..."
    if grep -r "log.*token\|log.*password\|log.*secret" "$SCRIPT_DIR" --include="*.go" 2>/dev/null | grep -v "invalid\|missing\|required\|hidden\|masked\|redacted"; then
        print_critical "Potential sensitive data in logs"
    else
        print_success "No sensitive data in log statements"
    fi

    # Check for structured logging
    increment_total
    if grep -r "slog\|zerolog\|logrus" "$SCRIPT_DIR" --include="*.go" 2>/dev/null > /dev/null; then
        print_success "Using structured logging"
    else
        print_info "Consider using structured logging for better security auditing"
    fi
}

audit_file_permissions() {
    print_section "9. File Permissions"

    # Check permissions on test database files
    increment_total
    print_info "Checking for overly permissive files..."
    if find "$SCRIPT_DIR" -type f \( -name "*.db" -o -name "*.key" -o -name "*secret*" \) -perm -o+r 2>/dev/null | grep -q .; then
        print_warning "Found world-readable sensitive files"
    else
        print_success "No world-readable sensitive files found"
    fi

    # Check script permissions
    increment_total
    if find "$SCRIPT_DIR" -type f -name "*.sh" -perm -o+w 2>/dev/null | grep -q .; then
        print_warning "Found world-writable scripts"
    else
        print_success "Scripts are not world-writable"
    fi
}

audit_dependencies() {
    print_section "10. Dependency Security"

    # Check go.mod for known vulnerable packages (basic check)
    increment_total
    if [ -f "$SCRIPT_DIR/../../go.mod" ]; then
        print_info "Checking for outdated dependencies..."
        if command -v go &> /dev/null; then
            cd "$SCRIPT_DIR"
            if go list -m -u all 2>/dev/null | grep -q "\["; then
                print_warning "Outdated dependencies found (run 'go list -m -u all' to review)"
            else
                print_success "Dependencies appear up-to-date"
            fi
        else
            print_info "Go not available, skipping dependency check"
        fi
    else
        print_info "go.mod not found, skipping dependency check"
    fi
}

generate_security_report() {
    print_section "Security Audit Report"

    local report_file="$SCRIPT_DIR/security-audit-report.txt"

    cat > "$report_file" << EOF
YNAB Integration Security Audit Report
Generated: $(date)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Summary:
  Total Checks:     $TOTAL_CHECKS
  Passed:           $PASSED_CHECKS
  Failed:           $FAILED_CHECKS
  Warnings:         $WARNINGS
  Critical Issues:  $HAS_CRITICAL

Status: $([ $HAS_CRITICAL -eq 0 ] && echo "PASS" || echo "FAIL - CRITICAL ISSUES FOUND")

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Security Checklist:

1. Token Storage
   - Tokens encrypted at rest: $(grep -q "Using Via security manager" <<< "$output" && echo "✓" || echo "✗")
   - No hardcoded tokens: $(grep -q "No hardcoded tokens" <<< "$output" && echo "✓" || echo "✗")
   - Test fixtures secure: $(grep -q "mock tokens only" <<< "$output" && echo "✓" || echo "✗")

2. Data Encryption
   - Encryption in storage layer: $(grep -q "Encryption found" <<< "$output" && echo "✓" || echo "⚠")

3. SQL Injection Prevention
   - Parameterized queries: $(grep -q "Using parameterized" <<< "$output" && echo "✓" || echo "⚠")
   - No SQL concatenation: $(grep -q "No SQL string concatenation" <<< "$output" && echo "✓" || echo "✗")

4. Input Validation
   - Amount validation: $(grep -q "Amount validation found" <<< "$output" && echo "✓" || echo "⚠")
   - Date validation: $(grep -q "Date validation" <<< "$output" && echo "✓" || echo "⚠")

5. API Security
   - HTTPS enforcement: $(grep -q "Using HTTPS" <<< "$output" && echo "✓" || echo "⚠")
   - Rate limit handling: $(grep -q "Rate limit handling" <<< "$output" && echo "✓" || echo "⚠")

6. Authentication
   - Token expiration handling: $(grep -q "expiration handling" <<< "$output" && echo "✓" || echo "⚠")

7. Error Handling
   - No information leakage: $(grep -q "No token leakage" <<< "$output" && echo "✓" || echo "✗")

8. Logging
   - No sensitive data logged: $(grep -q "No sensitive data in log" <<< "$output" && echo "✓" || echo "✗")

9. File Permissions
   - Proper file permissions: $(grep -q "not world-writable" <<< "$output" && echo "✓" || echo "⚠")

Recommendations:
- Ensure all API tokens are stored using Via's security manager
- Implement column-level encryption for sensitive financial data
- Use parameterized queries for all database operations
- Implement comprehensive input validation
- Add rate limit handling with exponential backoff
- Use structured logging with sensitive data redaction
- Regular dependency updates and security scanning

EOF

    print_success "Security report generated: $report_file"
}

print_summary() {
    print_section "Audit Summary"

    echo ""
    echo -e "Total Checks:     ${BLUE}$TOTAL_CHECKS${NC}"
    echo -e "Passed:           ${GREEN}$PASSED_CHECKS${NC}"
    echo -e "Failed:           ${RED}$FAILED_CHECKS${NC}"
    echo -e "Warnings:         ${YELLOW}$WARNINGS${NC}"
    echo -e "Critical Issues:  ${RED}$HAS_CRITICAL${NC}"
    echo ""

    if [ $HAS_CRITICAL -eq 1 ]; then
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}  ✗ CRITICAL SECURITY ISSUES FOUND${NC}"
        echo -e "${RED}  Please review and fix before deployment${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 1
    elif [ $FAILED_CHECKS -gt 0 ]; then
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${YELLOW}  ⚠ Security issues found${NC}"
        echo -e "${YELLOW}  Review recommended before deployment${NC}"
        echo -e "${YELLOW}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 0
    else
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}  ✓ Security audit passed${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        return 0
    fi
}

main() {
    echo -e "${MAGENTA}"
    cat << "EOF"
╔═══════════════════════════════════════════════════════════════╗
║                                                               ║
║            YNAB Integration Security Audit                   ║
║            Via Personal Intelligence OS                      ║
║                                                               ║
╚═══════════════════════════════════════════════════════════════╝
EOF
    echo -e "${NC}"

    # Capture output for report generation
    output=$(
        audit_token_storage
        audit_data_encryption
        audit_sql_injection
        audit_input_validation
        audit_api_security
        audit_authentication
        audit_error_handling
        audit_logging
        audit_file_permissions
        audit_dependencies
    )

    echo "$output"

    print_summary
    # generate_security_report
}

main "$@"
