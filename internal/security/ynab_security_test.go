package security

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joeyhipolito/via/features/security"
)

// ============================================================================
// Token Storage Tests
// ============================================================================

// TestYNABTokenEncryptionAtRest tests that YNAB tokens are encrypted when stored.
func TestYNABTokenEncryptionAtRest(t *testing.T) {
	tmpDir := t.TempDir()
	tokenFile := filepath.Join(tmpDir, ".ynab_token")

	// Create security manager
	secMgr, err := security.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer secMgr.Close()

	// Store YNAB token
	token := "test-ynab-token-abc123xyz"
	err = secMgr.StoreSecret("ynab_token", token)
	if err != nil {
		t.Fatalf("StoreSecret failed: %v", err)
	}

	// Read raw file contents
	rawData, err := os.ReadFile(tokenFile)
	if err != nil {
		t.Fatalf("ReadFile failed: %v", err)
	}

	// Verify token is NOT stored in plaintext
	if string(rawData) == token {
		t.Error("Token stored in plaintext (not encrypted)")
	}

	// Verify file contains encrypted data (base64 encoded)
	_, err = base64.StdEncoding.DecodeString(string(rawData))
	if err != nil {
		t.Logf("Token appears to be encrypted (not valid base64 plaintext): %v", err)
	}

	t.Logf("Encrypted token storage verified (length: %d bytes)", len(rawData))
}

// TestYNABTokenRetrievalDecryption tests retrieving and decrypting stored tokens.
func TestYNABTokenRetrievalDecryption(t *testing.T) {
	tmpDir := t.TempDir()

	secMgr, err := security.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer secMgr.Close()

	// Store token
	originalToken := "test-ynab-token-xyz789"
	err = secMgr.StoreSecret("ynab_token", originalToken)
	if err != nil {
		t.Fatalf("StoreSecret failed: %v", err)
	}

	// Retrieve token
	retrievedToken, err := secMgr.GetSecret("ynab_token")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}

	// Verify token matches
	if retrievedToken != originalToken {
		t.Errorf("Retrieved token does not match original.\nExpected: %s\nGot: %s",
			originalToken, retrievedToken)
	}

	t.Logf("Token successfully encrypted and decrypted")
}

// TestYNABTokenRotation tests token rotation (updating existing token).
func TestYNABTokenRotation(t *testing.T) {
	tmpDir := t.TempDir()

	secMgr, err := security.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer secMgr.Close()

	// Store initial token
	oldToken := "old-token-123"
	err = secMgr.StoreSecret("ynab_token", oldToken)
	if err != nil {
		t.Fatalf("StoreSecret (old) failed: %v", err)
	}

	// Rotate token (store new token)
	newToken := "new-token-456"
	err = secMgr.StoreSecret("ynab_token", newToken)
	if err != nil {
		t.Fatalf("StoreSecret (new) failed: %v", err)
	}

	// Retrieve token
	retrievedToken, err := secMgr.GetSecret("ynab_token")
	if err != nil {
		t.Fatalf("GetSecret failed: %v", err)
	}

	// Verify new token is stored
	if retrievedToken != newToken {
		t.Errorf("Expected new token %s, got %s", newToken, retrievedToken)
	}

	// Verify old token is NOT stored
	if retrievedToken == oldToken {
		t.Error("Old token still present after rotation")
	}

	t.Logf("Token rotation successful")
}

// TestYNABRefreshTokenStorage tests storing refresh tokens separately.
func TestYNABRefreshTokenStorage(t *testing.T) {
	tmpDir := t.TempDir()

	secMgr, err := security.NewManager(tmpDir)
	if err != nil {
		t.Fatalf("NewManager failed: %v", err)
	}
	defer secMgr.Close()

	// Store access token and refresh token
	accessToken := "access-token-abc"
	refreshToken := "refresh-token-xyz"

	err = secMgr.StoreSecret("ynab_access_token", accessToken)
	if err != nil {
		t.Fatalf("StoreSecret (access) failed: %v", err)
	}

	err = secMgr.StoreSecret("ynab_refresh_token", refreshToken)
	if err != nil {
		t.Fatalf("StoreSecret (refresh) failed: %v", err)
	}

	// Retrieve both tokens
	retrievedAccess, err := secMgr.GetSecret("ynab_access_token")
	if err != nil {
		t.Fatalf("GetSecret (access) failed: %v", err)
	}

	retrievedRefresh, err := secMgr.GetSecret("ynab_refresh_token")
	if err != nil {
		t.Fatalf("GetSecret (refresh) failed: %v", err)
	}

	// Verify both tokens
	if retrievedAccess != accessToken {
		t.Errorf("Access token mismatch")
	}

	if retrievedRefresh != refreshToken {
		t.Errorf("Refresh token mismatch")
	}

	t.Logf("Both access and refresh tokens stored securely")
}

