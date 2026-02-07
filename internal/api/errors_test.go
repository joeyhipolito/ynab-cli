package api

import (
	"errors"
	"net/http"
	"testing"
)

func TestYNABError_IsAuthError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"401 Unauthorized", http.StatusUnauthorized, true},
		{"403 Forbidden", http.StatusForbidden, false},
		{"404 Not Found", http.StatusNotFound, false},
		{"500 Server Error", http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &YNABError{StatusCode: tt.statusCode}
			if got := err.IsAuthError(); got != tt.want {
				t.Errorf("IsAuthError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYNABError_IsRateLimitError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"429 Too Many Requests", http.StatusTooManyRequests, true},
		{"401 Unauthorized", http.StatusUnauthorized, false},
		{"500 Server Error", http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &YNABError{StatusCode: tt.statusCode}
			if got := err.IsRateLimitError(); got != tt.want {
				t.Errorf("IsRateLimitError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYNABError_IsServerError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"500 Internal Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
		{"504 Gateway Timeout", http.StatusGatewayTimeout, true},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"404 Not Found", http.StatusNotFound, false},
		{"429 Rate Limit", http.StatusTooManyRequests, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &YNABError{StatusCode: tt.statusCode}
			if got := err.IsServerError(); got != tt.want {
				t.Errorf("IsServerError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYNABError_IsNotFoundError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"404 Not Found", http.StatusNotFound, true},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"500 Server Error", http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &YNABError{StatusCode: tt.statusCode}
			if got := err.IsNotFoundError(); got != tt.want {
				t.Errorf("IsNotFoundError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYNABError_IsBadRequestError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"400 Bad Request", http.StatusBadRequest, true},
		{"404 Not Found", http.StatusNotFound, false},
		{"500 Server Error", http.StatusInternalServerError, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &YNABError{StatusCode: tt.statusCode}
			if got := err.IsBadRequestError(); got != tt.want {
				t.Errorf("IsBadRequestError() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestYNABError_IsRetryable(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		want       bool
	}{
		{"429 Rate Limit", http.StatusTooManyRequests, true},
		{"500 Server Error", http.StatusInternalServerError, true},
		{"502 Bad Gateway", http.StatusBadGateway, true},
		{"503 Service Unavailable", http.StatusServiceUnavailable, true},
		{"400 Bad Request", http.StatusBadRequest, false},
		{"401 Unauthorized", http.StatusUnauthorized, false},
		{"404 Not Found", http.StatusNotFound, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := &YNABError{StatusCode: tt.statusCode}
			if got := err.IsRetryable(); got != tt.want {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsYNABError(t *testing.T) {
	ynabErr := &YNABError{Message: "test"}
	stdErr := errors.New("standard error")

	if !IsYNABError(ynabErr) {
		t.Error("IsYNABError() should return true for YNABError")
	}

	if IsYNABError(stdErr) {
		t.Error("IsYNABError() should return false for standard error")
	}
}

func TestIsAuthError(t *testing.T) {
	authErr := &YNABError{StatusCode: http.StatusUnauthorized}
	otherErr := &YNABError{StatusCode: http.StatusBadRequest}
	stdErr := errors.New("standard error")

	if !IsAuthError(authErr) {
		t.Error("IsAuthError() should return true for 401 error")
	}

	if IsAuthError(otherErr) {
		t.Error("IsAuthError() should return false for non-401 error")
	}

	if IsAuthError(stdErr) {
		t.Error("IsAuthError() should return false for standard error")
	}
}

func TestIsRateLimitError(t *testing.T) {
	rateLimitErr := &YNABError{StatusCode: http.StatusTooManyRequests}
	otherErr := &YNABError{StatusCode: http.StatusBadRequest}
	stdErr := errors.New("standard error")

	if !IsRateLimitError(rateLimitErr) {
		t.Error("IsRateLimitError() should return true for 429 error")
	}

	if IsRateLimitError(otherErr) {
		t.Error("IsRateLimitError() should return false for non-429 error")
	}

	if IsRateLimitError(stdErr) {
		t.Error("IsRateLimitError() should return false for standard error")
	}
}

func TestIsServerError(t *testing.T) {
	serverErr := &YNABError{StatusCode: http.StatusInternalServerError}
	clientErr := &YNABError{StatusCode: http.StatusBadRequest}
	stdErr := errors.New("standard error")

	if !IsServerError(serverErr) {
		t.Error("IsServerError() should return true for 5xx error")
	}

	if IsServerError(clientErr) {
		t.Error("IsServerError() should return false for 4xx error")
	}

	if IsServerError(stdErr) {
		t.Error("IsServerError() should return false for standard error")
	}
}

func TestIsNotFoundError(t *testing.T) {
	notFoundErr := &YNABError{StatusCode: http.StatusNotFound}
	otherErr := &YNABError{StatusCode: http.StatusBadRequest}
	stdErr := errors.New("standard error")

	if !IsNotFoundError(notFoundErr) {
		t.Error("IsNotFoundError() should return true for 404 error")
	}

	if IsNotFoundError(otherErr) {
		t.Error("IsNotFoundError() should return false for non-404 error")
	}

	if IsNotFoundError(stdErr) {
		t.Error("IsNotFoundError() should return false for standard error")
	}
}

func TestNewAuthError(t *testing.T) {
	err := NewAuthError()

	if err.StatusCode != http.StatusUnauthorized {
		t.Errorf("NewAuthError() status code = %d, want %d", err.StatusCode, http.StatusUnauthorized)
	}

	if !err.IsAuthError() {
		t.Error("NewAuthError() should create an auth error")
	}

	if err.Message == "" {
		t.Error("NewAuthError() should have a message")
	}
}

func TestNewRateLimitError(t *testing.T) {
	retryAfter := 120
	err := NewRateLimitError(retryAfter)

	if err.StatusCode != http.StatusTooManyRequests {
		t.Errorf("NewRateLimitError() status code = %d, want %d", err.StatusCode, http.StatusTooManyRequests)
	}

	if !err.IsRateLimitError() {
		t.Error("NewRateLimitError() should create a rate limit error")
	}

	if err.Message == "" {
		t.Error("NewRateLimitError() should have a message")
	}
}

func TestNewServerError(t *testing.T) {
	statusCode := http.StatusInternalServerError
	err := NewServerError(statusCode)

	if err.StatusCode != statusCode {
		t.Errorf("NewServerError() status code = %d, want %d", err.StatusCode, statusCode)
	}

	if !err.IsServerError() {
		t.Error("NewServerError() should create a server error")
	}

	if err.Message == "" {
		t.Error("NewServerError() should have a message")
	}
}
