/**
 * Test configuration fixtures
 */

export const testConfig = {
  api: {
    baseUrl: 'https://api.ynab.com/v1',
    timeout: 5000,
    retryAttempts: 3,
    retryDelay: 1000,
  },
  sync: {
    interval: 300000, // 5 minutes
    batchSize: 100,
    concurrency: 3,
  },
  storage: {
    path: ':memory:', // In-memory SQLite for tests
    ttl: 3600,
  },
  cache: {
    enabled: true,
    ttl: 300,
    maxSize: 1000,
  },
};

export const testApiKey = 'test-api-key-12345';

export const testEnvironment = {
  YNAB_API_KEY: testApiKey,
  VIA_ENV: 'test',
  NODE_ENV: 'test',
};

export const mockYnabConfig = {
  budgets: {
    active: ['test-budget-1'],
    default: 'test-budget-1',
  },
  sync: {
    enabled: true,
    interval: 300000,
    autoSync: true,
  },
  notifications: {
    enabled: false,
    overspending: false,
    goals: false,
  },
  privacy: {
    encrypt: true,
    local_only: true,
  },
};
