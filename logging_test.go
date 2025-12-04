package smartthings

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestWithLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	client, err := NewClient("token", WithLogger(logger))
	if err != nil {
		t.Fatalf("NewClient failed: %v", err)
	}

	if client.logger != logger {
		t.Error("logger not set")
	}
}

func TestLoggingTransport(t *testing.T) {
	t.Run("logs successful request", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("X-RateLimit-Remaining", "99")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{}`))
		}))
		defer server.Close()

		transport := &LoggingTransport{
			Base:   http.DefaultTransport,
			Logger: logger,
		}

		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		output := buf.String()
		if !strings.Contains(output, "api_request") {
			t.Error("expected api_request log")
		}
		if !strings.Contains(output, "api_response") {
			t.Error("expected api_response log")
		}
		if !strings.Contains(output, "rate_limit_remaining") {
			t.Error("expected rate_limit_remaining in log")
		}
	})

	t.Run("logs error response", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		transport := &LoggingTransport{
			Base:   http.DefaultTransport,
			Logger: logger,
		}

		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		output := buf.String()
		if !strings.Contains(output, "ERROR") {
			t.Errorf("expected ERROR level for 500 response, got: %s", output)
		}
	})

	t.Run("logs 4xx as warning", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		transport := &LoggingTransport{
			Base:   http.DefaultTransport,
			Logger: logger,
		}

		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()

		output := buf.String()
		if !strings.Contains(output, "WARN") {
			t.Errorf("expected WARN level for 404 response, got: %s", output)
		}
	})

	t.Run("handles nil logger", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		transport := &LoggingTransport{
			Base:   http.DefaultTransport,
			Logger: nil, // nil logger
		}

		client := &http.Client{Transport: transport}
		req, _ := http.NewRequest(http.MethodGet, server.URL+"/test", nil)
		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		resp.Body.Close()
		// Should not panic
	})
}

func TestClient_LogRequest(t *testing.T) {
	t.Run("logs with logger", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		client, _ := NewClient("token", WithLogger(logger))
		client.LogRequest(context.Background(), "GET", "/devices")

		if !strings.Contains(buf.String(), "api_request") {
			t.Error("expected api_request log")
		}
		if !strings.Contains(buf.String(), "/devices") {
			t.Error("expected path in log")
		}
	})

	t.Run("no-op without logger", func(t *testing.T) {
		client, _ := NewClient("token")
		// Should not panic
		client.LogRequest(context.Background(), "GET", "/devices")
	})
}

func TestClient_LogResponse(t *testing.T) {
	t.Run("logs success response", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		client, _ := NewClient("token", WithLogger(logger))
		client.LogResponse(context.Background(), "GET", "/devices", 200, 50*time.Millisecond, nil)

		output := buf.String()
		if !strings.Contains(output, "api_response") {
			t.Error("expected api_response log")
		}
		if !strings.Contains(output, "200") {
			t.Error("expected status code in log")
		}
	})

	t.Run("logs error response", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		client, _ := NewClient("token", WithLogger(logger))
		client.LogResponse(context.Background(), "GET", "/devices", 500, 50*time.Millisecond, ErrRateLimited)

		output := buf.String()
		if !strings.Contains(output, "ERROR") {
			t.Error("expected ERROR level")
		}
		if !strings.Contains(output, "error") {
			t.Error("expected error in log")
		}
	})
}

func TestClient_LogRateLimit(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

	client, _ := NewClient("token", WithLogger(logger))
	client.LogRateLimit(context.Background(), RateLimitInfo{
		Limit:     100,
		Remaining: 50,
		Reset:     time.Now().Add(1 * time.Hour),
	})

	output := buf.String()
	if !strings.Contains(output, "rate_limit") {
		t.Error("expected rate_limit log")
	}
	if !strings.Contains(output, "50") {
		t.Error("expected remaining in log")
	}
}

func TestClient_LogDeviceCommand(t *testing.T) {
	t.Run("logs successful command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

		client, _ := NewClient("token", WithLogger(logger))
		client.LogDeviceCommand(context.Background(), "device-123", "switch", "on", nil)

		output := buf.String()
		if !strings.Contains(output, "device_command") {
			t.Error("expected device_command log")
		}
		if !strings.Contains(output, "device-123") {
			t.Error("expected device_id in log")
		}
		if !strings.Contains(output, "switch") {
			t.Error("expected capability in log")
		}
	})

	t.Run("logs failed command", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

		client, _ := NewClient("token", WithLogger(logger))
		client.LogDeviceCommand(context.Background(), "device-123", "switch", "on", ErrDeviceOffline)

		output := buf.String()
		if !strings.Contains(output, "ERROR") {
			t.Error("expected ERROR level")
		}
		if !strings.Contains(output, "error") {
			t.Error("expected error in log")
		}
	})
}

func TestLogWebhookEvent(t *testing.T) {
	t.Run("logs event", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

		event := &WebhookEvent{
			Lifecycle: LifecycleEvent,
			AppID:     "app-123",
			InstallData: &InstallData{
				InstalledApp: InstalledAppRef{
					LocationID: "loc-456",
				},
			},
		}
		LogWebhookEvent(logger, context.Background(), event, nil)

		output := buf.String()
		if !strings.Contains(output, "webhook_event") {
			t.Error("expected webhook_event log")
		}
		if !strings.Contains(output, "EVENT") {
			t.Error("expected lifecycle in log")
		}
		if !strings.Contains(output, "app-123") {
			t.Error("expected app_id in log")
		}
	})

	t.Run("logs error", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))

		event := &WebhookEvent{Lifecycle: LifecycleEvent}
		LogWebhookEvent(logger, context.Background(), event, ErrInvalidSignature)

		output := buf.String()
		if !strings.Contains(output, "ERROR") {
			t.Error("expected ERROR level")
		}
	})

	t.Run("handles nil logger", func(t *testing.T) {
		event := &WebhookEvent{Lifecycle: LifecycleEvent}
		// Should not panic
		LogWebhookEvent(nil, context.Background(), event, nil)
	})
}

func TestNewLoggingClient(t *testing.T) {
	t.Run("creates client with logging", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		client, err := NewLoggingClient("token", logger)
		if err != nil {
			t.Fatalf("NewLoggingClient failed: %v", err)
		}

		if client.logger != logger {
			t.Error("logger not set on client")
		}
	})

	t.Run("returns error for empty token", func(t *testing.T) {
		logger := slog.New(slog.NewTextHandler(nil, nil))

		_, err := NewLoggingClient("", logger)
		if err != ErrEmptyToken {
			t.Errorf("expected ErrEmptyToken, got: %v", err)
		}
	})

	t.Run("logs actual requests", func(t *testing.T) {
		var buf bytes.Buffer
		logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items":[]}`))
		}))
		defer server.Close()

		client, _ := NewLoggingClient("token", logger, WithBaseURL(server.URL))
		client.ListDevices(context.Background())

		output := buf.String()
		if !strings.Contains(output, "api_request") {
			t.Error("expected api_request log")
		}
		if !strings.Contains(output, "api_response") {
			t.Error("expected api_response log")
		}
	})
}
