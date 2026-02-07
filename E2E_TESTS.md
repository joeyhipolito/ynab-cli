# YNAB Integration End-to-End Tests

This document describes the comprehensive end-to-end test suite for the YNAB integration in Via.

## Overview

The E2E test suite validates complete user workflows from CLI commands through to API interactions, storage, and real-time event propagation across multiple platforms.

## Test Files

### 1. `e2e_test.go` - Core Integration Tests

Tests fundamental YNAB integration workflows:

- **Authentication Flow** - Token storage, encryption, retrieval
- **Budget Sync** - Full sync workflow from API to local storage
- **Transaction Workflow** - Creating transactions from CLI to storage with events
- **Budget Limit Alerts** - Detection and event publishing for overspending
- **Offline Mode** - Local transaction creation and sync queue management
- **Multi-Platform Access** - Concurrent access from different interfaces
- **Error Recovery** - Error handling, categorization, and retry logic

**Key Tests:**

```go
TestE2E_YNABAuthenticationFlow          // Token encryption and secure storage
TestE2E_YNABBudgetSync                  // Complete budget sync with events
TestE2E_YNABTransactionWorkflow         // CLI -> Storage -> Events
TestE2E_YNABBudgetLimitAlert            // Overspending detection
TestE2E_YNABOfflineMode                 // Offline transaction queue
TestE2E_YNABMultiPlatformAccess         // Concurrent platform access
TestE2E_YNABErrorRecovery               // Error handling and recovery
```

### 2. `e2e_cli_test.go` - CLI Interface Tests

Tests the Via CLI commands for YNAB integration:

- **Setup Commands** - Initialization, authentication, configuration
- **Budget Commands** - List budgets, select default, view summary
- **Transaction Commands** - Add, list, filter transactions
- **Sync Commands** - Manual sync, sync status, full sync
- **Report Commands** - Category reports, monthly summaries, budget health
- **JSON Output** - Structured output for integration
- **Error Handling** - Invalid inputs, missing auth, missing arguments
- **Pipeline Integration** - Piping to jq, CSV export
- **Interactive Mode** - Interactive transaction entry, category selection

**Key Tests:**

```go
TestE2E_CLI_YNABSetup                   // via ynab init, auth, status
TestE2E_CLI_YNABBudgetCommands          // via ynab budgets list/select
TestE2E_CLI_YNABTransactionCommands     // via budget add, transactions
TestE2E_CLI_YNABSyncCommands            // via ynab sync
TestE2E_CLI_YNABReportCommands          // via budget report category/monthly
TestE2E_CLI_YNABJSONOutput              // --json flag support
TestE2E_CLI_YNABErrorHandling           // Error messages and validation
TestE2E_CLI_YNABPipelineIntegration     // Piping and CSV export
TestE2E_CLI_YNABInteractiveMode         // --interactive flag
```

### 3. `e2e_realtime_test.go` - Real-Time Event Tests

Tests real-time event-driven functionality across platforms:

- **Transaction Notifications** - Real-time updates to all platforms
- **Budget Alert Propagation** - Alert delivery with platform-specific delays
- **Sync Progress** - Real-time sync status updates
- **Concurrent Platform Writes** - Conflict detection and resolution
- **Event Persistence** - Event storage and replay capability
- **Cross-Platform Event Flow** - Event propagation from one platform to all others

**Key Tests:**

```go
TestE2E_RealTime_TransactionNotifications    // Real-time updates to all platforms
TestE2E_RealTime_BudgetAlertPropagation      // Alert delivery timing
TestE2E_RealTime_SyncProgress                // Progress event sequence
TestE2E_RealTime_ConcurrentPlatformWrites    // Concurrent write handling
TestE2E_RealTime_EventPersistence            // Event storage and replay
TestE2E_RealTime_CrossPlatformEventFlow      // Multi-platform propagation
```

## Running Tests

### Run All E2E Tests

```bash
cd features/ynab
go test -v -run TestE2E
```

### Run Specific Test Suite

```bash
# Core integration tests
go test -v -run TestE2E_YNAB

# CLI tests
go test -v -run TestE2E_CLI

# Real-time tests
go test -v -run TestE2E_RealTime
```

### Run Individual Test

```bash
go test -v -run TestE2E_YNABBudgetSync
```

### Skip E2E Tests (Short Mode)

```bash
go test -short
```

## Test Scenarios

### 1. Authentication and Security

**Scenario:** User authenticates with YNAB API token

```
User: via ynab auth --token <token>
 ↓
Token encrypted and stored securely
 ↓
Verify: Token file is encrypted (not plaintext)
 ↓
Retrieve and decrypt token
 ↓
✓ Authentication successful
```

**Tests:**
- `TestE2E_YNABAuthenticationFlow`
- `TestE2E_CLI_YNABSetup`

### 2. Budget Synchronization

**Scenario:** User syncs budget from YNAB API

```
User: via ynab sync
 ↓
Event: budget:sync:started
 ↓
Fetch budgets, accounts, categories, transactions from API
 ↓
Store locally in SQLite
 ↓
Event: budget:sync:progress (multiple)
 ↓
Event: budget:sync:completed
 ↓
All platforms receive real-time updates
 ↓
✓ Sync completed successfully
```

**Tests:**
- `TestE2E_YNABBudgetSync`
- `TestE2E_CLI_YNABSyncCommands`
- `TestE2E_RealTime_SyncProgress`

### 3. Transaction Creation

**Scenario:** User adds transaction via CLI

```
User: via budget add 25.50 "Coffee Shop" --category "Dining Out"
 ↓
Parse and validate input
 ↓
Create transaction in local storage
 ↓
Event: budget:transaction:added
 ↓
All platforms receive notification:
  - Web dashboard updates in real-time
  - Mobile app shows push notification
  - Telegram bot sends message
  - CLI confirms transaction
 ↓
✓ Transaction created and propagated
```

