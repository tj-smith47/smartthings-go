package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListLocations(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := locationListResponse{
				Items: []Location{
					{LocationID: "loc-1", Name: "Home", CountryCode: "US"},
					{LocationID: "loc-2", Name: "Office", CountryCode: "US"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		locations, err := client.ListLocations(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(locations) != 2 {
			t.Errorf("got %d locations, want 2", len(locations))
		}
		if locations[0].Name != "Home" {
			t.Errorf("locations[0].Name = %q, want %q", locations[0].Name, "Home")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(locationListResponse{Items: []Location{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		locations, err := client.ListLocations(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(locations) != 0 {
			t.Errorf("got %d locations, want 0", len(locations))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListLocations(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestClient_GetLocation(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123")
			}
			location := Location{
				LocationID:  "loc-123",
				Name:        "My Home",
				CountryCode: "US",
				TimeZoneID:  "America/New_York",
				Latitude:    40.7128,
				Longitude:   -74.0060,
			}
			json.NewEncoder(w).Encode(location)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		location, err := client.GetLocation(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if location.LocationID != "loc-123" {
			t.Errorf("LocationID = %q, want %q", location.LocationID, "loc-123")
		}
		if location.Name != "My Home" {
			t.Errorf("Name = %q, want %q", location.Name, "My Home")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetLocation(context.Background(), "")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("location not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetLocation(context.Background(), "missing-loc")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_CreateLocation(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req LocationCreate
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "New Home" {
				t.Errorf("Name = %q, want %q", req.Name, "New Home")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Location{
				LocationID:  "new-loc-123",
				Name:        req.Name,
				CountryCode: req.CountryCode,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		loc, err := client.CreateLocation(context.Background(), &LocationCreate{
			Name:        "New Home",
			CountryCode: "US",
			TimeZoneID:  "America/New_York",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if loc.LocationID != "new-loc-123" {
			t.Errorf("LocationID = %q, want %q", loc.LocationID, "new-loc-123")
		}
	})

	t.Run("nil location", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateLocation(context.Background(), nil)
		if err != ErrEmptyLocationName {
			t.Errorf("expected ErrEmptyLocationName, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateLocation(context.Background(), &LocationCreate{Name: ""})
		if err != ErrEmptyLocationName {
			t.Errorf("expected ErrEmptyLocationName, got %v", err)
		}
	})
}

func TestClient_UpdateLocation(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req LocationUpdate
			json.NewDecoder(r.Body).Decode(&req)

			json.NewEncoder(w).Encode(Location{
				LocationID: "loc-123",
				Name:       req.Name,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		loc, err := client.UpdateLocation(context.Background(), "loc-123", &LocationUpdate{
			Name: "Updated Home",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if loc.Name != "Updated Home" {
			t.Errorf("Name = %q, want %q", loc.Name, "Updated Home")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateLocation(context.Background(), "", &LocationUpdate{Name: "Test"})
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})
}

func TestClient_DeleteLocation(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteLocation(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteLocation(context.Background(), "")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})
}