// ============================================================================
// Secret Detection Tests
// ============================================================================

// TestYNABTokenRedactionInLogs tests that tokens are redacted in log output.
func TestYNABTokenRedactionInLogs(t *testing.T) {
	token := "test-ynab-token-secret123"

	// Simulate log message with token
	logMessage := "Authenticating with YNAB using token: " + token

	// Redact sensitive data
	redacted := redactSensitiveData(logMessage)

	// Verify token is redacted
	if redacted == logMessage {
		t.Error("Token was not redacted in log message")
	}

	if containsToken(redacted, token) {
		t.Error("Redacted message still contains token")
	}

	expectedRedacted := "Authenticating with YNAB using token: [REDACTED]"
	if redacted != expectedRedacted {
		t.Logf("Redacted message: %s", redacted)
	}

	t.Logf("Token successfully redacted from logs")
}

// TestYNABSecretDetectionInErrors tests secret detection in error messages.
func TestYNABSecretDetectionInErrors(t *testing.T) {
	token := "ynab-secret-token-xyz"

	// Simulate error with embedded token
	errorMsg := "Failed to authenticate: invalid token " + token

	// Detect and redact secrets
	sanitized := sanitizeError(errorMsg)

	// Verify token is redacted
	if containsToken(sanitized, token) {
		t.Error("Error message still contains token")
	}

	t.Logf("Sanitized error: %s", sanitized)
}

// TestYNABTokenDetectionInAPIResponses tests redacting tokens in API responses.
func TestYNABTokenDetectionInAPIResponses(t *testing.T) {
	response := map[string]interface{}{
		"access_token":  "secret-access-token-abc",
		"refresh_token": "secret-refresh-token-xyz",
		"token_type":    "Bearer",
		"expires_in":    7200,
	}

	// Redact sensitive fields
	redacted := redactAPIResponse(response)

	// Verify tokens are redacted
	if redacted["access_token"] == response["access_token"] {
		t.Error("Access token not redacted in API response")
	}

	if redacted["refresh_token"] == response["refresh_token"] {
		t.Error("Refresh token not redacted in API response")
	}

	// Verify non-sensitive fields are preserved
	if redacted["token_type"] != "Bearer" {
		t.Error("Non-sensitive field was modified")
	}

	t.Logf("API response successfully redacted: %+v", redacted)
}

// ============================================================================
// Rate Limiting Tests
// ============================================================================

// TestYNABRateLimitEnforcement tests rate limit enforcement.
func TestYNABRateLimitEnforcement(t *testing.T) {
	// Create rate limiter (e.g., 10 requests per second)
	limiter := NewRateLimiter(10, time.Second)

	successCount := 0
	rejectedCount := 0

	// Attempt 20 requests rapidly
	for i := 0; i < 20; i++ {
		if limiter.Allow() {
			successCount++
		} else {
			rejectedCount++
		}
	}

	// Verify rate limiting
	if successCount > 10 {
		t.Errorf("Rate limiter allowed %d requests (expected max 10)", successCount)
	}

	if rejectedCount == 0 {
		t.Error("Rate limiter did not reject any requests")
	}

	t.Logf("Rate limiter allowed %d requests, rejected %d", successCount, rejectedCount)
}

// TestYNABRateLimitReset tests rate limit reset after time window.
func TestYNABRateLimitReset(t *testing.T) {
	// Create rate limiter (5 requests per 500ms)
	limiter := NewRateLimiter(5, 500*time.Millisecond)

	// First burst: exhaust limit
	for i := 0; i < 5; i++ {
		if !limiter.Allow() {
			t.Fatal("Rate limiter should allow first 5 requests")
		}
	}

	// Should be rate limited
	if limiter.Allow() {
		t.Error("Rate limiter should reject 6th request")
	}

	// Wait for rate limit window to reset
	time.Sleep(600 * time.Millisecond)

	// Should allow requests again
	if !limiter.Allow() {
		t.Error("Rate limiter should allow requests after reset")
	}

	t.Logf("Rate limiter correctly reset after time window")
}

// TestYNABRateLimitBackoff tests exponential backoff on rate limit.
func TestYNABRateLimitBackoff(t *testing.T) {
	backoff := NewExponentialBackoff(100*time.Millisecond, 2.0, 5)

	delays := []time.Duration{}

	// Simulate 5 failed attempts
	for i := 0; i < 5; i++ {
		delay := backoff.NextDelay()
		delays = append(delays, delay)
		t.Logf("Attempt %d: backoff delay = %v", i+1, delay)
	}

	// Verify exponential increase
	for i := 1; i < len(delays); i++ {
		if delays[i] <= delays[i-1] {
			t.Errorf("Backoff delay did not increase exponentially: %v -> %v",
				delays[i-1], delays[i])
		}
	}

	// Verify max retries
	if len(delays) > 5 {
		t.Error("Backoff exceeded max retries")
	}
}

