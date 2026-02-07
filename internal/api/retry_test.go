package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestClient_RetryOnServerError(t *testing.T) {
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		if count < 3 {
			// First two attempts: return 500
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": {"id": "500", "name": "Internal Server Error", "detail": "Server error"}}`))
			return
		}
		// Third attempt: success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"budgets": []}}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	start := time.Now()
	_, err := client.GetBudgets()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Expected success after retries, got error: %v", err)
	}

	if atomic.LoadInt32(&attemptCount) != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}

	// Should have exponential backoff: 1s + 2s = 3s minimum
	if elapsed < 3*time.Second {
		t.Errorf("Expected at least 3s of backoff, got %v", elapsed)
	}
}

func TestClient_RetryOnRateLimit(t *testing.T) {
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		count := atomic.AddInt32(&attemptCount, 1)
		if count == 1 {
			// First attempt: rate limited with Retry-After header
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": {"id": "429", "name": "Too Many Requests", "detail": "Rate limit exceeded"}}`))
			return
		}
		// Second attempt: success
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"budgets": []}}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	start := time.Now()
	_, err := client.GetBudgets()
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Expected success after retry, got error: %v", err)
	}

	if atomic.LoadInt32(&attemptCount) != 2 {
		t.Errorf("Expected 2 attempts, got %d", attemptCount)
	}

	// Should wait for Retry-After: 2 seconds
	if elapsed < 2*time.Second {
		t.Errorf("Expected at least 2s wait for Retry-After, got %v", elapsed)
	}
}

func TestClient_NoRetryOnAuthError(t *testing.T) {
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"id": "401", "name": "Unauthorized", "detail": "Invalid token"}}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	_, err := client.GetBudgets()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	if !IsAuthError(err) {
		t.Errorf("Expected auth error, got: %v", err)
	}

	// Should not retry on auth errors
	if atomic.LoadInt32(&attemptCount) != 1 {
		t.Errorf("Expected 1 attempt (no retries), got %d", attemptCount)
	}
}

func TestClient_NoRetryOnBadRequest(t *testing.T) {
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"id": "400", "name": "Bad Request", "detail": "Invalid request"}}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	_, err := client.GetBudgets()

	if err == nil {
		t.Fatal("Expected error, got nil")
	}

	var ynabErr *YNABError
	if !IsYNABError(err) {
		t.Errorf("Expected YNAB error, got: %v", err)
	}

	if ynabErr != nil && ynabErr.StatusCode != http.StatusBadRequest {
		t.Errorf("Expected 400 status, got %d", ynabErr.StatusCode)
	}

	// Should not retry on client errors
	if atomic.LoadInt32(&attemptCount) != 1 {
		t.Errorf("Expected 1 attempt (no retries), got %d", attemptCount)
	}
}

func TestClient_MaxRetriesExhausted(t *testing.T) {
	var attemptCount int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attemptCount, 1)
		// Always return 500
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"id": "500", "name": "Internal Server Error", "detail": "Server error"}}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	start := time.Now()
	_, err := client.GetBudgets()
	elapsed := time.Since(start)

	if err == nil {
		t.Fatal("Expected error after max retries, got nil")
	}

	// Should attempt MaxRetries + 1 times (initial + retries)
	expectedAttempts := MaxRetries + 1
	if atomic.LoadInt32(&attemptCount) != int32(expectedAttempts) {
		t.Errorf("Expected %d attempts, got %d", expectedAttempts, attemptCount)
	}

	// With exponential backoff: 1s + 2s + 4s = 7s minimum
	if elapsed < 7*time.Second {
		t.Errorf("Expected at least 7s of backoff, got %v", elapsed)
	}

	// Error message should indicate retries were exhausted
	if err.Error() == "" {
		t.Error("Expected non-empty error message")
	}
}

func TestClient_ExponentialBackoffTiming(t *testing.T) {
	var attemptTimes []time.Time
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptTimes = append(attemptTimes, time.Now())
		// Always return 500 to trigger retries
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"id": "500", "name": "Internal Server Error", "detail": "Server error"}}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	client.GetBudgets()

	if len(attemptTimes) < 2 {
		t.Fatal("Not enough attempts to test backoff timing")
	}

	// Check exponential backoff between attempts
	// First backoff: ~1s
	backoff1 := attemptTimes[1].Sub(attemptTimes[0])
	if backoff1 < 900*time.Millisecond || backoff1 > 1500*time.Millisecond {
		t.Errorf("First backoff should be ~1s, got %v", backoff1)
	}

	if len(attemptTimes) >= 3 {
		// Second backoff: ~2s
		backoff2 := attemptTimes[2].Sub(attemptTimes[1])
		if backoff2 < 1800*time.Millisecond || backoff2 > 2500*time.Millisecond {
			t.Errorf("Second backoff should be ~2s, got %v", backoff2)
		}
	}

	if len(attemptTimes) >= 4 {
		// Third backoff: ~4s
		backoff3 := attemptTimes[3].Sub(attemptTimes[2])
		if backoff3 < 3800*time.Millisecond || backoff3 > 4500*time.Millisecond {
			t.Errorf("Third backoff should be ~4s, got %v", backoff3)
		}
	}
}

