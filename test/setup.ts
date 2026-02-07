/**
 * Global test setup
 * Runs before all tests to configure the test environment
 */

import { testEnvironment } from './fixtures/config/test-config';

// Set test environment variables
Object.entries(testEnvironment).forEach(([key, value]) => {
  process.env[key] = value;
});

// Suppress console output during tests (optional)
const SUPPRESS_LOGS = process.env.SUPPRESS_TEST_LOGS === 'true';

if (SUPPRESS_LOGS) {
  global.console = {
    ...console,
    log: jest.fn(),
    debug: jest.fn(),
    info: jest.fn(),
    warn: jest.fn(),
    // Keep error for debugging
    error: console.error,
  };
}

// Mock timers setup
beforeEach(() => {
  jest.clearAllMocks();
});

afterEach(() => {
  jest.restoreAllMocks();
});

// Clean up after all tests
afterAll(() => {
  // Clean up any resources
});

export {};
