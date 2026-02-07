package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestGetBudgets tests the GetBudgets method.
func TestGetBudgets(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/budgets" {
			t.Errorf("Expected path /budgets, got %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("Expected Authorization header with Bearer token")
		}

		response := BudgetsResponse{}
		response.Data.Budgets = []*Budget{
			{ID: "budget-1", Name: "Test Budget"},
			{ID: "budget-2", Name: "Another Budget"},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	budgets, err := client.GetBudgets()
	if err != nil {
		t.Fatalf("GetBudgets failed: %v", err)
	}

	if len(budgets) != 2 {
		t.Errorf("Expected 2 budgets, got %d", len(budgets))
	}
	if budgets[0].Name != "Test Budget" {
		t.Errorf("Expected first budget name 'Test Budget', got %s", budgets[0].Name)
	}
}

// TestGetAccounts tests the GetAccounts method.
func TestGetAccounts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/budgets/test-budget/accounts" {
			t.Errorf("Expected path /budgets/test-budget/accounts, got %s", r.URL.Path)
		}

		response := AccountsResponse{}
		response.Data.Accounts = []*Account{
			{ID: "acc-1", Name: "Checking", Type: "checking", Balance: 100000},
			{ID: "acc-2", Name: "Savings", Type: "savings", Balance: 500000},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	accounts, err := client.GetAccounts("test-budget")
	if err != nil {
		t.Fatalf("GetAccounts failed: %v", err)
	}

	if len(accounts) != 2 {
		t.Errorf("Expected 2 accounts, got %d", len(accounts))
	}
	if accounts[0].Name != "Checking" {
		t.Errorf("Expected first account name 'Checking', got %s", accounts[0].Name)
	}
}

// TestGetCategories tests the GetCategories method.
func TestGetCategories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/budgets/test-budget/categories" {
			t.Errorf("Expected path /budgets/test-budget/categories, got %s", r.URL.Path)
		}

		response := CategoriesResponse{}
		response.Data.CategoryGroups = []*CategoryGroup{
			{
				ID:   "group-1",
				Name: "Monthly Bills",
				Categories: []*Category{
					{ID: "cat-1", Name: "Rent", Budgeted: 100000},
					{ID: "cat-2", Name: "Utilities", Budgeted: 20000},
				},
			},
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	groups, err := client.GetCategories("test-budget")
	if err != nil {
		t.Fatalf("GetCategories failed: %v", err)
	}

	if len(groups) != 1 {
		t.Errorf("Expected 1 category group, got %d", len(groups))
	}
	if len(groups[0].Categories) != 2 {
		t.Errorf("Expected 2 categories, got %d", len(groups[0].Categories))
	}
}

// TestCreateTransaction tests the CreateTransaction method.
func TestCreateTransaction(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if r.URL.Path != "/budgets/test-budget/transactions" {
			t.Errorf("Expected path /budgets/test-budget/transactions, got %s", r.URL.Path)
		}

		// Decode request to verify it's correct
		var reqBody struct {
			Transaction map[string]interface{} `json:"transaction"`
		}
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Return created transaction
		response := TransactionResponse{}
		response.Data.Transaction = &Transaction{
			ID:        "txn-123",
			AccountID: "acc-1",
			Date:      "2024-01-15",
			Amount:    -5000,
			Memo:      "Test transaction",
			Approved:  true,
			Cleared:   "uncleared",
		}
		json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 5 * time.Second},
	}

	req := &TransactionRequest{
		BudgetID:  "test-budget",
		AccountID: "acc-1",
		Date:      "2024-01-15",
		Amount:    -5000,
		Memo:      "Test transaction",
		Cleared:   "uncleared",
		Approved:  true,
	}

	txn, err := client.CreateTransaction(req)
	if err != nil {
		t.Fatalf("CreateTransaction failed: %v", err)
	}

	if txn.ID != "txn-123" {
		t.Errorf("Expected transaction ID 'txn-123', got %s", txn.ID)
	}
	if txn.Amount != -5000 {
		t.Errorf("Expected amount -5000, got %d", txn.Amount)
	}
}

// TestTransactionRequestValidation tests the Validate method.
func TestTransactionRequestValidation(t *testing.T) {
	tests := []struct {
		name    string
		req     *TransactionRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &TransactionRequest{
				AccountID: "acc-1",
				Date:      "2024-01-15",
				Amount:    -5000,
			},
			wantErr: false,
		},
		{
			name: "missing account_id",
			req: &TransactionRequest{
				Date:   "2024-01-15",
				Amount: -5000,
			},
			wantErr: true,
		},
		{
			name: "missing date",
			req: &TransactionRequest{
				AccountID: "acc-1",
				Amount:    -5000,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
