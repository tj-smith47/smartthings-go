package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListDevicesWithOptions(t *testing.T) {
	t.Run("successful response with pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices")
			}
			if r.URL.Query().Get("max") != "10" {
				t.Errorf("max query = %q, want %q", r.URL.Query().Get("max"), "10")
			}
			// Note: page=0 is not sent because it's the default (only page > 0 is sent)

			resp := PagedDevices{
				Items: []Device{
					{DeviceID: "device-1", Label: "Device 1"},
					{DeviceID: "device-2", Label: "Device 2"},
				},
				PageInfo: PageInfo{
					TotalResults: 25,
					TotalPages:   3,
					CurrentPage:  0,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		result, err := client.ListDevicesWithOptions(context.Background(), &ListDevicesOptions{
			Max:  10,
			Page: 0,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 2 {
			t.Errorf("got %d items, want 2", len(result.Items))
		}
		if result.PageInfo.TotalResults != 25 {
			t.Errorf("TotalResults = %d, want 25", result.PageInfo.TotalResults)
		}
		if result.PageInfo.TotalPages != 3 {
			t.Errorf("TotalPages = %d, want 3", result.PageInfo.TotalPages)
		}
	})

	t.Run("with capability filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			caps := r.URL.Query()["capability"]
			if len(caps) != 2 {
				t.Errorf("got %d capability params, want 2", len(caps))
			}

			resp := PagedDevices{
				Items: []Device{
					{DeviceID: "device-1", Label: "Switch"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		result, err := client.ListDevicesWithOptions(context.Background(), &ListDevicesOptions{
			Capability: []string{"switch", "light"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("got %d items, want 1", len(result.Items))
		}
	})

	t.Run("with location filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			locs := r.URL.Query()["locationId"]
			if len(locs) != 1 || locs[0] != "loc-123" {
				t.Errorf("locationId = %v, want [loc-123]", locs)
			}

			resp := PagedDevices{
				Items: []Device{
					{DeviceID: "device-1", Label: "Device 1", LocationID: "loc-123"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		result, err := client.ListDevicesWithOptions(context.Background(), &ListDevicesOptions{
			LocationID: []string{"loc-123"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("got %d items, want 1", len(result.Items))
		}
	})

	t.Run("with device ID filter", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ids := r.URL.Query()["deviceId"]
			if len(ids) != 2 {
				t.Errorf("got %d deviceId params, want 2", len(ids))
			}

			resp := PagedDevices{
				Items: []Device{
					{DeviceID: "device-1"},
					{DeviceID: "device-2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		result, err := client.ListDevicesWithOptions(context.Background(), &ListDevicesOptions{
			DeviceID: []string{"device-1", "device-2"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 2 {
			t.Errorf("got %d items, want 2", len(result.Items))
		}
	})

	t.Run("with nil options", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := PagedDevices{
				Items: []Device{
					{DeviceID: "device-1"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		result, err := client.ListDevicesWithOptions(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Items) != 1 {
			t.Errorf("got %d items, want 1", len(result.Items))
		}
	})

	t.Run("with links for pagination", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := PagedDevices{
				Items: []Device{
					{DeviceID: "device-1"},
				},
				Links: Links{
					Next:     "/devices?page=1",
					Previous: "",
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		result, err := client.ListDevicesWithOptions(context.Background(), nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.Links.Next != "/devices?page=1" {
			t.Errorf("Links.Next = %q, want %q", result.Links.Next, "/devices?page=1")
		}
	})
}

func TestClient_ListAllDevices(t *testing.T) {
	t.Run("single page", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := PagedDevices{
				Items: []Device{
					{DeviceID: "device-1"},
					{DeviceID: "device-2"},
				},
				PageInfo: PageInfo{
					TotalResults: 2,
					TotalPages:   1,
					CurrentPage:  0,
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		devices, err := client.ListAllDevices(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(devices) != 2 {
			t.Errorf("got %d devices, want 2", len(devices))
		}
	})

	t.Run("multiple pages", func(t *testing.T) {
		pageNum := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var resp PagedDevices
			switch pageNum {
			case 0:
				resp = PagedDevices{
					Items: []Device{
						{DeviceID: "device-1"},
						{DeviceID: "device-2"},
					},
					Links: Links{Next: "/devices?page=1"},
				}
			case 1:
				resp = PagedDevices{
					Items: []Device{
						{DeviceID: "device-3"},
						{DeviceID: "device-4"},
					},
					Links: Links{Next: "/devices?page=2"},
				}
			case 2:
				resp = PagedDevices{
					Items: []Device{
						{DeviceID: "device-5"},
					},
					Links: Links{Next: ""},
				}
			}
			pageNum++
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		devices, err := client.ListAllDevices(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(devices) != 5 {
			t.Errorf("got %d devices, want 5", len(devices))
		}
	})

	t.Run("empty response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := PagedDevices{
				Items: []Device{},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		devices, err := client.ListAllDevices(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(devices) != 0 {
			t.Errorf("got %d devices, want 0", len(devices))
		}
	})
}

func TestClient_DeleteDevice(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteDevice(context.Background(), "device-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteDevice(context.Background(), "")
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})
}

func TestClient_UpdateDevice(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req DeviceUpdate
			json.NewDecoder(r.Body).Decode(&req)
			if req.Label != "New Label" {
				t.Errorf("Label = %q, want %q", req.Label, "New Label")
			}

			json.NewEncoder(w).Encode(Device{
				DeviceID: "device-123",
				Label:    req.Label,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		device, err := client.UpdateDevice(context.Background(), "device-123", &DeviceUpdate{
			Label: "New Label",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if device.Label != "New Label" {
			t.Errorf("Label = %q, want %q", device.Label, "New Label")
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateDevice(context.Background(), "", &DeviceUpdate{Label: "Test"})
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})
}

func TestClient_GetDeviceHealth(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123/health" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123/health")
			}
			health := DeviceHealth{
				DeviceID: "device-123",
				State:    "ONLINE",
			}
			json.NewEncoder(w).Encode(health)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		health, err := client.GetDeviceHealth(context.Background(), "device-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if health.DeviceID != "device-123" {
			t.Errorf("DeviceID = %q, want %q", health.DeviceID, "device-123")
		}
		if health.State != "ONLINE" {
			t.Errorf("State = %q, want %q", health.State, "ONLINE")
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetDeviceHealth(context.Background(), "")
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})

	t.Run("offline device", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			health := DeviceHealth{
				DeviceID: "device-123",
				State:    "OFFLINE",
			}
			json.NewEncoder(w).Encode(health)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		health, err := client.GetDeviceHealth(context.Background(), "device-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if health.State != "OFFLINE" {
			t.Errorf("State = %q, want %q", health.State, "OFFLINE")
		}
	})
}
