package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestDefaultBatchConfig(t *testing.T) {
	cfg := DefaultBatchConfig()
	if cfg.MaxConcurrent != 10 {
		t.Errorf("expected MaxConcurrent=10, got %d", cfg.MaxConcurrent)
	}
	if cfg.StopOnError {
		t.Error("expected StopOnError=false")
	}
}

func TestClient_ExecuteCommandsBatch(t *testing.T) {
	t.Run("empty batch returns nil", func(t *testing.T) {
		client, _ := NewClient("token")
		results := client.ExecuteCommandsBatch(context.Background(), nil, nil)
		if results != nil {
			t.Error("expected nil for empty batch")
		}
	})

	t.Run("successful batch execution", func(t *testing.T) {
		var callCount atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount.Add(1)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		batch := []BatchCommand{
			{DeviceID: "device1", Commands: []Command{{Capability: "switch", Command: "on"}}},
			{DeviceID: "device2", Commands: []Command{{Capability: "switch", Command: "off"}}},
			{DeviceID: "device3", Commands: []Command{{Capability: "switch", Command: "on"}}},
		}

		results := client.ExecuteCommandsBatch(context.Background(), batch, nil)

		if len(results) != 3 {
			t.Fatalf("expected 3 results, got %d", len(results))
		}
		for i, r := range results {
			if r.Error != nil {
				t.Errorf("result[%d] unexpected error: %v", i, r.Error)
			}
			if r.DeviceID != batch[i].DeviceID {
				t.Errorf("result[%d] deviceID mismatch", i)
			}
		}
		if callCount.Load() != 3 {
			t.Errorf("expected 3 API calls, got %d", callCount.Load())
		}
	})

	t.Run("respects max concurrency", func(t *testing.T) {
		var concurrent atomic.Int32
		var maxConcurrent atomic.Int32

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			c := concurrent.Add(1)
			// Track max concurrency seen
			for {
				old := maxConcurrent.Load()
				if c <= old || maxConcurrent.CompareAndSwap(old, c) {
					break
				}
			}
			time.Sleep(50 * time.Millisecond) // Simulate work
			concurrent.Add(-1)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))

		// Create 10 commands
		batch := make([]BatchCommand, 10)
		for i := range batch {
			batch[i] = BatchCommand{
				DeviceID: "device" + string(rune('0'+i)),
				Commands: []Command{{Capability: "switch", Command: "on"}},
			}
		}

		// Limit to 3 concurrent
		cfg := &BatchConfig{MaxConcurrent: 3}
		client.ExecuteCommandsBatch(context.Background(), batch, cfg)

		if maxConcurrent.Load() > 3 {
			t.Errorf("exceeded max concurrency: %d > 3", maxConcurrent.Load())
		}
	})

	t.Run("handles errors without stopping", func(t *testing.T) {
		var callCount atomic.Int32
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount.Add(1)
			// Fail the second device
			if r.URL.Path == "/devices/device2/commands" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":{"message":"not found"}}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		batch := []BatchCommand{
			{DeviceID: "device1", Commands: []Command{{Capability: "switch", Command: "on"}}},
			{DeviceID: "device2", Commands: []Command{{Capability: "switch", Command: "on"}}},
			{DeviceID: "device3", Commands: []Command{{Capability: "switch", Command: "on"}}},
		}

		results := client.ExecuteCommandsBatch(context.Background(), batch, &BatchConfig{StopOnError: false})

		// All should be processed
		if callCount.Load() != 3 {
			t.Errorf("expected 3 API calls, got %d", callCount.Load())
		}

		// Check results
		if results[0].Error != nil {
			t.Error("device1 should succeed")
		}
		if results[1].Error == nil {
			t.Error("device2 should fail")
		}
		if results[2].Error != nil {
			t.Error("device3 should succeed")
		}
	})

	t.Run("stop on error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// All requests fail
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error":{"message":"server error"}}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		batch := make([]BatchCommand, 10)
		for i := range batch {
			batch[i] = BatchCommand{
				DeviceID: "device" + string(rune('0'+i)),
				Commands: []Command{{Capability: "switch", Command: "on"}},
			}
		}

		// Use single concurrency to ensure deterministic order
		cfg := &BatchConfig{MaxConcurrent: 1, StopOnError: true}
		results := client.ExecuteCommandsBatch(context.Background(), batch, cfg)

		// First should have real error, rest should be canceled
		if results[0].Error == nil {
			t.Error("first result should have error")
		}
		// Some subsequent results should be canceled
		canceledCount := 0
		for _, r := range results[1:] {
			if r.Error == context.Canceled {
				canceledCount++
			}
		}
		if canceledCount == 0 {
			t.Error("expected some results to be canceled")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		batch := make([]BatchCommand, 5)
		for i := range batch {
			batch[i] = BatchCommand{
				DeviceID: "device" + string(rune('0'+i)),
				Commands: []Command{{Capability: "switch", Command: "on"}},
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		results := client.ExecuteCommandsBatch(ctx, batch, &BatchConfig{MaxConcurrent: 1})

		// Some should have context errors
		contextErrors := 0
		for _, r := range results {
			if r.Error == context.DeadlineExceeded || r.Error == context.Canceled {
				contextErrors++
			}
		}
		if contextErrors == 0 {
			t.Error("expected some context errors")
		}
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		batch := []BatchCommand{
			{DeviceID: "device1", Commands: []Command{{Capability: "switch", Command: "on"}}},
		}

		results := client.ExecuteCommandsBatch(context.Background(), batch, nil)
		if len(results) != 1 || results[0].Error != nil {
			t.Error("batch should succeed with nil config")
		}
	})

	t.Run("zero max concurrent uses default", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		batch := []BatchCommand{
			{DeviceID: "device1", Commands: []Command{{Capability: "switch", Command: "on"}}},
		}

		cfg := &BatchConfig{MaxConcurrent: 0}
		results := client.ExecuteCommandsBatch(context.Background(), batch, cfg)
		if len(results) != 1 || results[0].Error != nil {
			t.Error("batch should succeed with zero max concurrent")
		}
	})
}

