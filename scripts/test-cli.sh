#!/bin/bash
# Integration test script for ynab-cli
# Requires YNAB_ACCESS_TOKEN to be set

set -e

echo "========================================="
echo "YNAB CLI Integration Tests"
echo "========================================="
echo ""

# Check if token is set
if [ -z "$YNAB_ACCESS_TOKEN" ]; then
    echo "ERROR: YNAB_ACCESS_TOKEN environment variable is required"
    echo "Set it with: export YNAB_ACCESS_TOKEN='your-token-here'"
    exit 1
fi

# Check if binary exists
if [ ! -f ~/bin/ynab-cli ]; then
    echo "ERROR: ynab-cli not found at ~/bin/ynab-cli"
    echo "Run: make install"
    exit 1
fi

echo "✓ Token is set (${YNAB_ACCESS_TOKEN:0:10}...)"
echo "✓ Binary found at ~/bin/ynab-cli"
echo ""

# Test 1: Status command (human-readable)
echo "========================================="
echo "Test 1: status (human-readable)"
echo "========================================="
~/bin/ynab-cli status
echo ""

# Test 2: Status command (JSON)
echo "========================================="
echo "Test 2: status --json"
echo "========================================="
~/bin/ynab-cli status --json | jq '.' 2>/dev/null || ~/bin/ynab-cli status --json
echo ""

# Test 3: Balance command (human-readable)
echo "========================================="
echo "Test 3: balance (human-readable)"
echo "========================================="
~/bin/ynab-cli balance
echo ""

# Test 4: Balance command (JSON)
echo "========================================="
echo "Test 4: balance --json"
echo "========================================="
~/bin/ynab-cli balance --json | jq '.' 2>/dev/null || ~/bin/ynab-cli balance --json
echo ""

# Test 5: Budget command (human-readable)
echo "========================================="
echo "Test 5: budget (human-readable)"
echo "========================================="
~/bin/ynab-cli budget
echo ""

# Test 6: Budget command (JSON)
echo "========================================="
echo "Test 6: budget --json"
echo "========================================="
~/bin/ynab-cli budget --json | jq '.' 2>/dev/null || ~/bin/ynab-cli budget --json
echo ""

# Test 7: Categories command (human-readable)
echo "========================================="
echo "Test 7: categories (human-readable)"
echo "========================================="
~/bin/ynab-cli categories
echo ""

# Test 8: Categories command (JSON)
echo "========================================="
echo "Test 8: categories --json"
echo "========================================="
~/bin/ynab-cli categories --json | jq '.' 2>/dev/null || ~/bin/ynab-cli categories --json
echo ""

# Test 9: Balance with filter (human-readable)
echo "========================================="
echo "Test 9: balance checking (filter test)"
echo "========================================="
~/bin/ynab-cli balance checking
echo ""

# Test 10: Help and version
echo "========================================="
echo "Test 10: --help and --version"
echo "========================================="
~/bin/ynab-cli --help > /dev/null && echo "✓ --help works"
~/bin/ynab-cli --version && echo "✓ --version works"
echo ""

echo "========================================="
echo "All tests completed successfully!"
echo "========================================="
