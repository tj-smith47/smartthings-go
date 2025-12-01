package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	t.Run("default configuration", func(t *testing.T) {
		client, err := NewClient("test-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client == nil {
			t.Fatal("NewClient returned nil")
		}
		if client.token != "test-token" {
			t.Errorf("token = %q, want %q", client.token, "test-token")
		}
		if client.baseURL != DefaultBaseURL {
			t.Errorf("baseURL = %q, want %q", client.baseURL, DefaultBaseURL)
		}
		if client.httpClient == nil {
			t.Error("httpClient is nil")
		}
	})

	t.Run("with custom base URL", func(t *testing.T) {
		customURL := "https://custom.api.com"
		client, err := NewClient("token", WithBaseURL(customURL))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.baseURL != customURL {
			t.Errorf("baseURL = %q, want %q", client.baseURL, customURL)
		}
	})

	t.Run("with custom timeout", func(t *testing.T) {
		customTimeout := 5 * time.Second
		client, err := NewClient("token", WithTimeout(customTimeout))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.httpClient.Timeout != customTimeout {
			t.Errorf("timeout = %v, want %v", client.httpClient.Timeout, customTimeout)
		}
	})

	t.Run("with custom HTTP client", func(t *testing.T) {
		customHTTPClient := &http.Client{Timeout: 10 * time.Second}
		client, err := NewClient("token", WithHTTPClient(customHTTPClient))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if client.httpClient != customHTTPClient {
			t.Error("httpClient was not set correctly")
		}
	})

	t.Run("empty token returns error", func(t *testing.T) {
		client, err := NewClient("")
		if err == nil {
			t.Fatal("expected error for empty token")
		}
		if err != ErrEmptyToken {
			t.Errorf("error = %v, want ErrEmptyToken", err)
		}
		if client != nil {
			t.Error("client should be nil on error")
		}
	})
}

func TestClient_do(t *testing.T) {
	t.Run("successful GET request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Errorf("expected GET, got %s", r.Method)
			}
			if auth := r.Header.Get("Authorization"); auth != "Bearer test-token" {
				t.Errorf("Authorization header = %q, want %q", auth, "Bearer test-token")
			}
			if accept := r.Header.Get("Accept"); accept != "application/json" {
				t.Errorf("Accept header = %q, want %q", accept, "application/json")
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
		}))
		defer server.Close()

		client, _ := NewClient("test-token", WithBaseURL(server.URL))
		data, err := client.get(context.Background(), "/test")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data == nil {
			t.Fatal("data is nil")
		}
	})

	t.Run("successful POST request", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if ct := r.Header.Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type header = %q, want %q", ct, "application/json")
			}
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]string{"result": "success"})
		}))
		defer server.Close()

		client, _ := NewClient("test-token", WithBaseURL(server.URL))
		data, err := client.post(context.Background(), "/test", map[string]string{"key": "value"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if data == nil {
			t.Fatal("data is nil")
		}
	})

	t.Run("401 unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, _ := NewClient("bad-token", WithBaseURL(server.URL))
		_, err := client.get(context.Background(), "/test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got: %v", err)
		}
	})

	t.Run("404 not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.get(context.Background(), "/missing")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got: %v", err)
		}
	})

	t.Run("429 rate limited", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusTooManyRequests)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.get(context.Background(), "/test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		if !IsRateLimited(err) {
			t.Errorf("expected rate limited error, got: %v", err)
		}
	})

	t.Run("500 server error with message", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]any{
				"requestId": "req-123",
				"error": map[string]string{
					"code":    "InternalError",
					"message": "Something went wrong",
				},
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.get(context.Background(), "/test")
		if err == nil {
			t.Fatal("expected error, got nil")
		}
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("expected *APIError, got %T", err)
		}
		if apiErr.StatusCode != 500 {
			t.Errorf("StatusCode = %d, want 500", apiErr.StatusCode)
		}
		if apiErr.Message != "Something went wrong" {
			t.Errorf("Message = %q, want %q", apiErr.Message, "Something went wrong")
		}
		if apiErr.RequestID != "req-123" {
			t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "req-123")
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := client.get(ctx, "/test")
		if err == nil {
			t.Fatal("expected error due to cancelled context")
		}
	})

	t.Run("request with nil body", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Verify no Content-Type for nil body
			if ct := r.Header.Get("Content-Type"); ct != "" {
				t.Errorf("Content-Type should be empty for nil body, got %q", ct)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.do(context.Background(), http.MethodPost, "/test", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestClient_handleError(t *testing.T) {
	client, _ := NewClient("token")

	t.Run("parse error response", func(t *testing.T) {
		body := []byte(`{"requestId":"abc","error":{"code":"BadRequest","message":"Invalid input"}}`)
		err := client.handleError(400, body)
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("expected *APIError, got %T", err)
		}
		if apiErr.Message != "Invalid input" {
			t.Errorf("Message = %q, want %q", apiErr.Message, "Invalid input")
		}
		if apiErr.RequestID != "abc" {
			t.Errorf("RequestID = %q, want %q", apiErr.RequestID, "abc")
		}
	})

	t.Run("invalid JSON falls back to body", func(t *testing.T) {
		body := []byte("not json")
		err := client.handleError(400, body)
		apiErr, ok := err.(*APIError)
		if !ok {
			t.Fatalf("expected *APIError, got %T", err)
		}
		if apiErr.Message != "not json" {
			t.Errorf("Message = %q, want %q", apiErr.Message, "not json")
		}
	})
}

func TestWithTimeout_initializesClient(t *testing.T) {
	// Test that WithTimeout initializes httpClient if nil
	c := &Client{
		baseURL:    DefaultBaseURL,
		token:      "token",
		httpClient: nil, // Explicitly nil
	}

	opt := WithTimeout(5 * time.Second)
	// This should not panic
	opt(c)

	// httpClient should now be initialized with the timeout
	if c.httpClient == nil {
		t.Error("expected httpClient to be initialized")
	}
	if c.httpClient.Timeout != 5*time.Second {
		t.Errorf("expected timeout to be 5s, got %v", c.httpClient.Timeout)
	}
}
