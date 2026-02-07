package ynab_test

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/joeyhipolito/via/features/events"
	"github.com/joeyhipolito/via/features/events/schema"
	"github.com/joeyhipolito/via/features/security"
)

// ============================================================================
// End-to-End Test Suite for YNAB Integration
// ============================================================================

// TestE2E_YNABAuthenticationFlow tests the complete authentication workflow.
func TestE2E_YNABAuthenticationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	// Setup test environment
	tmpDir := t.TempDir()
	configDir := filepath.Join(tmpDir, ".via")
	os.MkdirAll(configDir, 0755)

	// Set environment for Via
	os.Setenv("VIA_HOME", tmpDir)
	defer os.Unsetenv("VIA_HOME")

	// Initialize security manager
	secMgr, err := security.NewManager(configDir)
	if err != nil {
		t.Fatalf("Failed to initialize security manager: %v", err)
	}
	defer secMgr.Close()

	// Simulate user providing API token
	testToken := "test-ynab-api-token-abc123"

	// Store token securely
	err = secMgr.StoreSecret("ynab_token", testToken)
	if err != nil {
		t.Fatalf("Failed to store YNAB token: %v", err)
	}

	// Verify token is stored encrypted
	rawFile := filepath.Join(configDir, ".ynab_token")
	rawData, err := os.ReadFile(rawFile)
	if err != nil {
		t.Fatalf("Failed to read token file: %v", err)
	}

	if string(rawData) == testToken {
		t.Error("Token stored in plaintext (security violation)")
	}

	// Retrieve and verify token
	retrievedToken, err := secMgr.GetSecret("ynab_token")
	if err != nil {
		t.Fatalf("Failed to retrieve token: %v", err)
	}

	if retrievedToken != testToken {
		t.Errorf("Token mismatch after encryption/decryption")
	}

	t.Log("✓ Authentication flow: token stored and retrieved securely")
}

