/**
 * Test database seed data for sync state
 */

import type { SyncState, SyncLog } from '../../../src/types';

export const mockSyncStates: SyncState[] = [
  {
    budget_id: 'test-budget-1',
    last_sync: '2026-02-01T12:00:00.000Z',
    server_knowledge: 12350,
    status: 'success',
    error: null,
  },
  {
    budget_id: 'test-budget-2',
    last_sync: '2026-01-31T18:30:00.000Z',
    server_knowledge: 11000,
    status: 'success',
    error: null,
  },
];

export const mockSyncLogs: SyncLog[] = [
  {
    id: 'log-1',
    budget_id: 'test-budget-1',
    timestamp: '2026-02-01T12:00:00.000Z',
    status: 'success',
    records_synced: 45,
    duration_ms: 1234,
    error: null,
  },
  {
    id: 'log-2',
    budget_id: 'test-budget-1',
    timestamp: '2026-02-01T06:00:00.000Z',
    status: 'success',
    records_synced: 12,
    duration_ms: 567,
    error: null,
  },
  {
    id: 'log-3',
    budget_id: 'test-budget-1',
    timestamp: '2026-01-31T12:00:00.000Z',
    status: 'error',
    records_synced: 0,
    duration_ms: 2345,
    error: 'Rate limit exceeded',
  },
  {
    id: 'log-4',
    budget_id: 'test-budget-2',
    timestamp: '2026-01-31T18:30:00.000Z',
    status: 'success',
    records_synced: 23,
    duration_ms: 890,
    error: null,
  },
];

export const mockOfflineQueue = [
  {
    id: 'queue-1',
    budget_id: 'test-budget-1',
    operation: 'create_transaction',
    data: {
      account_id: 'account-checking-1',
      date: '2026-02-02',
      amount: -2500,
      payee_name: 'Coffee Shop',
      category_id: 'cat-dining-1',
      memo: 'Morning coffee',
    },
    timestamp: '2026-02-02T08:15:00.000Z',
    retry_count: 0,
    last_error: null,
  },
  {
    id: 'queue-2',
    budget_id: 'test-budget-1',
    operation: 'update_transaction',
    data: {
      transaction_id: 'txn-3',
      cleared: 'cleared',
    },
    timestamp: '2026-02-02T08:20:00.000Z',
    retry_count: 1,
    last_error: 'Network timeout',
  },
];
