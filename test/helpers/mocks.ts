/**
 * Mock implementations for testing
 */

import type { YnabApi } from '../../src/api/client';
import type { StorageAdapter } from '../../src/storage/adapter';
import type { EventBus } from '../../src/events/bus';

/**
 * Mock YNAB API client
 */
export class MockYnabApi implements Partial<YnabApi> {
  public callLog: Array<{ method: string; args: any[] }> = [];
  private mockResponses: Map<string, any> = new Map();

  mockResponse(method: string, response: any): void {
    this.mockResponses.set(method, response);
  }

  private logCall(method: string, ...args: any[]): any {
    this.callLog.push({ method, args });
    const response = this.mockResponses.get(method);
    if (response instanceof Error) {
      throw response;
    }
    return response;
  }

  async getBudgets() {
    return this.logCall('getBudgets');
  }

  async getBudget(budgetId: string) {
    return this.logCall('getBudget', budgetId);
  }

  async getAccounts(budgetId: string) {
    return this.logCall('getAccounts', budgetId);
  }

  async getAccount(budgetId: string, accountId: string) {
    return this.logCall('getAccount', budgetId, accountId);
  }

  async getCategories(budgetId: string) {
    return this.logCall('getCategories', budgetId);
  }

  async getTransactions(budgetId: string, sinceDate?: string) {
    return this.logCall('getTransactions', budgetId, sinceDate);
  }

  async getTransactionsByAccount(budgetId: string, accountId: string, sinceDate?: string) {
    return this.logCall('getTransactionsByAccount', budgetId, accountId, sinceDate);
  }

  async createTransaction(budgetId: string, transaction: any) {
    return this.logCall('createTransaction', budgetId, transaction);
  }

  async updateTransaction(budgetId: string, transactionId: string, transaction: any) {
    return this.logCall('updateTransaction', budgetId, transactionId, transaction);
  }

  async deleteTransaction(budgetId: string, transactionId: string) {
    return this.logCall('deleteTransaction', budgetId, transactionId);
  }

  async getPayees(budgetId: string) {
    return this.logCall('getPayees', budgetId);
  }

  async getMonthBudget(budgetId: string, month: string) {
    return this.logCall('getMonthBudget', budgetId, month);
  }

  reset(): void {
    this.callLog = [];
    this.mockResponses.clear();
  }

  wasCalledWith(method: string, ...args: any[]): boolean {
    return this.callLog.some(
      (call) =>
        call.method === method &&
        call.args.every((arg, i) => JSON.stringify(arg) === JSON.stringify(args[i]))
    );
  }
}

/**
 * Mock Storage Adapter
 */
export class MockStorageAdapter implements Partial<StorageAdapter> {
  private data: Map<string, any> = new Map();
  public callLog: Array<{ method: string; args: any[] }> = [];

  private logCall(method: string, ...args: any[]): void {
    this.callLog.push({ method, args });
  }

  async saveBudget(budget: any): Promise<void> {
    this.logCall('saveBudget', budget);
    this.data.set(`budget:${budget.id}`, budget);
  }

  async getBudget(budgetId: string): Promise<any> {
    this.logCall('getBudget', budgetId);
    return this.data.get(`budget:${budgetId}`);
  }

  async getAllBudgets(): Promise<any[]> {
    this.logCall('getAllBudgets');
    const budgets: any[] = [];
    this.data.forEach((value, key) => {
      if (key.startsWith('budget:')) {
        budgets.push(value);
      }
    });
    return budgets;
  }

  async saveAccount(budgetId: string, account: any): Promise<void> {
    this.logCall('saveAccount', budgetId, account);
    this.data.set(`account:${account.id}`, account);
  }

  async getAccount(accountId: string): Promise<any> {
    this.logCall('getAccount', accountId);
    return this.data.get(`account:${accountId}`);
  }

  async getAccountsByBudget(budgetId: string): Promise<any[]> {
    this.logCall('getAccountsByBudget', budgetId);
    const accounts: any[] = [];
    this.data.forEach((value, key) => {
      if (key.startsWith('account:') && value.budget_id === budgetId) {
        accounts.push(value);
      }
    });
    return accounts;
  }

