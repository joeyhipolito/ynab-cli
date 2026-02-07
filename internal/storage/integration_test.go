package storage

import (
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

// ============================================================================
// Database Operations Under Load
// ============================================================================

// TestYNABStoreConcurrentWrites tests concurrent write operations to the store.
func TestYNABStoreConcurrentWrites(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ynab_concurrent.db")

	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	// Create a budget first
	budget := Budget{
		ID:   "test-budget",
		Name: "Test Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}

	if err := store.CreateBudget(budget); err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	// Create an account for transactions
	account := Account{
		ID:       "test-account",
		BudgetID: "test-budget",
		Name:     "Checking",
		Type:     "checking",
		Balance:  1000000,
	}

	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	// Create a category for transactions
	category := Category{
		ID:       "test-category",
		BudgetID: "test-budget",
		Name:     "Groceries",
	}

	if err := store.CreateCategory(category); err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	// Concurrently create 100 transactions
	var wg sync.WaitGroup
	errors := make(chan error, 100)

	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			tx := Transaction{
				ID:         fmt.Sprintf("tx-%d", index),
				BudgetID:   "test-budget",
				AccountID:  "test-account",
				CategoryID: "test-category",
				Date:       "2026-02-02",
				Amount:     int64(-5000 * index),
				Memo:       fmt.Sprintf("Transaction %d", index),
			}

			if err := store.CreateTransaction(tx); err != nil {
				errors <- err
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Errorf("Concurrent write error: %v", err)
	}

	// Verify all transactions were created
	transactions, err := store.ListTransactionsByAccount("test-account")
	if err != nil {
		t.Fatalf("ListTransactionsByAccount failed: %v", err)
	}

	if len(transactions) != 100 {
		t.Errorf("Expected 100 transactions, got %d", len(transactions))
	}
}

// TestYNABStoreTransactionRollback tests transaction rollback on error.
func TestYNABStoreTransactionRollback(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ynab_rollback.db")

	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	budget := Budget{
		ID:   "test-budget",
		Name: "Test Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}

	if err := store.CreateBudget(budget); err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	// Begin transaction
	tx, err := store.db.Begin()
	if err != nil {
		t.Fatalf("Begin transaction failed: %v", err)
	}

	// Create account
	_, err = tx.Exec(`
		INSERT INTO accounts (id, budget_id, name, type, balance)
		VALUES (?, ?, ?, ?, ?)
	`, "test-account", "test-budget", "Checking", "checking", 100000)
	if err != nil {
		t.Fatalf("Insert account failed: %v", err)
	}

	// Attempt to create account with duplicate ID (should fail)
	_, err = tx.Exec(`
		INSERT INTO accounts (id, budget_id, name, type, balance)
		VALUES (?, ?, ?, ?, ?)
	`, "test-account", "test-budget", "Savings", "savings", 200000)

	if err == nil {
		t.Fatal("Expected duplicate key error, got nil")
	}

	// Rollback
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify account was not created
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM accounts WHERE id = ?", "test-account").Scan(&count)
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}

	if count != 0 {
		t.Errorf("Expected 0 accounts after rollback, got %d", count)
	}
}

// TestYNABStoreLargeDatasetPerformance tests performance with large datasets.
func TestYNABStoreLargeDatasetPerformance(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping performance test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ynab_performance.db")

	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	// Setup
	budget := Budget{
		ID:   "perf-budget",
		Name: "Performance Test Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}

	if err := store.CreateBudget(budget); err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	account := Account{
		ID:       "perf-account",
		BudgetID: "perf-budget",
		Name:     "Test Account",
		Type:     "checking",
		Balance:  1000000,
	}

	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	category := Category{
		ID:       "perf-category",
		BudgetID: "perf-budget",
		Name:     "Test Category",
	}

	if err := store.CreateCategory(category); err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	// Insert 10,000 transactions
	start := time.Now()
	numTransactions := 10000

	for i := 0; i < numTransactions; i++ {
		tx := Transaction{
			ID:         fmt.Sprintf("perf-tx-%d", i),
			BudgetID:   "perf-budget",
			AccountID:  "perf-account",
			CategoryID: "perf-category",
			Date:       "2026-02-02",
			Amount:     int64(-5000 * i),
			Memo:       fmt.Sprintf("Performance Transaction %d", i),
		}

		if err := store.CreateTransaction(tx); err != nil {
			t.Fatalf("CreateTransaction failed: %v", err)
		}
	}

	insertDuration := time.Since(start)
	t.Logf("Inserted %d transactions in %v (%.2f tx/s)",
		numTransactions, insertDuration, float64(numTransactions)/insertDuration.Seconds())

	// Query performance
	start = time.Now()
	transactions, err := store.ListTransactionsByAccount("perf-account")
	if err != nil {
		t.Fatalf("ListTransactionsByAccount failed: %v", err)
	}
	queryDuration := time.Since(start)

	t.Logf("Queried %d transactions in %v", len(transactions), queryDuration)

	if len(transactions) != numTransactions {
		t.Errorf("Expected %d transactions, got %d", numTransactions, len(transactions))
	}

	// Performance assertions (adjust thresholds as needed)
	maxInsertDuration := 10 * time.Second
	if insertDuration > maxInsertDuration {
		t.Errorf("Insert took too long: %v (max %v)", insertDuration, maxInsertDuration)
	}

	maxQueryDuration := 1 * time.Second
	if queryDuration > maxQueryDuration {
		t.Errorf("Query took too long: %v (max %v)", queryDuration, maxQueryDuration)
	}
}

// ============================================================================
// Foreign Key Constraints
// ============================================================================

// TestYNABStoreCascadeDelete tests cascade delete behavior.
func TestYNABStoreCascadeDelete(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ynab_cascade.db")

	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	// Create budget with accounts and transactions
	budget := Budget{
		ID:   "cascade-budget",
		Name: "Cascade Test Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}

	if err := store.CreateBudget(budget); err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	account := Account{
		ID:       "cascade-account",
		BudgetID: "cascade-budget",
		Name:     "Test Account",
		Type:     "checking",
		Balance:  100000,
	}

	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	category := Category{
		ID:       "cascade-category",
		BudgetID: "cascade-budget",
		Name:     "Test Category",
	}

	if err := store.CreateCategory(category); err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	tx := Transaction{
		ID:         "cascade-tx",
		BudgetID:   "cascade-budget",
		AccountID:  "cascade-account",
		CategoryID: "cascade-category",
		Date:       "2026-02-02",
		Amount:     -5000,
		Memo:       "Test Transaction",
	}

	if err := store.CreateTransaction(tx); err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	// Delete budget (should cascade delete accounts and transactions)
	if err := store.DeleteBudget("cascade-budget"); err != nil {
		t.Fatalf("DeleteBudget failed: %v", err)
	}

	// Verify account was deleted
	var accountCount int
	err = store.db.QueryRow("SELECT COUNT(*) FROM accounts WHERE id = ?", "cascade-account").Scan(&accountCount)
	if err != nil {
		t.Fatalf("Query accounts failed: %v", err)
	}
	if accountCount != 0 {
		t.Errorf("Expected 0 accounts after cascade delete, got %d", accountCount)
	}

	// Verify transaction was deleted
	var txCount int
	err = store.db.QueryRow("SELECT COUNT(*) FROM transactions WHERE id = ?", "cascade-tx").Scan(&txCount)
	if err != nil {
		t.Fatalf("Query transactions failed: %v", err)
	}
	if txCount != 0 {
		t.Errorf("Expected 0 transactions after cascade delete, got %d", txCount)
	}
}

// TestYNABStoreForeignKeyViolation tests foreign key constraint enforcement.
func TestYNABStoreForeignKeyViolation(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ynab_fk.db")

	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	// Attempt to create account without budget (should fail)
	account := Account{
		ID:       "orphan-account",
		BudgetID: "nonexistent-budget",
		Name:     "Orphan Account",
		Type:     "checking",
		Balance:  100000,
	}

	err = store.CreateAccount(account)
	if err == nil {
		t.Fatal("Expected foreign key violation, got nil")
	}

	t.Logf("Foreign key violation correctly caught: %v", err)
}

// ============================================================================
// Query Performance & Indexing
// ============================================================================

// TestYNABStoreIndexUsage tests that indexes are properly used.
func TestYNABStoreIndexUsage(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ynab_index.db")

	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	// Check that indexes exist
	rows, err := store.db.Query(`
		SELECT name FROM sqlite_master
		WHERE type = 'index' AND tbl_name = 'transactions'
	`)
	if err != nil {
		t.Fatalf("Query indexes failed: %v", err)
	}
	defer rows.Close()

	indexes := []string{}
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		indexes = append(indexes, name)
	}

	// Verify key indexes exist
	expectedIndexes := []string{"idx_transactions_account", "idx_transactions_date", "idx_transactions_budget"}
	for _, expected := range expectedIndexes {
		found := false
		for _, actual := range indexes {
			if actual == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected index %s not found", expected)
		}
	}

	t.Logf("Found indexes: %v", indexes)
}

// TestYNABStoreComplexQueryOptimization tests complex query performance.
func TestYNABStoreComplexQueryOptimization(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping optimization test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "ynab_query_opt.db")

	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	// Setup test data
	budget := Budget{
		ID:   "opt-budget",
		Name: "Optimization Test",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}

	if err := store.CreateBudget(budget); err != nil {
		t.Fatalf("CreateBudget failed: %v", err)
	}

	account := Account{
		ID:       "opt-account",
		BudgetID: "opt-budget",
		Name:     "Test Account",
		Type:     "checking",
		Balance:  1000000,
	}

	if err := store.CreateAccount(account); err != nil {
		t.Fatalf("CreateAccount failed: %v", err)
	}

	category := Category{
		ID:       "opt-category",
		BudgetID: "opt-budget",
		Name:     "Test Category",
	}

	if err := store.CreateCategory(category); err != nil {
		t.Fatalf("CreateCategory failed: %v", err)
	}

	// Insert test transactions
	for i := 0; i < 1000; i++ {
		tx := Transaction{
			ID:         fmt.Sprintf("opt-tx-%d", i),
			BudgetID:   "opt-budget",
			AccountID:  "opt-account",
			CategoryID: "opt-category",
			Date:       fmt.Sprintf("2026-01-%02d", (i%28)+1),
			Amount:     int64(-5000 * i),
		}

		if err := store.CreateTransaction(tx); err != nil {
			t.Fatalf("CreateTransaction failed: %v", err)
		}
	}

	// Test complex query with date range
	start := time.Now()
	transactions, err := store.ListTransactionsByDateRange("opt-budget", "2026-01-01", "2026-01-31")
	if err != nil {
		t.Fatalf("ListTransactionsByDateRange failed: %v", err)
	}
	duration := time.Since(start)

	t.Logf("Complex query returned %d transactions in %v", len(transactions), duration)

	// Performance assertion
	maxDuration := 500 * time.Millisecond
	if duration > maxDuration {
		t.Errorf("Query took too long: %v (max %v)", duration, maxDuration)
	}
}

// ============================================================================
// Helper Functions
// ============================================================================
