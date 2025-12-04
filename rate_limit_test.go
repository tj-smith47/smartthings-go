package smartthings

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitError(t *testing.T) {
	t.Run("error message with retry-after", func(t *testing.T) {
		err := &RateLimitError{RetryAfter: 30 * time.Second}
		if err.Error() != "smartthings: rate limited (retry after 30s)" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	t.Run("error message without retry-after", func(t *testing.T) {
		err := &RateLimitError{}
		if err.Error() != "smartthings: rate limited" {
			t.Errorf("unexpected error message: %s", err.Error())
		}
	})

	t.Run("Is matches ErrRateLimited", func(t *testing.T) {
		err := &RateLimitError{}
		if !errors.Is(err, ErrRateLimited) {
			t.Error("expected Is to match ErrRateLimited")
		}
	})
}

func TestParseRetryAfter(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected time.Duration
	}{
		{"empty", "", 0},
		{"seconds", "120", 120 * time.Second},
		{"zero seconds", "0", 0},
		{"negative", "-5", 0},
		{"invalid", "abc", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseRetryAfter(tt.value)
			if got != tt.expected {
				t.Errorf("parseRetryAfter(%q) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestClient_WaitForRateLimit(t *testing.T) {
	t.Run("returns immediately when no rate limit info", func(t *testing.T) {
		client, _ := NewClient("token")
		start := time.Now()
		err := client.WaitForRateLimit(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) > 100*time.Millisecond {
			t.Error("should return immediately")
		}
	})

	t.Run("returns immediately when reset time has passed", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{
			Reset: time.Now().Add(-1 * time.Hour), // In the past
		}
		client.rateLimitMu.Unlock()

		start := time.Now()
		err := client.WaitForRateLimit(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) > 100*time.Millisecond {
			t.Error("should return immediately")
		}
	})

	t.Run("waits until reset time", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{
			Reset: time.Now().Add(100 * time.Millisecond),
		}
		client.rateLimitMu.Unlock()

		start := time.Now()
		err := client.WaitForRateLimit(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) < 90*time.Millisecond {
			t.Error("should wait for reset time")
		}
	})

	t.Run("respects context cancellation", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{
			Reset: time.Now().Add(10 * time.Second),
		}
		client.rateLimitMu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := client.WaitForRateLimit(ctx)
		if err != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got: %v", err)
		}
	})
}

func TestClient_WaitForRateLimitErr(t *testing.T) {
	t.Run("returns immediately for non-RateLimitError", func(t *testing.T) {
		client, _ := NewClient("token")
		start := time.Now()
		err := client.WaitForRateLimitErr(context.Background(), errors.New("other error"))
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) > 100*time.Millisecond {
			t.Error("should return immediately")
		}
	})

	t.Run("waits for retry-after duration", func(t *testing.T) {
		client, _ := NewClient("token")
		rle := &RateLimitError{RetryAfter: 100 * time.Millisecond}

		start := time.Now()
		err := client.WaitForRateLimitErr(context.Background(), rle)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) < 90*time.Millisecond {
			t.Error("should wait for retry-after")
		}
	})

	t.Run("falls back to rate limit info when no retry-after", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{
			Reset: time.Now().Add(100 * time.Millisecond),
		}
		client.rateLimitMu.Unlock()

		rle := &RateLimitError{} // No RetryAfter

		start := time.Now()
		err := client.WaitForRateLimitErr(context.Background(), rle)
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) < 90*time.Millisecond {
			t.Error("should wait for reset time")
		}
	})
}

func TestClient_ShouldThrottle(t *testing.T) {
	t.Run("returns false when no rate limit info", func(t *testing.T) {
		client, _ := NewClient("token")
		if client.ShouldThrottle(10) {
			t.Error("should not throttle without rate limit info")
		}
	})

	t.Run("returns false when remaining above threshold", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Remaining: 50}
		client.rateLimitMu.Unlock()

		if client.ShouldThrottle(10) {
			t.Error("should not throttle when remaining > threshold")
		}
	})

	t.Run("returns true when remaining below threshold", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Remaining: 5}
		client.rateLimitMu.Unlock()

		if !client.ShouldThrottle(10) {
			t.Error("should throttle when remaining < threshold")
		}
	})

	t.Run("returns true when remaining equals threshold minus one", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Remaining: 9}
		client.rateLimitMu.Unlock()

		if !client.ShouldThrottle(10) {
			t.Error("should throttle when remaining < threshold")
		}
	})
}

func TestClient_RemainingRequests(t *testing.T) {
	t.Run("returns -1 when no rate limit info", func(t *testing.T) {
		client, _ := NewClient("token")
		if client.RemainingRequests() != -1 {
			t.Error("should return -1")
		}
	})

	t.Run("returns remaining from rate limit info", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Remaining: 42}
		client.rateLimitMu.Unlock()

		if client.RemainingRequests() != 42 {
			t.Errorf("expected 42, got %d", client.RemainingRequests())
		}
	})
}

func TestClient_RateLimitResetTime(t *testing.T) {
	t.Run("returns zero time when no rate limit info", func(t *testing.T) {
		client, _ := NewClient("token")
		if !client.RateLimitResetTime().IsZero() {
			t.Error("should return zero time")
		}
	})

	t.Run("returns reset time from rate limit info", func(t *testing.T) {
		client, _ := NewClient("token")
		resetTime := time.Now().Add(1 * time.Hour)
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Reset: resetTime}
		client.rateLimitMu.Unlock()

		if !client.RateLimitResetTime().Equal(resetTime) {
			t.Error("should return reset time")
		}
	})
}

func TestRateLimitThrottler(t *testing.T) {
	t.Run("Wait returns immediately when not throttling", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Remaining: 100}
		client.rateLimitMu.Unlock()

		throttler := NewRateLimitThrottler(client, 10, 100*time.Millisecond)

		start := time.Now()
		err := throttler.Wait(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) > 50*time.Millisecond {
			t.Error("should return immediately")
		}
	})

	t.Run("Wait delays when throttling", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Remaining: 5}
		client.rateLimitMu.Unlock()

		throttler := NewRateLimitThrottler(client, 10, 100*time.Millisecond)

		start := time.Now()
		err := throttler.Wait(context.Background())
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if time.Since(start) < 90*time.Millisecond {
			t.Error("should delay when throttling")
		}
	})

	t.Run("Wait respects context cancellation", func(t *testing.T) {
		client, _ := NewClient("token")
		client.rateLimitMu.Lock()
		client.lastRateLimit = &RateLimitInfo{Remaining: 5}
		client.rateLimitMu.Unlock()

		throttler := NewRateLimitThrottler(client, 10, 1*time.Second)

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		err := throttler.Wait(ctx)
		if err != context.DeadlineExceeded {
			t.Errorf("expected DeadlineExceeded, got: %v", err)
		}
	})
}

func TestClient_RateLimitErrorFromAPI(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Retry-After", "60")
		w.Header().Set("X-RateLimit-Limit", "100")
		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error":{"message":"rate limited"}}`))
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))

	_, err := client.ListDevices(context.Background())

	// Check that we get a RateLimitError
	var rle *RateLimitError
	if !errors.As(err, &rle) {
		t.Fatalf("expected RateLimitError, got: %T", err)
	}

	// Check RetryAfter is parsed
	if rle.RetryAfter != 60*time.Second {
		t.Errorf("expected RetryAfter=60s, got %v", rle.RetryAfter)
	}

	// Check that errors.Is still works
	if !errors.Is(err, ErrRateLimited) {
		t.Error("expected errors.Is to match ErrRateLimited")
	}
}