func TestClient_ParseErrorResponse(t *testing.T) {
	tests := []struct {
		name         string
		statusCode   int
		responseBody string
		wantMessage  string
		wantErrorID  string
		wantDetail   string
	}{
		{
			name:         "well-formed error response",
			statusCode:   400,
			responseBody: `{"error": {"id": "err_123", "name": "Bad Request", "detail": "Missing required field"}}`,
			wantMessage:  "Bad Request",
			wantErrorID:  "err_123",
			wantDetail:   "Missing required field",
		},
		{
			name:         "malformed error response",
			statusCode:   500,
			responseBody: `invalid json`,
			wantMessage:  "HTTP request failed: 500 Internal Server Error",
			wantErrorID:  "",
			wantDetail:   "",
		},
		{
			name:         "empty error response",
			statusCode:   404,
			responseBody: `{"error": {}}`,
			wantMessage:  "HTTP request failed: 404 Not Found",
			wantErrorID:  "",
			wantDetail:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.responseBody))
			}))
			defer server.Close()

			client := &Client{
				token:      "test-token",
				baseURL:    server.URL,
				httpClient: &http.Client{Timeout: 30 * time.Second},
			}

			_, err := client.GetBudgets()

			if err == nil {
				t.Fatal("Expected error, got nil")
			}

			if !IsYNABError(err) {
				t.Fatalf("Expected YNABError, got: %T", err)
			}

			// Extract the error details
			if ynabErr, ok := err.(*YNABError); ok {
				if ynabErr.Message != tt.wantMessage {
					t.Errorf("Message = %q, want %q", ynabErr.Message, tt.wantMessage)
				}
				if ynabErr.ErrorID != tt.wantErrorID {
					t.Errorf("ErrorID = %q, want %q", ynabErr.ErrorID, tt.wantErrorID)
				}
				if ynabErr.Detail != tt.wantDetail {
					t.Errorf("Detail = %q, want %q", ynabErr.Detail, tt.wantDetail)
				}
			}
		})
	}
}

func TestClient_RetryAfterHeaderParsing(t *testing.T) {
	tests := []struct {
		name        string
		retryAfter  string
		minDuration time.Duration
		maxDuration time.Duration
	}{
		{
			name:        "valid retry-after seconds",
			retryAfter:  "5",
			minDuration: 5 * time.Second,
			maxDuration: 6 * time.Second,
		},
		{
			name:        "invalid retry-after (defaults to 60)",
			retryAfter:  "invalid",
			minDuration: 60 * time.Second,
			maxDuration: 61 * time.Second,
		},
		{
			name:        "missing retry-after (defaults to 60)",
			retryAfter:  "",
			minDuration: 60 * time.Second,
			maxDuration: 61 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var attemptCount int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				count := atomic.AddInt32(&attemptCount, 1)
				if count == 1 {
					if tt.retryAfter != "" {
						w.Header().Set("Retry-After", tt.retryAfter)
					}
					w.WriteHeader(http.StatusTooManyRequests)
					w.Write([]byte(`{"error": {"name": "Too Many Requests"}}`))
					return
				}
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"data": {"budgets": []}}`))
			}))
			defer server.Close()

			client := &Client{
				token:      "test-token",
				baseURL:    server.URL,
				httpClient: &http.Client{Timeout: 120 * time.Second},
			}

			start := time.Now()
			_, err := client.GetBudgets()
			elapsed := time.Since(start)

			if err != nil {
				t.Fatalf("Expected success, got error: %v", err)
			}

			if elapsed < tt.minDuration {
				t.Errorf("Expected at least %v wait, got %v", tt.minDuration, elapsed)
			}

			// Allow for test overhead (up to 500ms)
			if elapsed > tt.maxDuration+500*time.Millisecond {
				t.Errorf("Expected at most %v wait, got %v", tt.maxDuration, elapsed)
			}
		})
	}
}

func BenchmarkClient_RetryLogic(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"data": {"budgets": []}}`))
	}))
	defer server.Close()

	client := &Client{
		token:      "test-token",
		baseURL:    server.URL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.GetBudgets()
	}
}

func ExampleYNABError_handling() {
	// Simulate different error scenarios
	authErr := NewAuthError()
	fmt.Println("Auth error:", authErr.IsAuthError())

	rateLimitErr := NewRateLimitError(120)
	fmt.Println("Rate limit retryable:", rateLimitErr.IsRetryable())

	serverErr := NewServerError(500)
	fmt.Println("Server error retryable:", serverErr.IsRetryable())

	// Output:
	// Auth error: true
	// Rate limit retryable: true
	// Server error retryable: true
}