**Tests:**
- `TestE2E_YNABTransactionWorkflow`
- `TestE2E_CLI_YNABTransactionCommands`
- `TestE2E_RealTime_TransactionNotifications`

### 4. Budget Limit Alert

**Scenario:** User exceeds budget limit for category

```
Transaction added: -$95.00 (Groceries)
 ↓
Calculate category total: $425.00
 ↓
Compare to budget: $400.00
 ↓
Detect overspending: $25.00
 ↓
Event: budget:limit:exceeded
 ↓
Platforms respond:
  - Web: Show alert banner
  - Mobile: Push notification
  - Telegram: Send warning message
  - CLI: Display warning if active
 ↓
✓ Alert propagated to all platforms
```

**Tests:**
- `TestE2E_YNABBudgetLimitAlert`
- `TestE2E_RealTime_BudgetAlertPropagation`

### 5. Offline Mode

**Scenario:** User adds transaction while offline

```
User: via budget add 15.00 "Cash purchase" (offline)
 ↓
Create transaction in local storage
 ↓
Mark transaction for sync
 ↓
Add to sync queue
 ↓
(Later, when online)
 ↓
Sync queue processes pending transactions
 ↓
Upload to YNAB API
 ↓
Clear sync flag
 ↓
✓ Offline transaction synced
```

**Tests:**
- `TestE2E_YNABOfflineMode`

### 6. Multi-Platform Access

**Scenario:** Multiple platforms access budget simultaneously

```
Web Dashboard: Create transaction
Mobile App:    Create transaction  } Concurrent
CLI:           Create transaction  } writes
Telegram Bot:  Create transaction
 ↓
All transactions stored successfully
 ↓
Each platform receives updates about others' transactions
 ↓
Conflict detection (if any)
 ↓
✓ All platforms synchronized
```

**Tests:**
- `TestE2E_YNABMultiPlatformAccess`
- `TestE2E_RealTime_ConcurrentPlatformWrites`
- `TestE2E_RealTime_CrossPlatformEventFlow`

## Test Coverage

### Components Tested

- ✅ **Security Manager** - Token encryption and storage
- ✅ **YNAB Store** - Local SQLite storage
- ✅ **Event Bus** - Real-time event publishing and subscription
- ✅ **CLI Commands** - All via budget/ynab commands
- ✅ **Transaction Validation** - Input validation and sanitization
- ✅ **Sync Engine** - API sync and local storage updates
- ✅ **Multi-Platform** - Event propagation to all interfaces

### User Workflows Tested

- ✅ Initial setup and authentication
- ✅ Budget synchronization
- ✅ Transaction creation and management
- ✅ Budget limit monitoring
- ✅ Offline transaction handling
- ✅ Real-time notifications
- ✅ Error recovery
- ✅ CSV export and reporting
- ✅ Interactive CLI mode

## Test Data

All tests use isolated test environments:

- **Temporary directories** for each test (`t.TempDir()`)
- **Mock data** for budgets, accounts, categories, transactions
- **Test event bus** instances
- **In-memory or temporary SQLite** databases

No tests interact with:
- Real YNAB API
- User's actual Via configuration
- Production databases

## Continuous Integration

These E2E tests are designed to run in CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
name: E2E Tests
on: [push, pull_request]
jobs:
  e2e:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run E2E Tests
        run: |
          cd features/ynab
          go test -v -run TestE2E
```

## Performance Benchmarks

Expected performance characteristics:

- **Authentication**: < 100ms
- **Transaction creation**: < 50ms
- **Event propagation**: < 100ms to all platforms
- **Budget sync (100 transactions)**: < 2s
- **Query recent transactions**: < 100ms
- **CSV export (1000 transactions)**: < 500ms

Performance tests can be run with:

```bash
go test -run TestE2E_YNABStoreLargeDatasetPerformance
```

## Debugging E2E Tests

### Enable Verbose Output

```bash
go test -v -run TestE2E_YNABBudgetSync
```

### Run Single Test with Debugging

```go
// Add debugging output in tests
t.Logf("Debug: %+v", data)
```

### Check Test Artifacts

Tests create temporary directories. To inspect:

```go
// In test, before cleanup
tmpDir := t.TempDir()
t.Logf("Test artifacts in: %s", tmpDir)
time.Sleep(60 * time.Second) // Pause for inspection
```

### Simulate Test Scenarios Manually

```bash
# Setup test environment
export VIA_HOME=/tmp/via-test
mkdir -p $VIA_HOME/.via

# Run CLI commands
via ynab init
via ynab auth --token test-token
via budget add 25.50 "Test"
```

## Future Enhancements

Potential additions to the E2E test suite:

- [ ] Load testing with 10,000+ transactions
- [ ] Network failure simulation and retry logic
- [ ] API rate limiting scenarios
- [ ] Data migration from YNAB to Via format
- [ ] Multi-budget management
- [ ] Shared budget collaboration
- [ ] Transaction reconciliation
- [ ] Budget forecasting
- [ ] Recurring transaction detection

## Contributing

When adding new YNAB features:

1. Add E2E tests for the complete user workflow
2. Test both happy path and error cases
3. Test multi-platform behavior
4. Test offline mode if applicable
5. Update this documentation

## References

- [Unit Tests](internal/security/ynab_security_test.go)
- [Integration Tests](internal/storage/integration_test.go)
- [Event Tests](internal/events/ynab_events_test.go)
- [Via Skills Guide](../../docs/skills/guides/AUTHOR.md)
- [Via Architecture](../../docs/vision/architecture/MULTI-PLATFORM.md)