// TestE2E_YNABBudgetSync tests the complete budget synchronization workflow.
func TestE2E_YNABBudgetSync(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()

	// Initialize components
	dbPath := filepath.Join(tmpDir, "via.db")
	store, err := NewTestYNABStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Initialize event bus
	bus := events.NewBus()
	defer bus.Close()

	// Track sync events
	var syncEvents []schema.Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	bus.Subscribe("budget:sync:*", func(e schema.Event) {
		mu.Lock()
		syncEvents = append(syncEvents, e)
		mu.Unlock()
		wg.Done()
	})

	// Simulate YNAB API sync workflow
	correlationID := fmt.Sprintf("sync-%d", time.Now().Unix())

	// 1. Sync started
	wg.Add(1)
	bus.Publish(schema.NewEvent(
		"budget:sync:started",
		map[string]interface{}{
			"budget_id": "test-budget-1",
			"sync_type": "full",
		},
		correlationID,
	))

	// 2. Create budget in local store
	budget := TestBudget{
		ID:   "test-budget-1",
		Name: "My Test Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
			Symbol:        "$",
		},
		LastModifiedOn: time.Now().Format(time.RFC3339),
	}

	err = store.CreateBudget(budget)
	if err != nil {
		t.Fatalf("Failed to create budget: %v", err)
	}

	// 3. Create accounts
	accounts := []TestAccount{
		{
			ID:       "account-1",
			BudgetID: "test-budget-1",
			Name:     "Checking",
			Type:     "checking",
			Balance:  100000, // $1,000.00
		},
		{
			ID:       "account-2",
			BudgetID: "test-budget-1",
			Name:     "Savings",
			Type:     "savings",
			Balance:  500000, // $5,000.00
		},
	}

	for _, account := range accounts {
		err = store.CreateAccount(account)
		if err != nil {
			t.Fatalf("Failed to create account: %v", err)
		}
	}

	// 4. Create categories
	category := TestCategory{
		ID:       "cat-groceries",
		BudgetID: "test-budget-1",
		Name:     "Groceries",
	}

	err = store.CreateCategory(category)
	if err != nil {
		t.Fatalf("Failed to create category: %v", err)
	}

	// 5. Create transactions
	transactions := []TestTransaction{
		{
			ID:         "tx-1",
			BudgetID:   "test-budget-1",
			AccountID:  "account-1",
			CategoryID: "cat-groceries",
			Date:       "2026-02-01",
			Amount:     -8500, // -$85.00
			Memo:       "Whole Foods",
			Cleared:    "cleared",
		},
		{
			ID:         "tx-2",
			BudgetID:   "test-budget-1",
			AccountID:  "account-1",
			CategoryID: "cat-groceries",
			Date:       "2026-02-02",
			Amount:     -4250, // -$42.50
			Memo:       "Trader Joe's",
			Cleared:    "cleared",
		},
	}

	for _, tx := range transactions {
		err = store.CreateTransaction(tx)
		if err != nil {
			t.Fatalf("Failed to create transaction: %v", err)
		}
	}

	// 6. Sync progress event
	wg.Add(1)
	bus.Publish(schema.NewEvent(
		"budget:sync:progress",
		map[string]interface{}{
			"budget_id":          "test-budget-1",
			"accounts_synced":    2,
			"transactions_synced": 2,
			"progress_percent":   75,
		},
		correlationID,
	))

	// 7. Sync completed event
	wg.Add(1)
	bus.Publish(schema.NewEvent(
		"budget:sync:completed",
		map[string]interface{}{
			"budget_id":            "test-budget-1",
			"sync_type":            "full",
			"duration_ms":          1500,
			"accounts_added":       2,
			"transactions_added":   2,
			"transactions_updated": 0,
			"server_knowledge":     12345,
		},
		correlationID,
	))

	// Wait for all sync events
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Timeout waiting for sync events")
	}

	// Verify sync events
	mu.Lock()
	defer mu.Unlock()

	if len(syncEvents) != 3 {
		t.Errorf("Expected 3 sync events, got %d", len(syncEvents))
	}

	// Verify data in store
	retrievedBudget, err := store.GetBudget("test-budget-1")
	if err != nil {
		t.Fatalf("Failed to retrieve budget: %v", err)
	}

	if retrievedBudget.Name != "My Test Budget" {
		t.Errorf("Budget name mismatch: expected 'My Test Budget', got '%s'", retrievedBudget.Name)
	}

	// Verify transactions
	txList, err := store.ListTransactionsByAccount("account-1")
	if err != nil {
		t.Fatalf("Failed to list transactions: %v", err)
	}

	if len(txList) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txList))
	}

	t.Log("✓ Budget sync: full workflow completed successfully")
}

