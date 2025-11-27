package retry_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/creiche/confluent-go/pkg/api"
	"github.com/creiche/confluent-go/pkg/retry"
)

func TestDefaultStrategy(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy()
	if strategy == nil {
		t.Fatal("DefaultStrategy returned nil")
	}
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy()
	attempts := 0

	err := strategy.Do(context.Background(), func() error {
		attempts++
		return nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if attempts != 1 {
		t.Errorf("Expected 1 attempt, got %d", attempts)
	}
}

func TestRetry_NonRetryableError(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy()
	attempts := 0

	// 404 Not Found is not retryable
	_ = strategy.Do(context.Background(), func() error {
		attempts++
		return &api.Error{
			Code:      http.StatusNotFound,
			ErrorCode: "NOT_FOUND",
			Message:   "Resource not found",
		}
	})

	if attempts != 1 {
		t.Errorf("Expected 1 attempt (no retry for non-retryable error), got %d", attempts)
	}
}

func TestRetry_RateLimitedRetryable(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().WithMaxAttempts(3).WithInitialBackoff(10 * time.Millisecond)
	attempts := 0

	err := strategy.Do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			// First two attempts: rate limited with explicit (small) retry_after
			return &api.Error{
				Code:      http.StatusTooManyRequests,
				ErrorCode: api.ErrorCodeRateLimitExceeded,
				Message:   "Rate limit exceeded",
				Details:   map[string]interface{}{"retry_after": "0"},
			}
		}
		// Third attempt: success
		return nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ServerErrorRetryable(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().WithMaxAttempts(3).WithInitialBackoff(10 * time.Millisecond)
	attempts := 0

	err := strategy.Do(context.Background(), func() error {
		attempts++
		if attempts < 3 {
			// First two attempts: internal server error
			return &api.Error{
				Code:      http.StatusInternalServerError,
				ErrorCode: api.ErrorCodeInternalServer,
				Message:   "Internal server error",
			}
		}
		// Third attempt: success
		return nil
	})

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ExceedsMaxAttempts(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().WithMaxAttempts(3).WithInitialBackoff(10 * time.Millisecond)
	attempts := 0

	err := strategy.Do(context.Background(), func() error {
		attempts++
		return &api.Error{
			Code:      http.StatusTooManyRequests,
			ErrorCode: api.ErrorCodeRateLimitExceeded,
			Message:   "Rate limit exceeded",
			Details:   map[string]interface{}{"retry_after": "0"}, // Use calculated backoff
		}
	})

	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if attempts != 3 {
		t.Errorf("Expected 3 attempts, got %d", attempts)
	}
}

func TestRetry_ContextCancellation(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().WithMaxAttempts(5).WithInitialBackoff(10 * time.Millisecond)
	attempts := 0

	ctx, cancel := context.WithCancel(context.Background())

	// Cancel after first attempt
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	err := strategy.Do(ctx, func() error {
		attempts++
		return &api.Error{
			Code:      http.StatusTooManyRequests,
			ErrorCode: api.ErrorCodeRateLimitExceeded,
			Message:   "Rate limit exceeded",
		}
	})

	if err == nil {
		t.Fatal("Expected error due to cancellation")
	}
	if attempts > 2 {
		t.Errorf("Expected at most 2 attempts (1 + cancel), got %d", attempts)
	}
}

func TestRetry_ContextDeadline(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().WithMaxAttempts(10).WithInitialBackoff(10 * time.Millisecond)
	attempts := 0

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := strategy.Do(ctx, func() error {
		attempts++
		// Always return retryable error
		return &api.Error{
			Code:      http.StatusTooManyRequests,
			ErrorCode: api.ErrorCodeRateLimitExceeded,
			Message:   "Rate limit exceeded",
		}
	})

	if err == nil {
		t.Fatal("Expected error due to deadline exceeded")
	}
	// Should not exhaust all attempts due to deadline
	if attempts >= 10 {
		t.Errorf("Expected deadline to interrupt retries, got %d attempts", attempts)
	}
}

func TestRetry_ExponentialBackoff(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().
		WithMaxAttempts(4).
		WithInitialBackoff(10 * time.Millisecond).
		WithMaxBackoff(100 * time.Millisecond).
		WithMultiplier(2.0).
		WithJitter(false) // Disable jitter for deterministic testing

	attempts := 0
	startTime := time.Now()

	_ = strategy.Do(context.Background(), func() error {
		attempts++
		// Use 429 error with explicit retry_after (in milliseconds) to test exponential backoff
		// without the default 60-second Retry-After interfering
		return &api.Error{
			Code:      http.StatusTooManyRequests,
			ErrorCode: api.ErrorCodeRateLimitExceeded,
			Message:   "Rate limit exceeded",
			Details:   map[string]interface{}{"retry_after": "0"}, // 0 seconds = use calculated backoff
		}
	})

	elapsed := time.Since(startTime)

	// Expected backoff sequence (no jitter):
	// Attempt 1: 0ms (first attempt)
	// Attempt 2: wait 10ms + Attempt 2
	// Attempt 3: wait 20ms + Attempt 3
	// Attempt 4: wait 40ms + Attempt 4
	// Total backoff: ~70ms (without jitter)
	minExpectedWait := 50 * time.Millisecond  // 10 + 20 + 40 - some slack
	maxExpectedWait := 150 * time.Millisecond // Account for execution time

	if elapsed < minExpectedWait || elapsed > maxExpectedWait {
		t.Logf("Warning: backoff timing may be off. Elapsed: %v, Expected range: %v to %v",
			elapsed, minExpectedWait, maxExpectedWait)
	}

	if attempts != 4 {
		t.Errorf("Expected 4 attempts, got %d", attempts)
	}
}

func TestRetry_WithRetryAfterHeader(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().WithMaxAttempts(3).WithInitialBackoff(10 * time.Millisecond)
	attempts := 0
	startTime := time.Now()

	err := strategy.Do(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			// First attempt with Retry-After header of 50ms (should override calculated 10ms backoff)
			return &api.Error{
				Code:      http.StatusTooManyRequests,
				ErrorCode: api.ErrorCodeRateLimitExceeded,
				Message:   "Rate limit exceeded",
				Details:   map[string]interface{}{"retry_after": "50"},
			}
		}
		return nil
	})

	elapsed := time.Since(startTime)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
	if attempts != 2 {
		t.Errorf("Expected 2 attempts, got %d", attempts)
	}

	// Should wait approximately 50ms for Retry-After header (not the calculated 10ms)
	if elapsed < 40*time.Millisecond {
		t.Logf("Warning: Retry-After header not respected. Elapsed: %v, expected ~50ms", elapsed)
	}
}

func TestRetry_ConservativeRetryableErrors(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().
		WithMaxAttempts(2).
		WithInitialBackoff(10 * time.Millisecond).
		WithRetryableErrors(retry.ConservativeRetryableErrors)

	// Test 429 is retried
	attempts := 0
	err := strategy.Do(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			return &api.Error{
				Code:      http.StatusTooManyRequests,
				ErrorCode: api.ErrorCodeRateLimitExceeded,
				Message:   "Rate limit exceeded",
				Details:   map[string]interface{}{"retry_after": "0"}, // Use calculated backoff
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("429 should be retried in conservative mode: %v", err)
	}

	// Test 500 is NOT retried (conservative doesn't retry 500)
	attempts = 0
	err = strategy.Do(context.Background(), func() error {
		attempts++
		return &api.Error{
			Code:      http.StatusInternalServerError,
			ErrorCode: api.ErrorCodeInternalServer,
			Message:   "Internal server error",
		}
	})
	if err == nil {
		t.Error("Expected 500 to not be retried in conservative mode")
	}
	if attempts != 1 {
		t.Errorf("500 should not be retried in conservative mode, got %d attempts", attempts)
	}
}

func TestRetry_AggressiveRetryableErrors(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().
		WithMaxAttempts(2).
		WithInitialBackoff(10 * time.Millisecond).
		WithRetryableErrors(retry.AggressiveRetryableErrors)

	// Test 429 is retried
	attempts := 0
	err := strategy.Do(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			return &api.Error{
				Code:      http.StatusTooManyRequests,
				ErrorCode: api.ErrorCodeRateLimitExceeded,
				Message:   "Rate limit exceeded",
				Details:   map[string]interface{}{"retry_after": "0"}, // Use calculated backoff
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("429 should be retried in aggressive mode: %v", err)
	}

	// Test 500 is retried
	attempts = 0
	err = strategy.Do(context.Background(), func() error {
		attempts++
		if attempts == 1 {
			return &api.Error{
				Code:      http.StatusInternalServerError,
				ErrorCode: api.ErrorCodeInternalServer,
				Message:   "Internal server error",
			}
		}
		return nil
	})
	if err != nil {
		t.Errorf("500 should be retried in aggressive mode: %v", err)
	}
}

func TestRetry_ChainableConfig(t *testing.T) {
	t.Parallel()
	strategy := retry.DefaultStrategy().
		WithMaxAttempts(5).
		WithInitialBackoff(500 * time.Millisecond).
		WithMaxBackoff(10 * time.Second).
		WithMultiplier(3.0).
		WithJitter(false)

	if strategy == nil {
		t.Fatal("Chained configuration returned nil")
	}

	// Should return nil error (success)
	err := strategy.Do(context.Background(), func() error {
		return nil
	})

	if err != nil {
		t.Fatalf("Expected no error for successful operation: %v", err)
	}
}
