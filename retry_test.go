package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestWithRetry(t *testing.T) {
	t.Run("applies retry config", func(t *testing.T) {
		config := &RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 50 * time.Millisecond,
			MaxBackoff:     500 * time.Millisecond,
			Multiplier:     2.0,
		}
		client, err := NewClient("token", WithRetry(config))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.retryConfig == nil {
			t.Fatal("retryConfig is nil")
		}
		if client.retryConfig.MaxRetries != 3 {
			t.Errorf("MaxRetries = %d, want 3", client.retryConfig.MaxRetries)
		}
	})

	t.Run("nil config disables retry", func(t *testing.T) {
		client, err := NewClient("token", WithRetry(nil))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.retryConfig != nil {
			t.Error("retryConfig should be nil")
		}
	})
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()
	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.InitialBackoff != 100*time.Millisecond {
		t.Errorf("InitialBackoff = %v, want 100ms", config.InitialBackoff)
	}
	if config.MaxBackoff != 5*time.Second {
		t.Errorf("MaxBackoff = %v, want 5s", config.MaxBackoff)
	}
	if config.Multiplier != 2.0 {
		t.Errorf("Multiplier = %f, want 2.0", config.Multiplier)
	}
}

func TestClient_RetryOn429(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt <= 2 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(DeviceListResponse{
			Items: []Device{{DeviceID: "device-1"}},
		})
	}))
	defer server.Close()

	client, _ := NewClient("token",
		WithBaseURL(server.URL),
		WithRetry(&RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}),
	)

	devices, err := client.ListDevices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("got %d devices, want 1", len(devices))
	}
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("got %d attempts, want 3", attempts)
	}
}

func TestClient_RetryOn500(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt := atomic.AddInt32(&attempts, 1)
		if attempt == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		json.NewEncoder(w).Encode(DeviceListResponse{
			Items: []Device{{DeviceID: "device-1"}},
		})
	}))
	defer server.Close()

	client, _ := NewClient("token",
		WithBaseURL(server.URL),
		WithRetry(&RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}),
	)

	devices, err := client.ListDevices(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(devices) != 1 {
		t.Errorf("got %d devices, want 1", len(devices))
	}
	if atomic.LoadInt32(&attempts) != 2 {
		t.Errorf("got %d attempts, want 2", attempts)
	}
}

func TestClient_RetryExhausted(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, _ := NewClient("token",
		WithBaseURL(server.URL),
		WithRetry(&RetryConfig{
			MaxRetries:     2,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     50 * time.Millisecond,
			Multiplier:     2.0,
		}),
	)

	_, err := client.ListDevices(context.Background())
	if err == nil {
		t.Fatal("expected error after retries exhausted")
	}
	if !IsRateLimited(err) {
		t.Errorf("expected rate limited error, got %v", err)
	}
	// Initial attempt + 2 retries = 3 total attempts
	if atomic.LoadInt32(&attempts) != 3 {
		t.Errorf("got %d attempts, want 3", attempts)
	}
}

func TestClient_NoRetryOn400(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	defer server.Close()

	client, _ := NewClient("token",
		WithBaseURL(server.URL),
		WithRetry(&RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}),
	)

	_, err := client.ListDevices(context.Background())
	if err == nil {
		t.Fatal("expected error for bad request")
	}
	// Should not retry on 400
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("got %d attempts, want 1 (no retries on 400)", attempts)
	}
}

func TestClient_NoRetryOn401(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client, _ := NewClient("token",
		WithBaseURL(server.URL),
		WithRetry(&RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}),
	)

	_, err := client.ListDevices(context.Background())
	if err == nil {
		t.Fatal("expected error for unauthorized")
	}
	if !IsUnauthorized(err) {
		t.Errorf("expected unauthorized error, got %v", err)
	}
	// Should not retry on 401
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("got %d attempts, want 1 (no retries on 401)", attempts)
	}
}

func TestClient_NoRetryOn404(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client, _ := NewClient("token",
		WithBaseURL(server.URL),
		WithRetry(&RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 10 * time.Millisecond,
			MaxBackoff:     100 * time.Millisecond,
			Multiplier:     2.0,
		}),
	)

	_, err := client.GetDevice(context.Background(), "missing-device")
	if err == nil {
		t.Fatal("expected error for not found")
	}
	if !IsNotFound(err) {
		t.Errorf("expected not found error, got %v", err)
	}
	// Should not retry on 404
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("got %d attempts, want 1 (no retries on 404)", attempts)
	}
}

func TestClient_RetryContextCanceled(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client, _ := NewClient("token",
		WithBaseURL(server.URL),
		WithRetry(&RetryConfig{
			MaxRetries:     10,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     1 * time.Second,
			Multiplier:     2.0,
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := client.ListDevices(ctx)
	if err == nil {
		t.Fatal("expected error due to context cancellation")
	}
	// Should have stopped retrying due to context
	if atomic.LoadInt32(&attempts) > 2 {
		t.Errorf("got %d attempts, expected fewer due to context timeout", attempts)
	}
}

func TestClient_NoRetryWhenDisabled(t *testing.T) {
	var attempts int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempts, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	// Create client without retry config
	client, _ := NewClient("token", WithBaseURL(server.URL))

	_, err := client.ListDevices(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	// Should not retry when retry is disabled
	if atomic.LoadInt32(&attempts) != 1 {
		t.Errorf("got %d attempts, want 1 (retry disabled)", attempts)
	}
}
