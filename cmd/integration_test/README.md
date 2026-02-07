# YNAB API Integration Test

This test validates the YNAB API client methods against the real YNAB API.

## Prerequisites

You need a valid YNAB Personal Access Token set in your environment:

```bash
export YNAB_ACCESS_TOKEN="your-token-here"
```

To get a token:
1. Go to https://app.youneedabudget.com/settings/developer
2. Click "New Token"
3. Give it a name (e.g., "Via CLI")
4. Copy the token

## Running the Integration Test

```bash
# Build the test
go build -tags=integration -o /tmp/ynab-integration-test ./cmd/integration_test

# Run it
/tmp/ynab-integration-test
```

Or directly:

```bash
go run -tags=integration ./cmd/integration_test
```

## What It Tests

1. **GetBudgets()** - Retrieves all budgets for your account
2. **GetDefaultBudgetID()** - Gets the first budget as default
3. **GetAccounts(budgetID)** - Lists all accounts for a budget
4. **GetAccounts("")** - Tests using default budget when ID is empty

## Expected Output

```
Testing YNAB API Client Integration
====================================

1. Testing GetBudgets()...
✓ Found 1 budget(s)
  1. My Budget (ID: abc-123-def)
     Last Modified: 2026-02-02T10:00:00Z

2. Testing GetDefaultBudgetID()...
✓ Default Budget ID: abc-123-def

3. Testing GetAccounts()...
✓ Found 3 account(s)
  1. Checking (checking) [Open]
     Balance: $1234.56
     Cleared Balance: $1200.00
  2. Savings (savings) [Open]
     Balance: $5000.00
     Cleared Balance: $5000.00
  3. Credit Card (creditCard) [Open]
     Balance: -$500.00
     Cleared Balance: -$450.00

4. Testing GetAccounts("") - using default budget...
✓ Found 3 account(s) using default budget

====================================
All tests passed!
```