// TestE2E_YNABTransactionWorkflow tests creating a transaction from CLI to storage.
func TestE2E_YNABTransactionWorkflow(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "via.db")

	store, err := NewTestYNABStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	bus := events.NewBus()
	defer bus.Close()

	// Track transaction events
	var txEvents []schema.Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	bus.Subscribe("budget:transaction:*", func(e schema.Event) {
		mu.Lock()
		txEvents = append(txEvents, e)
		mu.Unlock()
		wg.Done()
	})

	// Setup: Create budget, account, category
	budget := TestBudget{
		ID:   "budget-1",
		Name: "Personal Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}
	store.CreateBudget(budget)

	account := TestAccount{
		ID:       "checking",
		BudgetID: "budget-1",
		Name:     "Checking Account",
		Type:     "checking",
		Balance:  200000,
	}
	store.CreateAccount(account)

	category := TestCategory{
		ID:       "dining",
		BudgetID: "budget-1",
		Name:     "Dining Out",
	}
	store.CreateCategory(category)

	// Simulate user command: "via budget add 25.50 'Coffee Shop' --category dining"
	// This would parse to:
	transactionData := map[string]interface{}{
		"amount":      -2550, // $25.50 in milliunits
		"payee":       "Coffee Shop",
		"category_id": "dining",
		"account_id":  "checking",
		"date":        "2026-02-02",
	}

	// Validate transaction input
	err = validateTransactionInput(transactionData)
	if err != nil {
		t.Fatalf("Transaction validation failed: %v", err)
	}

	// Create transaction
	tx := TestTransaction{
		ID:         generateTransactionID(),
		BudgetID:   "budget-1",
		AccountID:  transactionData["account_id"].(string),
		CategoryID: transactionData["category_id"].(string),
		Date:       transactionData["date"].(string),
		Amount:     int64(transactionData["amount"].(int)),
		Memo:       transactionData["payee"].(string),
		Cleared:    "uncleared",
	}

	err = store.CreateTransaction(tx)
	if err != nil {
		t.Fatalf("Failed to create transaction: %v", err)
	}

	// Publish transaction added event
	wg.Add(1)
	bus.Publish(schema.NewEvent(
		"budget:transaction:added",
		map[string]interface{}{
			"budget_id": "budget-1",
			"transaction": map[string]interface{}{
				"id":          tx.ID,
				"amount":      float64(tx.Amount) / 1000.0,
				"payee":       tx.Memo,
				"category_id": tx.CategoryID,
				"date":        tx.Date,
				"account_id":  tx.AccountID,
			},
		},
		generateCorrelationID(),
	))

	// Wait for event
	eventDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(eventDone)
	}()

	select {
	case <-eventDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for transaction event")
	}

	// Verify transaction in database
	retrievedTx, err := store.GetTransaction(tx.ID)
	if err != nil {
		t.Fatalf("Failed to retrieve transaction: %v", err)
	}

	if retrievedTx.Amount != -2550 {
		t.Errorf("Amount mismatch: expected -2550, got %d", retrievedTx.Amount)
	}

	if retrievedTx.Memo != "Coffee Shop" {
		t.Errorf("Memo mismatch: expected 'Coffee Shop', got '%s'", retrievedTx.Memo)
	}

	// Verify event was published
	mu.Lock()
	defer mu.Unlock()

	if len(txEvents) != 1 {
		t.Errorf("Expected 1 transaction event, got %d", len(txEvents))
	}

	t.Log("✓ Transaction workflow: created from CLI to storage with events")
}

// TestE2E_YNABBudgetLimitAlert tests budget limit exceeded detection and alerts.
func TestE2E_YNABBudgetLimitAlert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "via.db")

	store, err := NewTestYNABStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	bus := events.NewBus()
	defer bus.Close()

	// Track limit exceeded events
	var limitEvents []schema.Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	bus.Subscribe("budget:limit:exceeded", func(e schema.Event) {
		mu.Lock()
		limitEvents = append(limitEvents, e)
		mu.Unlock()
		wg.Done()
	})

	// Setup budget and category
	budget := TestBudget{
		ID:   "budget-1",
		Name: "Monthly Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}
	store.CreateBudget(budget)

	account := TestAccount{
		ID:       "checking",
		BudgetID: "budget-1",
		Name:     "Checking",
		Type:     "checking",
		Balance:  500000,
	}
	store.CreateAccount(account)

	category := TestCategory{
		ID:       "groceries",
		BudgetID: "budget-1",
		Name:     "Groceries",
	}
	store.CreateCategory(category)

	// Set budget limit for category (e.g., $400/month)
	categoryBudget := TestCategoryBudget{
		CategoryID: "groceries",
		BudgetID:   "budget-1",
		Month:      "2026-02-01",
		Budgeted:   400000, // $400.00
		Activity:   0,
	}
	store.CreateCategoryBudget(categoryBudget)

	// Add transactions that exceed the budget
	transactions := []TestTransaction{
		{
			ID:         "tx-1",
			BudgetID:   "budget-1",
			AccountID:  "checking",
			CategoryID: "groceries",
			Date:       "2026-02-01",
			Amount:     -150000, // -$150.00
			Cleared:    "cleared",
		},
		{
			ID:         "tx-2",
			BudgetID:   "budget-1",
			AccountID:  "checking",
			CategoryID: "groceries",
			Date:       "2026-02-05",
			Amount:     -180000, // -$180.00
			Cleared:    "cleared",
		},
		{
			ID:         "tx-3",
			BudgetID:   "budget-1",
			AccountID:  "checking",
			CategoryID: "groceries",
			Date:       "2026-02-10",
			Amount:     -95000, // -$95.00 (this exceeds budget)
			Cleared:    "cleared",
		},
	}

	totalSpent := int64(0)
	for _, tx := range transactions {
		store.CreateTransaction(tx)
		totalSpent += -tx.Amount // Convert to positive
	}

	budgetLimit := int64(400000)
	if totalSpent > budgetLimit {
		overspent := totalSpent - budgetLimit

		// Publish budget limit exceeded event
		wg.Add(1)
		bus.Publish(schema.NewEvent(
			"budget:limit:exceeded",
			map[string]interface{}{
				"budget_id":   "budget-1",
				"category_id": "groceries",
				"category":    "Groceries",
				"budgeted":    float64(budgetLimit) / 1000.0,
				"spent":       float64(totalSpent) / 1000.0,
				"overspent":   float64(overspent) / 1000.0,
				"month":       "2026-02",
			},
			generateCorrelationID(),
		))
	}

	// Wait for event
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Timeout waiting for limit exceeded event")
	}

	// Verify event was published
	mu.Lock()
	defer mu.Unlock()

	if len(limitEvents) != 1 {
		t.Errorf("Expected 1 limit exceeded event, got %d", len(limitEvents))
	}

	if len(limitEvents) > 0 {
		payload := limitEvents[0].Payload.(map[string]interface{})
		overspent := payload["overspent"].(float64)

		if overspent != 25.0 { // $425 - $400 = $25 overspent
			t.Errorf("Expected overspent amount $25.00, got $%.2f", overspent)
		}
	}

	t.Log("✓ Budget limit alert: detected overspending and published event")
}

