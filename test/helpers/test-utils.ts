/**
 * Test helper utilities
 */

import { readFileSync } from 'fs';
import { join } from 'path';
import type { Budget, Account, Transaction, Category } from '../../src/types';

/**
 * Load JSON fixture from file
 */
export function loadFixture<T = any>(fixturePath: string): T {
  const fullPath = join(__dirname, '..', 'fixtures', fixturePath);
  const content = readFileSync(fullPath, 'utf-8');
  return JSON.parse(content);
}

/**
 * Load API response fixture
 */
export function loadApiFixture(name: string) {
  return loadFixture(`api/${name}.json`);
}

/**
 * Create mock API response
 */
export function mockApiResponse<T>(data: T, serverKnowledge?: number) {
  return {
    data,
    ...(serverKnowledge && { server_knowledge: serverKnowledge }),
  };
}

/**
 * Create mock error response
 */
export function mockErrorResponse(message: string, status = 400) {
  return {
    error: {
      id: 'test-error-id',
      name: 'test_error',
      detail: message,
    },
  };
}

/**
 * Wait for async operations
 */
export function wait(ms: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

/**
 * Create a mock date that can be controlled in tests
 */
export class MockDate {
  private currentDate: Date;

  constructor(initialDate?: string | Date) {
    this.currentDate = initialDate ? new Date(initialDate) : new Date();
  }

  now(): Date {
    return new Date(this.currentDate);
  }

  advance(ms: number): void {
    this.currentDate = new Date(this.currentDate.getTime() + ms);
  }

  setDate(date: string | Date): void {
    this.currentDate = new Date(date);
  }
}

/**
 * Generate test budget
 */
export function createTestBudget(overrides: Partial<Budget> = {}): Budget {
  return {
    id: 'test-budget-id',
    name: 'Test Budget',
    last_modified_on: '2026-02-01T00:00:00.000Z',
    first_month: '2025-01-01',
    last_month: '2026-12-01',
    date_format: { format: 'MM/DD/YYYY' },
    currency_format: {
      iso_code: 'USD',
      example_format: '123,456.78',
      decimal_digits: 2,
      decimal_separator: '.',
      symbol_first: true,
      group_separator: ',',
      currency_symbol: '$',
      display_symbol: true,
    },
    ...overrides,
  };
}

/**
 * Generate test account
 */
export function createTestAccount(overrides: Partial<Account> = {}): Account {
  return {
    id: 'test-account-id',
    name: 'Test Account',
    type: 'checking',
    on_budget: true,
    closed: false,
    note: null,
    balance: 0,
    cleared_balance: 0,
    uncleared_balance: 0,
    transfer_payee_id: 'test-transfer-payee',
    direct_import_linked: false,
    direct_import_in_error: false,
    deleted: false,
    ...overrides,
  };
}

/**
 * Generate test transaction
 */
export function createTestTransaction(overrides: Partial<Transaction> = {}): Transaction {
  return {
    id: 'test-txn-id',
    date: '2026-02-01',
    amount: -1000,
    memo: null,
    cleared: 'uncleared',
    approved: false,
    flag_color: null,
    account_id: 'test-account-id',
    payee_id: 'test-payee-id',
    category_id: 'test-category-id',
    transfer_account_id: null,
    transfer_transaction_id: null,
    matched_transaction_id: null,
    import_id: null,
    import_payee_name: null,
    import_payee_name_original: null,
    deleted: false,
    subtransactions: [],
    ...overrides,
  };
}

/**
 * Generate test category
 */
export function createTestCategory(overrides: Partial<Category> = {}): Category {
  return {
    id: 'test-category-id',
    category_group_id: 'test-group-id',
    name: 'Test Category',
    hidden: false,
    note: null,
    budgeted: 0,
    activity: 0,
    balance: 0,
    goal_type: null,
    deleted: false,
    ...overrides,
  };
}

/**
 * Assert deep equality with helpful error messages
 */
export function assertDeepEqual<T>(actual: T, expected: T, message?: string): void {
  const actualStr = JSON.stringify(actual, null, 2);
  const expectedStr = JSON.stringify(expected, null, 2);

  if (actualStr !== expectedStr) {
    const errorMsg = message
      ? `${message}\nExpected:\n${expectedStr}\nActual:\n${actualStr}`
      : `Expected:\n${expectedStr}\nActual:\n${actualStr}`;
    throw new Error(errorMsg);
  }
}

/**
 * Create a mock HTTP client for testing
 */
export class MockHttpClient {
  private responses: Map<string, any> = new Map();
  private callHistory: Array<{ url: string; method: string; data?: any }> = [];

  mockResponse(url: string, response: any): void {
    this.responses.set(url, response);
  }

  async request(method: string, url: string, data?: any): Promise<any> {
    this.callHistory.push({ url, method, data });

    const response = this.responses.get(url);
    if (!response) {
      throw new Error(`No mock response configured for ${method} ${url}`);
    }

    // Simulate network delay
    await wait(10);

    return response;
  }

  getCallHistory() {
    return this.callHistory;
  }

  reset(): void {
    this.responses.clear();
    this.callHistory = [];
  }
}

/**
 * Create a mock storage for testing
 */
export class MockStorage {
  private data: Map<string, any> = new Map();

  async get(key: string): Promise<any> {
    return this.data.get(key);
  }

  async set(key: string, value: any): Promise<void> {
    this.data.set(key, value);
  }

  async delete(key: string): Promise<void> {
    this.data.delete(key);
  }

  async clear(): Promise<void> {
    this.data.clear();
  }

  async has(key: string): Promise<boolean> {
    return this.data.has(key);
  }

  async keys(): Promise<string[]> {
    return Array.from(this.data.keys());
  }
}

/**
 * Spy on function calls
 */
export class FunctionSpy<T extends (...args: any[]) => any> {
  private calls: Array<{ args: Parameters<T>; result: ReturnType<T> | Error }> = [];
  private implementation?: T;

  constructor(implementation?: T) {
    this.implementation = implementation;
  }

  fn = ((...args: Parameters<T>): ReturnType<T> => {
    try {
      const result = this.implementation ? this.implementation(...args) : undefined;
      this.calls.push({ args, result });
      return result;
    } catch (error) {
      this.calls.push({ args, result: error as Error });
      throw error;
    }
  }) as T;

  getCalls() {
    return this.calls;
  }

  getCallCount() {
    return this.calls.length;
  }

  wasCalledWith(...args: Parameters<T>) {
    return this.calls.some((call) =>
      call.args.every((arg, i) => JSON.stringify(arg) === JSON.stringify(args[i]))
    );
  }

  reset() {
    this.calls = [];
  }
}

/**
 * Test database setup helper
 */
export async function setupTestDatabase(storage: any) {
  // Initialize test schema
  await storage.exec(`
    CREATE TABLE IF NOT EXISTS budgets (
      id TEXT PRIMARY KEY,
      name TEXT NOT NULL,
      data TEXT NOT NULL,
      last_modified TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS accounts (
      id TEXT PRIMARY KEY,
      budget_id TEXT NOT NULL,
      name TEXT NOT NULL,
      data TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS transactions (
      id TEXT PRIMARY KEY,
      budget_id TEXT NOT NULL,
      account_id TEXT NOT NULL,
      date TEXT NOT NULL,
      amount INTEGER NOT NULL,
      data TEXT NOT NULL
    );

    CREATE TABLE IF NOT EXISTS sync_state (
      budget_id TEXT PRIMARY KEY,
      last_sync TEXT NOT NULL,
      server_knowledge INTEGER NOT NULL,
      status TEXT NOT NULL
    );
  `);
}

/**
 * Clean up test database
 */
export async function cleanupTestDatabase(storage: any) {
  await storage.exec(`
    DROP TABLE IF EXISTS budgets;
    DROP TABLE IF EXISTS accounts;
    DROP TABLE IF EXISTS transactions;
    DROP TABLE IF EXISTS sync_state;
  `);
}