func TestClient_ExecuteCommandBatch(t *testing.T) {
	var callCount atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount.Add(1)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{}`))
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	cmd := Command{Capability: "switch", Command: "on"}

	results := client.ExecuteCommandBatch(context.Background(),
		[]string{"device1", "device2", "device3"},
		cmd, nil)

	if len(results) != 3 {
		t.Fatalf("expected 3 results, got %d", len(results))
	}
	if callCount.Load() != 3 {
		t.Errorf("expected 3 API calls, got %d", callCount.Load())
	}
}

func TestClient_GetDeviceStatusBatch(t *testing.T) {
	t.Run("empty list returns nil", func(t *testing.T) {
		client, _ := NewClient("token")
		results := client.GetDeviceStatusBatch(context.Background(), nil, nil)
		if results != nil {
			t.Error("expected nil for empty list")
		}
	})

	t.Run("successful batch fetch", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"components":{"main":{"switch":{"switch":{"value":"on"}}}}}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		results := client.GetDeviceStatusBatch(context.Background(),
			[]string{"device1", "device2"},
			nil)

		if len(results) != 2 {
			t.Fatalf("expected 2 results, got %d", len(results))
		}
		for i, r := range results {
			if r.Error != nil {
				t.Errorf("result[%d] unexpected error: %v", i, r.Error)
			}
			if r.Components == nil {
				t.Errorf("result[%d] missing components", i)
			}
		}
	})

	t.Run("handles mixed success and failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path == "/devices/device2/status" {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`{"error":{"message":"not found"}}`))
				return
			}
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"components":{"main":{}}}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		results := client.GetDeviceStatusBatch(context.Background(),
			[]string{"device1", "device2", "device3"},
			nil)

		if results[0].Error != nil {
			t.Error("device1 should succeed")
		}
		if results[1].Error == nil {
			t.Error("device2 should fail")
		}
		if results[2].Error != nil {
			t.Error("device3 should succeed")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"components":{}}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		results := client.GetDeviceStatusBatch(ctx,
			[]string{"device1", "device2", "device3"},
			&BatchConfig{MaxConcurrent: 1})

		contextErrors := 0
		for _, r := range results {
			if r.Error == context.DeadlineExceeded || r.Error == context.Canceled {
				contextErrors++
			}
		}
		if contextErrors == 0 {
			t.Error("expected context errors")
		}
	})

	t.Run("nil config uses defaults", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"components":{}}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		results := client.GetDeviceStatusBatch(context.Background(), []string{"device1"}, nil)
		if len(results) != 1 || results[0].Error != nil {
			t.Error("should succeed with nil config")
		}
	})

	t.Run("zero max concurrent uses default", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"components":{}}`))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		cfg := &BatchConfig{MaxConcurrent: 0}
		results := client.GetDeviceStatusBatch(context.Background(), []string{"device1"}, cfg)
		if len(results) != 1 || results[0].Error != nil {
			t.Error("should succeed with zero max concurrent")
		}
	})
}
