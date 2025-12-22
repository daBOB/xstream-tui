package xc

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.BaseDelay != time.Second {
		t.Errorf("BaseDelay = %v, want 1s", config.BaseDelay)
	}
	if config.MaxDelay != 10*time.Second {
		t.Errorf("MaxDelay = %v, want 10s", config.MaxDelay)
	}
}

func TestIsRetryable_NilError(t *testing.T) {
	if isRetryable(nil) {
		t.Error("nil error should not be retryable")
	}
}

func TestIsRetryable_ContextCanceled(t *testing.T) {
	if isRetryable(context.Canceled) {
		t.Error("context.Canceled should not be retryable")
	}
}

func TestIsRetryable_ContextDeadlineExceeded(t *testing.T) {
	if isRetryable(context.DeadlineExceeded) {
		t.Error("context.DeadlineExceeded should not be retryable")
	}
}

func TestIsRetryable_RetryableError(t *testing.T) {
	err := &retryableError{err: errors.New("test"), retryable: true}
	if !isRetryable(err) {
		t.Error("retryable error with retryable=true should be retryable")
	}

	err2 := &retryableError{err: errors.New("test"), retryable: false}
	if isRetryable(err2) {
		t.Error("retryable error with retryable=false should not be retryable")
	}
}

func TestIsRetryable_ConnectionErrors(t *testing.T) {
	tests := []struct {
		errMsg   string
		expected bool
	}{
		{"connection refused", true},
		{"connection reset", true},
		{"no such host", true},
		{"i/o timeout", true},
		{"some other error", false},
	}

	for _, tt := range tests {
		t.Run(tt.errMsg, func(t *testing.T) {
			err := errors.New(tt.errMsg)
			if isRetryable(err) != tt.expected {
				t.Errorf("isRetryable(%q) = %v, want %v", tt.errMsg, isRetryable(err), tt.expected)
			}
		})
	}
}

func TestIsRetryableStatus(t *testing.T) {
	tests := []struct {
		status   int
		expected bool
	}{
		{429, true},  // Too Many Requests
		{502, true},  // Bad Gateway
		{503, true},  // Service Unavailable
		{504, true},  // Gateway Timeout
		{200, false}, // OK
		{400, false}, // Bad Request
		{401, false}, // Unauthorized
		{403, false}, // Forbidden
		{404, false}, // Not Found
		{500, false}, // Internal Server Error
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.status)), func(t *testing.T) {
			if isRetryableStatus(tt.status) != tt.expected {
				t.Errorf("isRetryableStatus(%d) = %v, want %v", tt.status, isRetryableStatus(tt.status), tt.expected)
			}
		})
	}
}

func TestIsAuthError(t *testing.T) {
	tests := []struct {
		status   int
		expected bool
	}{
		{401, true},  // Unauthorized
		{402, true},  // Payment Required
		{403, true},  // Forbidden
		{200, false}, // OK
		{400, false}, // Bad Request
		{404, false}, // Not Found
		{500, false}, // Internal Server Error
	}

	for _, tt := range tests {
		t.Run(string(rune(tt.status)), func(t *testing.T) {
			if isAuthError(tt.status) != tt.expected {
				t.Errorf("isAuthError(%d) = %v, want %v", tt.status, isAuthError(tt.status), tt.expected)
			}
		})
	}
}

func TestCalculateBackoff(t *testing.T) {
	config := RetryConfig{
		BaseDelay: time.Second,
		MaxDelay:  10 * time.Second,
	}

	tests := []struct {
		attempt  int
		expected time.Duration
	}{
		{0, 1 * time.Second},  // 2^0 = 1
		{1, 2 * time.Second},  // 2^1 = 2
		{2, 4 * time.Second},  // 2^2 = 4
		{3, 8 * time.Second},  // 2^3 = 8
		{4, 10 * time.Second}, // 2^4 = 16, capped at 10
		{5, 10 * time.Second}, // 2^5 = 32, capped at 10
	}

	for _, tt := range tests {
		delay := calculateBackoff(tt.attempt, config)
		if delay != tt.expected {
			t.Errorf("calculateBackoff(%d) = %v, want %v", tt.attempt, delay, tt.expected)
		}
	}
}

func TestRetryableError(t *testing.T) {
	innerErr := errors.New("inner error")
	err := &retryableError{err: innerErr, statusCode: 503, retryable: true}

	if err.Error() != "inner error" {
		t.Errorf("Error() = %q, want 'inner error'", err.Error())
	}

	if !errors.Is(err, innerErr) {
		t.Error("errors.Is should find inner error")
	}
}

// Mock net.Error for testing
type mockNetError struct {
	timeout   bool
	temporary bool
}

func (m mockNetError) Error() string   { return "mock network error" }
func (m mockNetError) Timeout() bool   { return m.timeout }
func (m mockNetError) Temporary() bool { return m.temporary }

func TestIsRetryable_NetError(t *testing.T) {
	// Timeout errors are retryable
	timeoutErr := mockNetError{timeout: true}
	if !isRetryable(timeoutErr) {
		t.Error("timeout network error should be retryable")
	}

	// Non-timeout errors are not automatically retryable
	nonTimeoutErr := mockNetError{timeout: false}
	if isRetryable(nonTimeoutErr) {
		t.Error("non-timeout network error should not be retryable")
	}
}

func TestIsRetryable_DNSError(t *testing.T) {
	// Temporary DNS errors are retryable
	tempDNS := &net.DNSError{IsTemporary: true}
	if !isRetryable(tempDNS) {
		t.Error("temporary DNS error should be retryable")
	}

	// Non-temporary DNS errors are not retryable
	permDNS := &net.DNSError{IsTemporary: false}
	if isRetryable(permDNS) {
		t.Error("non-temporary DNS error should not be retryable")
	}
}
