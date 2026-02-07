# YNAB Integration Test Suite

Comprehensive test suite for the YNAB integration feature in Via.

## Directory Structure

```
test/
├── fixtures/           # Test data fixtures
│   ├── api/           # Mock YNAB API responses (JSON)
│   ├── database/      # Test database seed data (TypeScript)
│   └── config/        # Test configuration files
├── helpers/           # Test utilities and mocks
│   ├── test-utils.ts  # Helper functions for testing
│   ├── mocks.ts       # Mock implementations
│   └── index.ts       # Exports all helpers
├── unit/              # Unit tests
├── integration/       # Integration tests
├── e2e/              # End-to-end tests
├── setup.ts          # Global test setup
└── README.md         # This file
```

## Test Data Fixtures

### API Fixtures (`fixtures/api/`)

JSON files containing realistic YNAB API responses:

- `budgets.json` - Multiple budgets with different currencies
- `accounts.json` - Various account types (checking, savings, credit card, tracking)
- `categories.json` - Category groups with goals and spending
- `transactions.json` - Transactions including splits, transfers, and flags
- `payees.json` - Payees and transfer payees
- `months.json` - Month-by-month budget data

### Database Fixtures (`fixtures/database/`)

TypeScript modules exporting test data:

- `budgets.ts` - Mock budgets, accounts, categories, transactions, payees
- `sync-state.ts` - Mock sync states, logs, and offline queue

### Config Fixtures (`fixtures/config/`)

- `test-config.ts` - Test configuration objects
- `test-env.json` - Environment variables for different test types

## Test Helpers

### Test Utils (`helpers/test-utils.ts`)

Utility functions for testing:

```typescript
// Load fixtures
const data = loadApiFixture('budgets');

// Create test entities
const budget = createTestBudget({ name: 'My Budget' });
const account = createTestAccount({ type: 'checking' });
const transaction = createTestTransaction({ amount: -5000 });

// Helper classes
const mockDate = new MockDate('2026-02-01');
const mockHttp = new MockHttpClient();
const mockStorage = new MockStorage();
const spy = new FunctionSpy(myFunction);

// Database helpers
await setupTestDatabase(storage);
await cleanupTestDatabase(storage);
```

### Mocks (`helpers/mocks.ts`)

Mock implementations of core interfaces:

```typescript
// Create mocks
const api = new MockYnabApi();
const storage = new MockStorageAdapter();
const events = new MockEventBus();
const gateway = new MockGateway();

// Or create all at once
const context = createMockContext();

// Configure mock responses
api.mockResponse('getBudgets', mockBudgets);
storage.saveBudget(testBudget);
events.on('transaction:created', handler);

// Verify calls
expect(api.wasCalledWith('getBudget', 'budget-id')).toBe(true);
expect(events.wasEmitted('sync:complete')).toBe(true);
```

## Running Tests

```bash
# All tests
npm test

# Unit tests only
npm run test:unit

# Integration tests
npm run test:integration

# End-to-end tests
npm run test:e2e

# Watch mode
npm run test:watch

# Coverage
npm run test:coverage

# Specific test file
npm test -- transactions.test.ts
```

## Writing Tests

### Unit Test Example

```typescript
import { describe, it, expect, beforeEach } from '@jest/globals';
import { MockYnabApi, createTestBudget } from '../helpers';
import { BudgetService } from '../../src/services/budget';

describe('BudgetService', () => {
  let api: MockYnabApi;
  let service: BudgetService;

  beforeEach(() => {
    api = new MockYnabApi();
    service = new BudgetService(api);
  });

  it('should fetch budgets from API', async () => {
    const mockBudget = createTestBudget();
    api.mockResponse('getBudgets', { budgets: [mockBudget] });

    const result = await service.getBudgets();

    expect(result).toHaveLength(1);
    expect(result[0].id).toBe(mockBudget.id);
    expect(api.wasCalledWith('getBudgets')).toBe(true);
  });
});
```

### Integration Test Example

```typescript
import { describe, it, expect, beforeEach, afterEach } from '@jest/globals';
import { createMockContext, setupTestDatabase, cleanupTestDatabase } from '../helpers';
import { SyncService } from '../../src/services/sync';

describe('SyncService Integration', () => {
  const context = createMockContext();
  let service: SyncService;

  beforeEach(async () => {
    await setupTestDatabase(context.storage);
    service = new SyncService(context.api, context.storage, context.events);
  });

  afterEach(async () => {
    await cleanupTestDatabase(context.storage);
  });

  it('should sync budget and emit events', async () => {
    // Test implementation
  });
});
```

## Test Data Characteristics

### Realistic Scenarios

- **Multiple budgets** with different currencies (USD, EUR)
- **Various account types** (checking, savings, credit card, asset tracking)
- **Complex transactions** including:
  - Simple transactions
  - Split transactions (multiple categories)
  - Transfer transactions (between accounts)
  - Imported transactions with reconciliation data
  - Flagged transactions
  - Cleared and uncleared states
- **Category goals** with different types (Target Balance, Target by Date)
- **Month-by-month data** showing budget evolution over time

### Edge Cases

- Overspending scenarios (negative category balances)
- Transfer pairs (linked transactions)
- Deleted/hidden entities
- Sync failures and retry scenarios
- Offline queue with pending operations
- Rate limiting and error responses

## Best Practices

1. **Use fixtures** instead of inline data
2. **Use helpers** to create test entities
3. **Reset mocks** between tests using `beforeEach`
4. **Test both success and failure** paths
5. **Verify mock calls** to ensure correct API usage
6. **Clean up resources** in `afterEach` hooks
7. **Use descriptive test names** that explain the scenario
8. **Group related tests** with `describe` blocks

## Coverage Goals

- **Unit tests**: 80%+ coverage of business logic
- **Integration tests**: Cover all service interactions
- **E2E tests**: Cover critical user workflows

## Troubleshooting

### Tests failing with "No mock response configured"

Make sure to configure mock responses before calling the service:

```typescript
api.mockResponse('getBudgets', mockData);
```

### Database tests failing

Ensure you're calling setup/cleanup in the right lifecycle hooks:

```typescript
beforeEach(async () => {
  await setupTestDatabase(storage);
});

afterEach(async () => {
  await cleanupTestDatabase(storage);
});
```

### Event tests not working

Check that you're setting up event listeners before emitting:

```typescript
const received = [];
events.on('my:event', (data) => received.push(data));
// Now emit or trigger the event
```
