package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListModes(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/modes" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/modes")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := modeListResponse{
				Items: []Mode{
					{ID: "mode-1", Name: "Home", Label: "Home"},
					{ID: "mode-2", Name: "Away", Label: "Away"},
					{ID: "mode-3", Name: "Night", Label: "Night"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		modes, err := client.ListModes(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(modes) != 3 {
			t.Errorf("got %d modes, want 3", len(modes))
		}
		if modes[0].Name != "Home" {
			t.Errorf("modes[0].Name = %q, want %q", modes[0].Name, "Home")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.ListModes(context.Background(), "")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(modeListResponse{Items: []Mode{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		modes, err := client.ListModes(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(modes) != 0 {
			t.Errorf("got %d modes, want 0", len(modes))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListModes(context.Background(), "loc-123")
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
		_, err := client.ListModes(context.Background(), "loc-123")
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}

func TestClient_GetMode(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/modes/mode-456" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/modes/mode-456")
			}
			mode := Mode{
				ID:    "mode-456",
				Name:  "Night",
				Label: "Night Mode",
			}
			json.NewEncoder(w).Encode(mode)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		mode, err := client.GetMode(context.Background(), "loc-123", "mode-456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode.ID != "mode-456" {
			t.Errorf("ID = %q, want %q", mode.ID, "mode-456")
		}
		if mode.Name != "Night" {
			t.Errorf("Name = %q, want %q", mode.Name, "Night")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetMode(context.Background(), "", "mode-123")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("empty mode ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetMode(context.Background(), "loc-123", "")
		if err != ErrEmptyModeID {
			t.Errorf("expected ErrEmptyModeID, got %v", err)
		}
	})

	t.Run("mode not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetMode(context.Background(), "loc-123", "missing-mode")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_GetCurrentMode(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/modes/current" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/modes/current")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			mode := Mode{
				ID:    "mode-home",
				Name:  "Home",
				Label: "Home",
			}
			json.NewEncoder(w).Encode(mode)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		mode, err := client.GetCurrentMode(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode.Name != "Home" {
			t.Errorf("Name = %q, want %q", mode.Name, "Home")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetCurrentMode(context.Background(), "")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("invalid"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetCurrentMode(context.Background(), "loc-123")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestClient_SetCurrentMode(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/modes/current" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/modes/current")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req setCurrentModeRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}
			if req.ModeID != "mode-night" {
				t.Errorf("modeId = %q, want %q", req.ModeID, "mode-night")
			}

			mode := Mode{
				ID:    "mode-night",
				Name:  "Night",
				Label: "Night",
			}
			json.NewEncoder(w).Encode(mode)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		mode, err := client.SetCurrentMode(context.Background(), "loc-123", "mode-night")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if mode.Name != "Night" {
			t.Errorf("Name = %q, want %q", mode.Name, "Night")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.SetCurrentMode(context.Background(), "", "mode-123")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("empty mode ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.SetCurrentMode(context.Background(), "loc-123", "")
		if err != ErrEmptyModeID {
			t.Errorf("expected ErrEmptyModeID, got %v", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.SetCurrentMode(context.Background(), "loc-123", "mode-123")
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})

	t.Run("mode not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.SetCurrentMode(context.Background(), "loc-123", "missing-mode")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}
