package xc

import (
	"context"
	"errors"
	"net"
	"net/http"
	"strings"
	"time"
)

// RetryConfig configures retry behavior for API calls.
type RetryConfig struct {
	MaxRetries int           // Maximum number of retry attempts (default: 3)
	BaseDelay  time.Duration // Initial delay between retries (default: 1s)
	MaxDelay   time.Duration // Maximum delay between retries (default: 10s)
}

// DefaultRetryConfig returns sensible default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries: 3,
		BaseDelay:  1 * time.Second,
		MaxDelay:   10 * time.Second,
	}
}

// retryableError wraps errors with retry information.
type retryableError struct {
	err        error
	statusCode int
	retryable  bool
}

func (e *retryableError) Error() string {
	return e.err.Error()
}

func (e *retryableError) Unwrap() error {
	return e.err
}

// isRetryable determines if an error warrants a retry attempt.
// Returns true for transient network errors and specific HTTP status codes.
func isRetryable(err error) bool {
	if err == nil {
		return false
	}

	// Check for context cancellation - never retry
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	// Check for retryable error wrapper
	var re *retryableError
	if errors.As(err, &re) {
		return re.retryable
	}

	// Check for network errors (typically transient)
	var netErr net.Error
	if errors.As(err, &netErr) {
		// Timeout errors are retryable
		if netErr.Timeout() {
			return true
		}
	}

	// Check for DNS errors
	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		// Temporary DNS failures are retryable
		return dnsErr.Temporary()
	}

	// Check for connection refused/reset (common transient issues)
	errStr := err.Error()
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "i/o timeout") {
		return true
	}

	return false
}

// isRetryableStatus returns true for HTTP status codes that warrant retry.
func isRetryableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusTooManyRequests,     // 429 - Rate limited
		http.StatusBadGateway,           // 502 - Gateway error
		http.StatusServiceUnavailable,   // 503 - Service unavailable
		http.StatusGatewayTimeout:       // 504 - Gateway timeout
		return true
	default:
		return false
	}
}

// isAuthError returns true for authentication/authorization errors.
// These should never be retried as they indicate credential problems.
func isAuthError(statusCode int) bool {
	switch statusCode {
	case http.StatusUnauthorized, // 401
		http.StatusForbidden,    // 403
		http.StatusPaymentRequired: // 402 (some APIs use this for expired accounts)
		return true
	default:
		return false
	}
}

// calculateBackoff returns the delay for the given retry attempt.
// Uses exponential backoff: baseDelay * 2^attempt, capped at maxDelay.
func calculateBackoff(attempt int, config RetryConfig) time.Duration {
	delay := config.BaseDelay * (1 << attempt) // 2^attempt
	if delay > config.MaxDelay {
		delay = config.MaxDelay
	}
	return delay
}
