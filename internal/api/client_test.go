package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name      string
		token     string
		envToken  string
		wantError bool
	}{
		{
			name:      "with token parameter",
			token:     "test-token",
			wantError: false,
		},
		{
			name:      "with env token",
			envToken:  "env-test-token",
			wantError: false,
		},
		{
			name:      "no token",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.envToken != "" {
				os.Setenv("YNAB_ACCESS_TOKEN", tt.envToken)
				defer os.Unsetenv("YNAB_ACCESS_TOKEN")
			}

			client, err := NewClient(tt.token)
			if tt.wantError {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if client == nil {
				t.Error("expected client, got nil")
			}
		})
	}
}

func TestClient_Request_Success(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
			t.Errorf("expected Authorization header 'Bearer test-token', got '%s'", auth)
		}
		if ct := r.Header.Get("Content-Type"); ct != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", ct)
		}

		// Return success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"message": "success",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-token")
	client.baseURL = server.URL

	respBody, err := client.request("GET", "/test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		t.Fatalf("failed to parse response: %v", err)
	}

	data, ok := response["data"].(map[string]interface{})
	if !ok {
		t.Fatal("expected data object in response")
	}

	if msg := data["message"].(string); msg != "success" {
		t.Errorf("expected message 'success', got '%s'", msg)
	}
}

func TestClient_Request_RateLimit(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// First request: rate limited
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"id":     "429",
					"name":   "rate_limit_exceeded",
					"detail": "Too many requests",
				},
			})
		} else {
			// Second request: success
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"message": "success",
				},
			})
		}
	}))
	defer server.Close()

	client, _ := NewClient("test-token")
	client.baseURL = server.URL

	start := time.Now()
	_, err := client.request("GET", "/test", nil)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have retried after ~1 second
	if elapsed < time.Second {
		t.Errorf("expected retry delay, got %v", elapsed)
	}

	if attempts != 2 {
		t.Errorf("expected 2 attempts, got %d", attempts)
	}
}

func TestClient_Request_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"id":     "400",
				"name":   "bad_request",
				"detail": "Invalid request",
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-token")
	client.baseURL = server.URL

	_, err := client.request("GET", "/test", nil)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	ynabErr, ok := err.(*YNABError)
	if !ok {
		t.Fatalf("expected YNABError, got %T", err)
	}

	if ynabErr.StatusCode != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", ynabErr.StatusCode)
	}

	if ynabErr.ErrorID != "400" {
		t.Errorf("expected error ID '400', got '%s'", ynabErr.ErrorID)
	}
}

func TestClient_GetBudgets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/budgets" {
			t.Errorf("expected path '/budgets', got '%s'", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BudgetsResponse{
			Data: struct {
				Budgets []*Budget `json:"budgets"`
			}{
				Budgets: []*Budget{
					{
						ID:             "budget-1",
						Name:           "My Budget",
						LastModifiedOn: "2026-01-01T00:00:00Z",
					},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-token")
	client.baseURL = server.URL

	budgets, err := client.GetBudgets()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(budgets) != 1 {
		t.Fatalf("expected 1 budget, got %d", len(budgets))
	}

	if budgets[0].ID != "budget-1" {
		t.Errorf("expected budget ID 'budget-1', got '%s'", budgets[0].ID)
	}
}

func TestClient_GetDefaultBudgetID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BudgetsResponse{
			Data: struct {
				Budgets []*Budget `json:"budgets"`
			}{
				Budgets: []*Budget{
					{ID: "default-budget", Name: "Default Budget"},
				},
			},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-token")
	client.baseURL = server.URL

	// First call should fetch from API
	id1, err := client.GetDefaultBudgetID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id1 != "default-budget" {
		t.Errorf("expected 'default-budget', got '%s'", id1)
	}

	// Second call should use cached value
	id2, err := client.GetDefaultBudgetID()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if id2 != id1 {
		t.Errorf("expected cached value '%s', got '%s'", id1, id2)
	}
}

func TestYNABError_Error(t *testing.T) {
	tests := []struct {
		name     string
		err      *YNABError
		contains []string
	}{
		{
			name: "basic error",
			err: &YNABError{
				Message: "test error",
			},
			contains: []string{"[YNAB]", "test error"},
		},
		{
			name: "error with status code",
			err: &YNABError{
				Message:    "test error",
				StatusCode: 400,
			},
			contains: []string{"[YNAB]", "test error", "HTTP 400"},
		},
		{
			name: "error with all fields",
			err: &YNABError{
				Message:    "test error",
				StatusCode: 400,
				ErrorID:    "ERR123",
				Detail:     "detailed message",
			},
			contains: []string{"[YNAB]", "test error", "HTTP 400", "ERR123", "detailed message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errMsg := tt.err.Error()
			for _, substr := range tt.contains {
				if !strings.Contains(errMsg, substr) {
					t.Errorf("expected error message to contain '%s', got: %s", substr, errMsg)
				}
			}
		})
	}
}

func TestTransactionRequest_Validate(t *testing.T) {
	tests := []struct {
		name      string
		req       *TransactionRequest
		wantError bool
	}{
		{
			name: "valid request",
			req: &TransactionRequest{
				AccountID: "account-1",
				Date:      "2026-01-01",
				Amount:    -10000,
			},
			wantError: false,
		},
		{
			name: "missing account_id",
			req: &TransactionRequest{
				Date:   "2026-01-01",
				Amount: -10000,
			},
			wantError: true,
		},
		{
			name: "missing date",
			req: &TransactionRequest{
				AccountID: "account-1",
				Amount:    -10000,
			},
			wantError: true,
		},
		{
			name: "cleared defaults to uncleared",
			req: &TransactionRequest{
				AccountID: "account-1",
				Date:      "2026-01-01",
				Amount:    -10000,
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantError && tt.req.Cleared == "" {
				if tt.req.Cleared != "uncleared" {
					t.Errorf("expected cleared to default to 'uncleared', got '%s'", tt.req.Cleared)
				}
			}
		})
	}
}
