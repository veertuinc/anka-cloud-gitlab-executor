package ankacloud

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/veertuinc/anka-cloud-gitlab-executor/internal/gitlab"
)

func TestCustomHeaders(t *testing.T) {
	const (
		headerName = "fake-header"
		headerVal  = "fake-value"
		path       = "/fakepath"
	)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(headerName) != headerVal {
			t.Errorf("expected header %q to be %q, got %q", headerName, headerVal, r.Header.Get(headerName))
		}

		json.NewEncoder(w).Encode(response{Status: statusOK})
	}))
	defer server.Close()

	client, err := NewAPIClient(APIClientConfig{
		BaseURL: server.URL,
		CustomHttpHeaders: map[string]string{
			headerName: headerVal,
		},
	})
	if err != nil {
		t.Error(err)
	}

	if _, err = client.Get(context.Background(), path, nil); err != nil {
		t.Error(err)
	}

	if _, err = client.Post(context.Background(), path, nil); err != nil {
		t.Error(err)
	}

	if _, err = client.Delete(context.Background(), path, nil); err != nil {
		t.Error(err)
	}
}

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "regular error",
			err:      errors.New("some error"),
			expected: false,
		},
		{
			name:     "transient error",
			err:      gitlab.TransientError(errors.New("timeout")),
			expected: true,
		},
		{
			name:     "error with deadline exceeded",
			err:      errors.New("context deadline exceeded"),
			expected: true,
		},
		{
			name:     "error with Client.Timeout",
			err:      errors.New("Client.Timeout exceeded while awaiting headers"),
			expected: true,
		},
		{
			name:     "wrapped deadline exceeded",
			err:      errors.New("failed to terminate: context deadline exceeded (Client.Timeout exceeded)"),
			expected: true,
		},
		{
			name:     "url.Error with timeout",
			err:      &url.Error{Op: "Get", URL: "http://test", Err: errors.New("timeout")},
			expected: false, // url.Error.Timeout() returns false for generic errors
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError(%v) = %v, expected %v", tt.err, result, tt.expected)
			}
		})
	}
}

func TestWithRetry_Success(t *testing.T) {
	callCount := 0
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
	}

	result, err := WithRetry(context.Background(), config, func() (string, error) {
		callCount++
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %q", result)
	}
	if callCount != 1 {
		t.Errorf("expected 1 call, got %d", callCount)
	}
}

func TestWithRetry_NonRetryableError(t *testing.T) {
	callCount := 0
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
	}

	_, err := WithRetry(context.Background(), config, func() (string, error) {
		callCount++
		return "", errors.New("permanent error")
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 1 {
		t.Errorf("expected 1 call (no retry for non-retryable error), got %d", callCount)
	}
}

func TestWithRetry_RetryableError_EventualSuccess(t *testing.T) {
	callCount := 0
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
	}

	result, err := WithRetry(context.Background(), config, func() (string, error) {
		callCount++
		if callCount < 3 {
			return "", gitlab.TransientError(errors.New("timeout"))
		}
		return "success", nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if result != "success" {
		t.Errorf("expected 'success', got %q", result)
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
}

func TestWithRetry_RetryableError_AllAttemptsFail(t *testing.T) {
	callCount := 0
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
	}

	_, err := WithRetry(context.Background(), config, func() (string, error) {
		callCount++
		return "", gitlab.TransientError(errors.New("timeout"))
	})

	if err == nil {
		t.Error("expected error, got nil")
	}
	if callCount != 3 {
		t.Errorf("expected 3 calls, got %d", callCount)
	}
	if !errors.Is(err, gitlab.ErrTransient) {
		t.Errorf("expected error to wrap ErrTransient, got %v", err)
	}
}

func TestWithRetry_ExponentialBackoff(t *testing.T) {
	callCount := 0
	var callTimes []time.Time
	config := RetryConfig{
		MaxAttempts:  4,
		InitialDelay: 50 * time.Millisecond,
		MaxDelay:     150 * time.Millisecond,
	}

	start := time.Now()
	_, _ = WithRetry(context.Background(), config, func() (string, error) {
		callTimes = append(callTimes, time.Now())
		callCount++
		return "", gitlab.TransientError(errors.New("timeout"))
	})

	if callCount != 4 {
		t.Errorf("expected 4 calls, got %d", callCount)
	}

	// Verify exponential backoff timing
	// Expected delays: 50ms, 100ms, 150ms (capped)
	// Total minimum time: 50 + 100 + 150 = 300ms
	elapsed := time.Since(start)
	minExpected := 250 * time.Millisecond // Allow some tolerance
	if elapsed < minExpected {
		t.Errorf("expected at least %v elapsed, got %v", minExpected, elapsed)
	}

	// Verify delay between calls increases (with tolerance for timing)
	if len(callTimes) >= 3 {
		delay1 := callTimes[1].Sub(callTimes[0])
		delay2 := callTimes[2].Sub(callTimes[1])
		// Second delay should be roughly double the first (with tolerance)
		if delay2 < delay1 {
			t.Errorf("expected exponential backoff: delay2 (%v) should be >= delay1 (%v)", delay2, delay1)
		}
	}
}

func TestWithRetry_ContextCancellation(t *testing.T) {
	callCount := 0
	config := RetryConfig{
		MaxAttempts:  5,
		InitialDelay: 100 * time.Millisecond,
		MaxDelay:     1 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := WithRetry(ctx, config, func() (string, error) {
		callCount++
		return "", gitlab.TransientError(errors.New("timeout"))
	})

	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled error, got %v", err)
	}
	// Should have been cancelled before all retries completed
	if callCount >= 5 {
		t.Errorf("expected fewer than 5 calls due to cancellation, got %d", callCount)
	}
}

func TestWithRetryNoResult(t *testing.T) {
	callCount := 0
	config := RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 10 * time.Millisecond,
		MaxDelay:     100 * time.Millisecond,
	}

	err := WithRetryNoResult(context.Background(), config, func() error {
		callCount++
		if callCount < 2 {
			return gitlab.TransientError(errors.New("timeout"))
		}
		return nil
	})

	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls, got %d", callCount)
	}
}
