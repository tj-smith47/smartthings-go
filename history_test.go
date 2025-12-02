package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestClient_GetDeviceEvents(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123/events" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123/events")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := PagedEvents{
				Items: []DeviceEvent{
					{
						DeviceID:    "device-123",
						ComponentID: "main",
						Capability:  "switch",
						Attribute:   "switch",
						Value:       "on",
						StateChange: true,
						Timestamp:   time.Now(),
					},
					{
						DeviceID:    "device-123",
						ComponentID: "main",
						Capability:  "switch",
						Attribute:   "switch",
						Value:       "off",
						StateChange: true,
						Timestamp:   time.Now().Add(-time.Hour),
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		events, err := client.GetDeviceEvents(context.Background(), "device-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events.Items) != 2 {
			t.Errorf("got %d events, want 2", len(events.Items))
		}
		if events.Items[0].Capability != "switch" {
			t.Errorf("Capability = %q, want %q", events.Items[0].Capability, "switch")
		}
	})

	t.Run("with options", func(t *testing.T) {
		before := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
		after := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("max") != "50" {
				t.Errorf("max = %q, want %q", r.URL.Query().Get("max"), "50")
			}
			if r.URL.Query().Get("page") != "2" {
				t.Errorf("page = %q, want %q", r.URL.Query().Get("page"), "2")
			}
			if r.URL.Query().Get("before") == "" {
				t.Error("expected before query param")
			}
			if r.URL.Query().Get("after") == "" {
				t.Error("expected after query param")
			}
			json.NewEncoder(w).Encode(PagedEvents{Items: []DeviceEvent{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDeviceEvents(context.Background(), "device-123", &HistoryOptions{
			Before: &before,
			After:  &after,
			Max:    50,
			Page:   2,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetDeviceEvents(context.Background(), "", nil)
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(PagedEvents{Items: []DeviceEvent{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		events, err := client.GetDeviceEvents(context.Background(), "device-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(events.Items) != 0 {
			t.Errorf("got %d events, want 0", len(events.Items))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDeviceEvents(context.Background(), "device-123", nil)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("device not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDeviceEvents(context.Background(), "missing-device", nil)
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})

	t.Run("with pagination info", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := PagedEvents{
				Items: []DeviceEvent{{DeviceID: "device-123", Capability: "switch"}},
				Links: Links{Next: "/devices/device-123/events?page=2"},
				PageInfo: PageInfo{
					TotalResults: 100,
					TotalPages:   5,
					CurrentPage:  1,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		events, err := client.GetDeviceEvents(context.Background(), "device-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if events.PageInfo.TotalResults != 100 {
			t.Errorf("TotalResults = %d, want 100", events.PageInfo.TotalResults)
		}
		if events.Links.Next == "" {
			t.Error("expected Next link")
		}
	})
}

func TestClient_GetDeviceStates(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123/states" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123/states")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := PagedStates{
				Items: []DeviceState{
					{
						ComponentID: "main",
						Capability:  "temperatureMeasurement",
						Attribute:   "temperature",
						Value:       72.5,
						Unit:        "F",
						Timestamp:   time.Now(),
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		states, err := client.GetDeviceStates(context.Background(), "device-123", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(states.Items) != 1 {
			t.Errorf("got %d states, want 1", len(states.Items))
		}
		if states.Items[0].Capability != "temperatureMeasurement" {
			t.Errorf("Capability = %q, want %q", states.Items[0].Capability, "temperatureMeasurement")
		}
	})

	t.Run("with options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("max") != "100" {
				t.Errorf("max = %q, want %q", r.URL.Query().Get("max"), "100")
			}
			json.NewEncoder(w).Encode(PagedStates{Items: []DeviceState{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDeviceStates(context.Background(), "device-123", &HistoryOptions{
			Max: 100,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetDeviceStates(context.Background(), "", nil)
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDeviceStates(context.Background(), "device-123", nil)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDeviceStates(context.Background(), "device-123", nil)
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}

func TestBuildHistoryQueryParams(t *testing.T) {
	t.Run("nil options", func(t *testing.T) {
		result := buildHistoryQueryParams(nil)
		if result != "" {
			t.Errorf("got %q, want empty string", result)
		}
	})

	t.Run("empty options", func(t *testing.T) {
		result := buildHistoryQueryParams(&HistoryOptions{})
		if result != "" {
			t.Errorf("got %q, want empty string", result)
		}
	})

	t.Run("with max only", func(t *testing.T) {
		result := buildHistoryQueryParams(&HistoryOptions{Max: 50})
		if result != "?max=50" {
			t.Errorf("got %q, want %q", result, "?max=50")
		}
	})

	t.Run("with all options", func(t *testing.T) {
		before := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)
		after := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
		result := buildHistoryQueryParams(&HistoryOptions{
			Before: &before,
			After:  &after,
			Max:    100,
			Page:   3,
		})
		// Query params order may vary, just check it starts with ?
		if len(result) < 10 || result[0] != '?' {
			t.Errorf("expected query string, got %q", result)
		}
	})
}