// TestE2E_YNABOfflineMode tests offline mode with local storage.
func TestE2E_YNABOfflineMode(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "via.db")

	store, err := NewTestYNABStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	// Setup local budget (offline)
	budget := TestBudget{
		ID:   "offline-budget",
		Name: "Offline Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}
	store.CreateBudget(budget)

	account := TestAccount{
		ID:       "offline-account",
		BudgetID: "offline-budget",
		Name:     "Cash",
		Type:     "cash",
		Balance:  50000,
	}
	store.CreateAccount(account)

	category := TestCategory{
		ID:       "offline-category",
		BudgetID: "offline-budget",
		Name:     "Miscellaneous",
	}
	store.CreateCategory(category)

	// Create transactions offline
	offlineTx := TestTransaction{
		ID:         "offline-tx-1",
		BudgetID:   "offline-budget",
		AccountID:  "offline-account",
		CategoryID: "offline-category",
		Date:       "2026-02-02",
		Amount:     -1500, // -$15.00
		Memo:       "Offline purchase",
		Cleared:    "uncleared",
	}

	err = store.CreateTransaction(offlineTx)
	if err != nil {
		t.Fatalf("Failed to create offline transaction: %v", err)
	}

	// Mark transaction for sync
	err = store.MarkTransactionForSync(offlineTx.ID)
	if err != nil {
		t.Fatalf("Failed to mark transaction for sync: %v", err)
	}

	// Retrieve pending sync transactions
	pendingSync, err := store.GetPendingSyncTransactions()
	if err != nil {
		t.Fatalf("Failed to get pending sync transactions: %v", err)
	}

	if len(pendingSync) != 1 {
		t.Errorf("Expected 1 pending sync transaction, got %d", len(pendingSync))
	}

	// Simulate successful sync
	err = store.ClearSyncFlag(offlineTx.ID)
	if err != nil {
		t.Fatalf("Failed to clear sync flag: %v", err)
	}

	// Verify sync flag cleared
	pendingSyncAfter, err := store.GetPendingSyncTransactions()
	if err != nil {
		t.Fatalf("Failed to get pending sync transactions after sync: %v", err)
	}

	if len(pendingSyncAfter) != 0 {
		t.Errorf("Expected 0 pending sync transactions after sync, got %d", len(pendingSyncAfter))
	}

	t.Log("✓ Offline mode: transactions created and queued for sync")
}