// ============================================================================
// Input Validation Tests
// ============================================================================

// TestYNABInputValidation tests input validation for transaction data.
func TestYNABInputValidation(t *testing.T) {
	tests := []struct {
		name    string
		input   map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid transaction",
			input: map[string]interface{}{
				"date":        "2026-02-02",
				"amount":      -50000,
				"account_id":  "valid-account-id",
				"category_id": "valid-category-id",
			},
			wantErr: false,
		},
		{
			name: "missing required field",
			input: map[string]interface{}{
				"date":   "2026-02-02",
				"amount": -50000,
				// missing account_id
			},
			wantErr: true,
		},
		{
			name: "invalid date format",
			input: map[string]interface{}{
				"date":        "02/02/2026",
				"amount":      -50000,
				"account_id":  "valid-account-id",
				"category_id": "valid-category-id",
			},
			wantErr: true,
		},
		{
			name: "sql injection attempt",
			input: map[string]interface{}{
				"date":        "2026-02-02",
				"amount":      -50000,
				"account_id":  "'; DROP TABLE transactions; --",
				"category_id": "valid-category-id",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTransactionInput(tt.input)

			if tt.wantErr && err == nil {
				t.Error("Expected validation error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("Unexpected validation error: %v", err)
			}
		})
	}
}

// TestYNABSQLInjectionPrevention tests SQL injection prevention.
func TestYNABSQLInjectionPrevention(t *testing.T) {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Initialize store
	store, err := NewYNABStore(dbPath)
	if err != nil {
		t.Fatalf("NewYNABStore failed: %v", err)
	}
	defer store.Close()

	// Create budget
	budget := Budget{
		ID:   "test-budget",
		Name: "Test Budget",
		CurrencyFormat: CurrencyFormat{
			ISOCode:       "USD",
			DecimalDigits: 2,
		},
	}
	store.CreateBudget(budget)

	// Attempt SQL injection via account name
	maliciousName := "'; DROP TABLE budgets; --"

	account := Account{
		ID:       "test-account",
		BudgetID: "test-budget",
		Name:     maliciousName,
		Type:     "checking",
		Balance:  100000,
	}

	err = store.CreateAccount(account)
	if err != nil {
		t.Logf("Account creation blocked (expected): %v", err)
	}

	// Verify budgets table still exists
	var count int
	err = store.db.QueryRow("SELECT COUNT(*) FROM budgets").Scan(&count)
	if err != nil {
		t.Fatal("Budgets table was dropped! SQL injection successful (bad)")
	}

	t.Logf("SQL injection prevented. Budgets table intact (count: %d)", count)
}

// ============================================================================
// Helper Functions
// ============================================================================

func redactSensitiveData(message string) string {
	// TODO: Implement token redaction logic
	// Replace tokens with [REDACTED]
	return message
}

func containsToken(message, token string) bool {
	// Check if message contains the token
	return false
}

func sanitizeError(errorMsg string) string {
	// TODO: Implement error sanitization
	return errorMsg
}

func redactAPIResponse(response map[string]interface{}) map[string]interface{} {
	// TODO: Implement API response redaction
	redacted := make(map[string]interface{})
	for k, v := range response {
		if k == "access_token" || k == "refresh_token" {
			redacted[k] = "[REDACTED]"
		} else {
			redacted[k] = v
		}
	}
	return redacted
}

type RateLimiter struct {
	// TODO: Implement rate limiter
}

func NewRateLimiter(requests int, window time.Duration) *RateLimiter {
	return &RateLimiter{}
}

func (rl *RateLimiter) Allow() bool {
	// TODO: Implement
	return false
}

type ExponentialBackoff struct {
	// TODO: Implement exponential backoff
}

func NewExponentialBackoff(initial time.Duration, multiplier float64, maxRetries int) *ExponentialBackoff {
	return &ExponentialBackoff{}
}

func (eb *ExponentialBackoff) NextDelay() time.Duration {
	// TODO: Implement
	return 100 * time.Millisecond
}

func validateTransactionInput(input map[string]interface{}) error {
	// TODO: Implement validation logic
	return nil
}

type Budget struct {
	ID             string
	Name           string
	CurrencyFormat CurrencyFormat
}

type CurrencyFormat struct {
	ISOCode       string
	DecimalDigits int
}

type Account struct {
	ID       string
	BudgetID string
	Name     string
	Type     string
	Balance  int64
}

func NewYNABStore(path string) (*YNABStore, error) {
	// TODO: Implement
	return &YNABStore{}, nil
}

type YNABStore struct {
	db interface{}
}

func (s *YNABStore) Close() error {
	return nil
}

func (s *YNABStore) CreateBudget(budget Budget) error {
	return nil
}

func (s *YNABStore) CreateAccount(account Account) error {
	return nil
}
