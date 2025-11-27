// Package retry provides retry logic with exponential backoff for API operations.
package retry

import (
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"fmt"
	"math"
	"time"

	"github.com/creiche/confluent-go/pkg/api"
)

// Strategy defines how retries should be performed.
type Strategy struct {
	maxAttempts     int
	initialBackoff  time.Duration
	maxBackoff      time.Duration
	multiplier      float64
	addJitter       bool
	retryableErrors func(*api.Error) bool
}

// DefaultStrategy returns a Strategy with sensible defaults:
// - Max 5 attempts
// - Initial backoff of 1 second
// - Max backoff of 60 seconds
// - Exponential multiplier of 2.0
// - Jitter enabled
// - Retries on 429 (rate limit) and 500+ (server errors)
func DefaultStrategy() *Strategy {
	return &Strategy{
		maxAttempts:     5,
		initialBackoff:  1 * time.Second,
		maxBackoff:      60 * time.Second,
		multiplier:      2.0,
		addJitter:       true,
		retryableErrors: DefaultRetryableErrors,
	}
}

// WithMaxAttempts sets the maximum number of attempts (including initial attempt).
// Default is 5.
func (s *Strategy) WithMaxAttempts(attempts int) *Strategy {
	if attempts < 1 {
		attempts = 1
	}
	s.maxAttempts = attempts
	return s
}

// WithInitialBackoff sets the initial backoff duration.
// Default is 1 second.
func (s *Strategy) WithInitialBackoff(d time.Duration) *Strategy {
	if d < 0 {
		d = 0
	}
	s.initialBackoff = d
	return s
}

// WithMaxBackoff sets the maximum backoff duration.
// Default is 60 seconds.
func (s *Strategy) WithMaxBackoff(d time.Duration) *Strategy {
	if d < 0 {
		d = 0
	}
	s.maxBackoff = d
	return s
}

// WithMultiplier sets the exponential multiplier for backoff.
// Default is 2.0 (doubling each attempt).
func (s *Strategy) WithMultiplier(m float64) *Strategy {
	if m < 1.0 {
		m = 1.0
	}
	s.multiplier = m
	return s
}

// WithJitter enables or disables jitter (random variation) in backoff times.
// Default is true. Jitter helps prevent thundering herd.
func (s *Strategy) WithJitter(enabled bool) *Strategy {
	s.addJitter = enabled
	return s
}

// WithRetryableErrors sets the function that determines if an error should be retried.
func (s *Strategy) WithRetryableErrors(fn func(*api.Error) bool) *Strategy {
	if fn != nil {
		s.retryableErrors = fn
	}
	return s
}

// Do executes the operation with retry logic.
// It retries on retryable errors (rate limiting and server errors) up to maxAttempts times.
// Returns the operation result or the last error if all retries fail.
func (s *Strategy) Do(ctx context.Context, operation func() error) error {
	var lastErr error

	for attempt := 1; attempt <= s.maxAttempts; attempt++ {
		// Check if context is already cancelled
		select {
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled: %w", ctx.Err())
		default:
		}

		// Execute the operation
		err := operation()
		if err == nil {
			return nil // Success
		}

		lastErr = err

		// Check if error is retryable
		apiErr, ok := err.(*api.Error)
		if !ok || !s.retryableErrors(apiErr) {
			// Not retryable, fail immediately
			return err
		}

		// Don't backoff on last attempt
		if attempt >= s.maxAttempts {
			break
		}

		// Calculate backoff duration
		waitDuration := s.calculateBackoff(attempt - 1)

		// Use Retry-After header if available
		if apiErr.IsRateLimited() {
			retryAfter := apiErr.RetryAfter()
			if retryAfter > 0 {
				waitDuration = time.Duration(retryAfter) * time.Second
			}
		}

		// Wait before retrying
		select {
		case <-time.After(waitDuration):
			// Continue to next attempt
		case <-ctx.Done():
			return fmt.Errorf("retry cancelled after attempt %d: %w", attempt, ctx.Err())
		}
	}

	if lastErr != nil {
		apiErr, ok := lastErr.(*api.Error)
		if ok {
			return fmt.Errorf("operation failed after %d attempts: %w", s.maxAttempts, apiErr)
		}
	}
	return lastErr
}

// calculateBackoff computes the backoff duration with optional jitter.
func (s *Strategy) calculateBackoff(attemptsSoFar int) time.Duration {
	// Exponential backoff: initialBackoff * multiplier^attemptsSoFar
	backoff := time.Duration(float64(s.initialBackoff) * math.Pow(s.multiplier, float64(attemptsSoFar)))

	// Cap at max backoff
	if backoff > s.maxBackoff {
		backoff = s.maxBackoff
	}

	// Add jitter (±20% random variation)
	if s.addJitter {
		jitterFraction := 0.2 // ±20%
		if r, err := secureRandUnitFloat64(); err == nil {
			jitterAmount := time.Duration(float64(backoff) * jitterFraction * (2*r - 1))
			backoff = backoff + jitterAmount
		}
		if backoff < 0 {
			backoff = 0
		}
	}

	return backoff
}

// secureRandUnitFloat64 returns a cryptographically secure random float64 in [0,1).
func secureRandUnitFloat64() (float64, error) {
	var b [8]byte
	if _, err := crand.Read(b[:]); err != nil {
		return 0, err
	}
	u := binary.LittleEndian.Uint64(b[:])
	// Scale to [0,1) by dividing by 2^64
	return float64(u) / (1 << 64), nil
}

// DefaultRetryableErrors returns true for errors that should be retried:
// - 429 (Too Many Requests / Rate Limited)
// - 500+ (Server Errors)
func DefaultRetryableErrors(err *api.Error) bool {
	if err == nil {
		return false
	}
	return err.IsRateLimited() || err.IsInternalServerError()
}

// AggressiveRetryableErrors returns true for a wider set of errors:
// - 429 (Rate Limited)
// - All 5xx errors (500-599, including server errors, service unavailable, gateway timeout, etc.)
func AggressiveRetryableErrors(err *api.Error) bool {
	if err == nil {
		return false
	}
	code := err.Code
	return code == 429 || (code >= 500 && code <= 599)
}

// ConservativeRetryableErrors returns true only for transient errors:
// - 429 (Rate Limited)
// - 503 (Service Unavailable)
// - 504 (Gateway Timeout)
//
// Does not retry on 500, 502, etc. as they may indicate persistent issues,
// unlike AggressiveRetryableErrors which retries all 5xx errors.
func ConservativeRetryableErrors(err *api.Error) bool {
	if err == nil {
		return false
	}
	code := err.Code
	return code == 429 || code == 503 || code == 504
}