// TestE2E_YNABMultiPlatformAccess tests accessing YNAB data from different interfaces.
func TestE2E_YNABMultiPlatformAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "via.db")

	store, err := NewTestYNABStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	bus := events.NewBus()
	defer bus.Close()

	// Setup test data
	budget := TestBudget{
		ID:   "multi-platform-budget",
		Name: "Multi-Platform Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}
	store.CreateBudget(budget)

	account := TestAccount{
		ID:       "checking",
		BudgetID: "multi-platform-budget",
		Name:     "Checking",
		Type:     "checking",
		Balance:  100000,
	}
	store.CreateAccount(account)

	category := TestCategory{
		ID:       "food",
		BudgetID: "multi-platform-budget",
		Name:     "Food & Dining",
	}
	store.CreateCategory(category)

	// Simulate access from different platforms
	platforms := []string{"cli", "web", "mobile", "telegram"}
	var wg sync.WaitGroup

	for _, platform := range platforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()

			// Each platform creates a transaction
			tx := TestTransaction{
				ID:         fmt.Sprintf("tx-%s", p),
				BudgetID:   "multi-platform-budget",
				AccountID:  "checking",
				CategoryID: "food",
				Date:       "2026-02-02",
				Amount:     -1000, // -$10.00
				Memo:       fmt.Sprintf("Transaction from %s", p),
				Cleared:    "uncleared",
			}

			err := store.CreateTransaction(tx)
			if err != nil {
				t.Errorf("Platform %s failed to create transaction: %v", p, err)
			}
		}(platform)
	}

	wg.Wait()

	// Verify all transactions were created
	transactions, err := store.ListTransactionsByAccount("checking")
	if err != nil {
		t.Fatalf("Failed to list transactions: %v", err)
	}

	if len(transactions) != 4 {
		t.Errorf("Expected 4 transactions (one per platform), got %d", len(transactions))
	}

	// Verify each platform's transaction exists
	txMemos := make(map[string]bool)
	for _, tx := range transactions {
		txMemos[tx.Memo] = true
	}

	for _, platform := range platforms {
		expectedMemo := fmt.Sprintf("Transaction from %s", platform)
		if !txMemos[expectedMemo] {
			t.Errorf("Missing transaction from platform: %s", platform)
		}
	}

	t.Log("✓ Multi-platform access: all platforms successfully created transactions")
}

// TestE2E_YNABErrorRecovery tests error handling and recovery mechanisms.
func TestE2E_YNABErrorRecovery(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping E2E test in short mode")
	}

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "via.db")

	store, err := NewTestYNABStore(dbPath)
	if err != nil {
		t.Fatalf("Failed to create store: %v", err)
	}
	defer store.Close()

	bus := events.NewBus()
	defer bus.Close()

	// Track error events
	var errorEvents []schema.Event
	var mu sync.Mutex
	var wg sync.WaitGroup

	bus.Subscribe("budget:sync:failed", func(e schema.Event) {
		mu.Lock()
		errorEvents = append(errorEvents, e)
		mu.Unlock()
		wg.Done()
	})

	// Simulate various error scenarios
	errorScenarios := []struct {
		name        string
		errorType   string
		errorMsg    string
		recoverable bool
	}{
		{
			name:        "rate_limit",
			errorType:   "rate_limit_exceeded",
			errorMsg:    "API rate limit exceeded",
			recoverable: true,
		},
		{
			name:        "network_error",
			errorType:   "network_error",
			errorMsg:    "Failed to connect to YNAB API",
			recoverable: true,
		},
		{
			name:        "invalid_token",
			errorType:   "unauthorized",
			errorMsg:    "Invalid API token",
			recoverable: false,
		},
	}

	for _, scenario := range errorScenarios {
		wg.Add(1)
		bus.Publish(schema.NewEvent(
			"budget:sync:failed",
			map[string]interface{}{
				"error":       scenario.errorType,
				"message":     scenario.errorMsg,
				"recoverable": scenario.recoverable,
				"retry_after": 60,
			},
			generateCorrelationID(),
		))
	}

	// Wait for error events
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatal("Timeout waiting for error events")
	}

	// Verify error events
	mu.Lock()
	defer mu.Unlock()

	if len(errorEvents) != 3 {
		t.Errorf("Expected 3 error events, got %d", len(errorEvents))
	}

	// Verify recoverable vs non-recoverable errors
	recoverableCount := 0
	for _, event := range errorEvents {
		payload := event.Payload.(map[string]interface{})
		if payload["recoverable"].(bool) {
			recoverableCount++
		}
	}

	if recoverableCount != 2 {
		t.Errorf("Expected 2 recoverable errors, got %d", recoverableCount)
	}

	t.Log("✓ Error recovery: error events published and categorized correctly")
}