  async saveTransaction(budgetId: string, transaction: any): Promise<void> {
    this.logCall('saveTransaction', budgetId, transaction);
    this.data.set(`transaction:${transaction.id}`, transaction);
  }

  async getTransaction(transactionId: string): Promise<any> {
    this.logCall('getTransaction', transactionId);
    return this.data.get(`transaction:${transactionId}`);
  }

  async getTransactionsByAccount(accountId: string): Promise<any[]> {
    this.logCall('getTransactionsByAccount', accountId);
    const transactions: any[] = [];
    this.data.forEach((value, key) => {
      if (key.startsWith('transaction:') && value.account_id === accountId) {
        transactions.push(value);
      }
    });
    return transactions;
  }

  async saveSyncState(budgetId: string, syncState: any): Promise<void> {
    this.logCall('saveSyncState', budgetId, syncState);
    this.data.set(`sync:${budgetId}`, syncState);
  }

  async getSyncState(budgetId: string): Promise<any> {
    this.logCall('getSyncState', budgetId);
    return this.data.get(`sync:${budgetId}`);
  }

  async clear(): Promise<void> {
    this.logCall('clear');
    this.data.clear();
  }

  reset(): void {
    this.data.clear();
    this.callLog = [];
  }
}

/**
 * Mock Event Bus
 */
export class MockEventBus implements Partial<EventBus> {
  private listeners: Map<string, Array<(...args: any[]) => void>> = new Map();
  public emittedEvents: Array<{ event: string; data: any }> = [];

  on(event: string, handler: (...args: any[]) => void): void {
    if (!this.listeners.has(event)) {
      this.listeners.set(event, []);
    }
    this.listeners.get(event)!.push(handler);
  }

  off(event: string, handler: (...args: any[]) => void): void {
    const handlers = this.listeners.get(event);
    if (handlers) {
      const index = handlers.indexOf(handler);
      if (index > -1) {
        handlers.splice(index, 1);
      }
    }
  }

  emit(event: string, data: any): void {
    this.emittedEvents.push({ event, data });
    const handlers = this.listeners.get(event);
    if (handlers) {
      handlers.forEach((handler) => handler(data));
    }
  }

  wasEmitted(event: string, data?: any): boolean {
    return this.emittedEvents.some(
      (e) => e.event === event && (data === undefined || JSON.stringify(e.data) === JSON.stringify(data))
    );
  }

  getEmittedEvents(event: string): any[] {
    return this.emittedEvents.filter((e) => e.event === event).map((e) => e.data);
  }

  reset(): void {
    this.listeners.clear();
    this.emittedEvents = [];
  }
}

/**
 * Mock Gateway for testing multi-platform features
 */
export class MockGateway {
  public sentMessages: Array<{ channel: string; message: any }> = [];
  private mockResponses: Map<string, any> = new Map();

  mockResponse(channel: string, response: any): void {
    this.mockResponses.set(channel, response);
  }

  async send(channel: string, message: any): Promise<void> {
    this.sentMessages.push({ channel, message });
  }

  async request(channel: string, data: any): Promise<any> {
    this.sentMessages.push({ channel, message: data });
    return this.mockResponses.get(channel);
  }

  wasSentTo(channel: string, message?: any): boolean {
    return this.sentMessages.some(
      (m) => m.channel === channel && (message === undefined || JSON.stringify(m.message) === JSON.stringify(message))
    );
  }

  reset(): void {
    this.sentMessages = [];
    this.mockResponses.clear();
  }
}

/**
 * Create a complete mock context for testing
 */
export function createMockContext() {
  return {
    api: new MockYnabApi(),
    storage: new MockStorageAdapter(),
    events: new MockEventBus(),
    gateway: new MockGateway(),
  };
}

/**
 * Reset all mocks in a context
 */
export function resetMockContext(context: ReturnType<typeof createMockContext>): void {
  context.api.reset();
  context.storage.reset();
  context.events.reset();
  context.gateway.reset();
}