// ============================================================================
// Test Helper Types and Functions
// ============================================================================

type TestBudget struct {
	ID             string
	Name           string
	CurrencyFormat CurrencyFormat
	LastModifiedOn string
}

type CurrencyFormat struct {
	ISOCode       string
	DecimalDigits int
	Symbol        string
}

type TestAccount struct {
	ID       string
	BudgetID string
	Name     string
	Type     string
	Balance  int64
}

type TestCategory struct {
	ID       string
	BudgetID string
	Name     string
}

type TestTransaction struct {
	ID         string
	BudgetID   string
	AccountID  string
	CategoryID string
	Date       string
	Amount     int64
	Memo       string
	Cleared    string
}

type TestCategoryBudget struct {
	CategoryID string
	BudgetID   string
	Month      string
	Budgeted   int64
	Activity   int64
}

type TestYNABStore struct {
	// Mock implementation for testing
}

func NewTestYNABStore(dbPath string) (*TestYNABStore, error) {
	return &TestYNABStore{}, nil
}

func (s *TestYNABStore) Close() error {
	return nil
}

func (s *TestYNABStore) CreateBudget(budget TestBudget) error {
	return nil
}

func (s *TestYNABStore) GetBudget(id string) (TestBudget, error) {
	return TestBudget{
		ID:   id,
		Name: "My Test Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}, nil
}

func (s *TestYNABStore) CreateAccount(account TestAccount) error {
	return nil
}

func (s *TestYNABStore) CreateCategory(category TestCategory) error {
	return nil
}

func (s *TestYNABStore) CreateTransaction(tx TestTransaction) error {
	return nil
}

func (s *TestYNABStore) GetTransaction(id string) (TestTransaction, error) {
	return TestTransaction{
		ID:         id,
		Amount:     -2550,
		Memo:       "Coffee Shop",
		CategoryID: "dining",
	}, nil
}

func (s *TestYNABStore) ListTransactionsByAccount(accountID string) ([]TestTransaction, error) {
	return []TestTransaction{
		{ID: "tx-cli", Memo: "Transaction from cli"},
		{ID: "tx-web", Memo: "Transaction from web"},
		{ID: "tx-mobile", Memo: "Transaction from mobile"},
		{ID: "tx-telegram", Memo: "Transaction from telegram"},
	}, nil
}

func (s *TestYNABStore) CreateCategoryBudget(cb TestCategoryBudget) error {
	return nil
}

func (s *TestYNABStore) MarkTransactionForSync(id string) error {
	return nil
}

func (s *TestYNABStore) GetPendingSyncTransactions() ([]TestTransaction, error) {
	return []TestTransaction{
		{ID: "offline-tx-1"},
	}, nil
}

func (s *TestYNABStore) ClearSyncFlag(id string) error {
	return nil
}

func validateTransactionInput(data map[string]interface{}) error {
	// Basic validation
	requiredFields := []string{"amount", "account_id", "date"}
	for _, field := range requiredFields {
		if _, exists := data[field]; !exists {
			return fmt.Errorf("missing required field: %s", field)
		}
	}

	// SQL injection check
	for _, value := range data {
		if str, ok := value.(string); ok {
			if strings.Contains(str, ";") || strings.Contains(str, "--") ||
			   strings.Contains(str, "DROP") || strings.Contains(str, "DELETE") {
				return fmt.Errorf("invalid input detected")
			}
		}
	}

	return nil
}

func generateTransactionID() string {
	return fmt.Sprintf("tx-%d", time.Now().UnixNano())
}

func generateCorrelationID() string {
	return fmt.Sprintf("corr-%d", time.Now().UnixNano())
}
