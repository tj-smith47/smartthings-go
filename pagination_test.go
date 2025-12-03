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

// Iterator tests

func TestClient_DevicesIterator(t *testing.T) {
	t.Run("iterates all devices", func(t *testing.T) {
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
					},
					Links: Links{Next: ""},
				}
			}
			pageNum++
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var devices []Device
		for device, err := range client.Devices(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			devices = append(devices, device)
		}
		if len(devices) != 3 {
			t.Errorf("got %d devices, want 3", len(devices))
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		client, _ := NewClient("token")
		var gotErr error
		for _, err := range client.Devices(ctx) {
			gotErr = err
			break
		}
		if gotErr != context.Canceled {
			t.Errorf("expected context.Canceled, got %v", gotErr)
		}
	})
}

func TestClient_LocationsIterator(t *testing.T) {
	t.Run("iterates all locations", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Location `json:"items"`
			}{
				Items: []Location{
					{LocationID: "loc-1", Name: "Home"},
					{LocationID: "loc-2", Name: "Office"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var locations []Location
		for loc, err := range client.Locations(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			locations = append(locations, loc)
		}
		if len(locations) != 2 {
			t.Errorf("got %d locations, want 2", len(locations))
		}
	})
}

func TestClient_RoomsIterator(t *testing.T) {
	t.Run("iterates all rooms", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Room `json:"items"`
			}{
				Items: []Room{
					{RoomID: "room-1", Name: "Living Room"},
					{RoomID: "room-2", Name: "Bedroom"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var rooms []Room
		for room, err := range client.Rooms(context.Background(), "loc-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			rooms = append(rooms, room)
		}
		if len(rooms) != 2 {
			t.Errorf("got %d rooms, want 2", len(rooms))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.Rooms(context.Background(), "loc-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_RulesIterator(t *testing.T) {
	t.Run("iterates all rules", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Rule `json:"items"`
			}{
				Items: []Rule{
					{ID: "rule-1", Name: "Rule 1"},
					{ID: "rule-2", Name: "Rule 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var rules []Rule
		for rule, err := range client.Rules(context.Background(), "loc-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			rules = append(rules, rule)
		}
		if len(rules) != 2 {
			t.Errorf("got %d rules, want 2", len(rules))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.Rules(context.Background(), "loc-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_ScenesIterator(t *testing.T) {
	t.Run("iterates all scenes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Scene `json:"items"`
			}{
				Items: []Scene{
					{SceneID: "scene-1", SceneName: "Morning"},
					{SceneID: "scene-2", SceneName: "Night"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var scenes []Scene
		for scene, err := range client.Scenes(context.Background(), "loc-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			scenes = append(scenes, scene)
		}
		if len(scenes) != 2 {
			t.Errorf("got %d scenes, want 2", len(scenes))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.Scenes(context.Background(), "loc-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_AppsIterator(t *testing.T) {
	t.Run("iterates all apps", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []App `json:"items"`
			}{
				Items: []App{
					{AppID: "app-1", AppName: "App 1"},
					{AppID: "app-2", AppName: "App 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var apps []App
		for app, err := range client.Apps(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			apps = append(apps, app)
		}
		if len(apps) != 2 {
			t.Errorf("got %d apps, want 2", len(apps))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.Apps(context.Background()) {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_DeviceProfilesIterator(t *testing.T) {
	t.Run("iterates all profiles", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []DeviceProfileFull `json:"items"`
			}{
				Items: []DeviceProfileFull{
					{ID: "profile-1", Name: "Profile 1"},
					{ID: "profile-2", Name: "Profile 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var profiles []DeviceProfileFull
		for profile, err := range client.DeviceProfiles(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			profiles = append(profiles, profile)
		}
		if len(profiles) != 2 {
			t.Errorf("got %d profiles, want 2", len(profiles))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.DeviceProfiles(context.Background()) {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_ModesIterator(t *testing.T) {
	t.Run("iterates all modes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Mode `json:"items"`
			}{
				Items: []Mode{
					{ID: "mode-1", Label: "Home"},
					{ID: "mode-2", Label: "Away"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var modes []Mode
		for mode, err := range client.Modes(context.Background(), "loc-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			modes = append(modes, mode)
		}
		if len(modes) != 2 {
			t.Errorf("got %d modes, want 2", len(modes))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.Modes(context.Background(), "loc-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_ChannelsIterator(t *testing.T) {
	t.Run("iterates all channels", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Channel `json:"items"`
			}{
				Items: []Channel{
					{ChannelID: "ch-1", Name: "Channel 1"},
					{ChannelID: "ch-2", Name: "Channel 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var channels []Channel
		for ch, err := range client.Channels(context.Background(), nil) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			channels = append(channels, ch)
		}
		if len(channels) != 2 {
			t.Errorf("got %d channels, want 2", len(channels))
		}
	})
}

func TestClient_DriversIterator(t *testing.T) {
	t.Run("iterates all drivers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []EdgeDriverSummary `json:"items"`
			}{
				Items: []EdgeDriverSummary{
					{DriverID: "driver-1", Name: "Driver 1"},
					{DriverID: "driver-2", Name: "Driver 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var drivers []EdgeDriverSummary
		for driver, err := range client.Drivers(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			drivers = append(drivers, driver)
		}
		if len(drivers) != 2 {
			t.Errorf("got %d drivers, want 2", len(drivers))
		}
	})
}

func TestClient_SchemaAppsIterator(t *testing.T) {
	t.Run("iterates all schema apps", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []SchemaApp `json:"items"`
			}{
				Items: []SchemaApp{
					{EndpointAppID: "app-1", AppName: "Schema App 1"},
					{EndpointAppID: "app-2", AppName: "Schema App 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var apps []SchemaApp
		for app, err := range client.SchemaApps(context.Background(), false) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			apps = append(apps, app)
		}
		if len(apps) != 2 {
			t.Errorf("got %d apps, want 2", len(apps))
		}
	})
}

func TestClient_InstalledAppsIterator(t *testing.T) {
	t.Run("iterates all installed apps", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []InstalledApp `json:"items"`
			}{
				Items: []InstalledApp{
					{InstalledAppID: "installed-1", DisplayName: "App 1"},
					{InstalledAppID: "installed-2", DisplayName: "App 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var apps []InstalledApp
		for app, err := range client.InstalledApps(context.Background(), "loc-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			apps = append(apps, app)
		}
		if len(apps) != 2 {
			t.Errorf("got %d apps, want 2", len(apps))
		}
	})
}

func TestClient_SubscriptionsIterator(t *testing.T) {
	t.Run("iterates all subscriptions", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Subscription `json:"items"`
			}{
				Items: []Subscription{
					{ID: "sub-1", SourceType: "DEVICE"},
					{ID: "sub-2", SourceType: "CAPABILITY"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var subs []Subscription
		for sub, err := range client.Subscriptions(context.Background(), "installed-app-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			subs = append(subs, sub)
		}
		if len(subs) != 2 {
			t.Errorf("got %d subscriptions, want 2", len(subs))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.Subscriptions(context.Background(), "installed-app-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_SchedulesIterator(t *testing.T) {
	t.Run("iterates all schedules", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Schedule `json:"items"`
			}{
				Items: []Schedule{
					{Name: "sched-1"},
					{Name: "sched-2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var schedules []Schedule
		for sched, err := range client.Schedules(context.Background(), "installed-app-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			schedules = append(schedules, sched)
		}
		if len(schedules) != 2 {
			t.Errorf("got %d schedules, want 2", len(schedules))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.Schedules(context.Background(), "installed-app-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_CapabilitiesIterator(t *testing.T) {
	t.Run("iterates all capabilities", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []CapabilityReference `json:"items"`
			}{
				Items: []CapabilityReference{
					{ID: "switch", Version: 1},
					{ID: "switchLevel", Version: 1},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var caps []CapabilityReference
		for cap, err := range client.Capabilities(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			caps = append(caps, cap)
		}
		if len(caps) != 2 {
			t.Errorf("got %d capabilities, want 2", len(caps))
		}
	})
}

func TestClient_DevicePreferencesIterator(t *testing.T) {
	t.Run("iterates all preferences", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []DevicePreference `json:"items"`
			}{
				Items: []DevicePreference{
					{PreferenceID: "pref-1"},
					{PreferenceID: "pref-2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var prefs []DevicePreference
		for pref, err := range client.DevicePreferences(context.Background(), "") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			prefs = append(prefs, pref)
		}
		if len(prefs) != 2 {
			t.Errorf("got %d preferences, want 2", len(prefs))
		}
	})
}

func TestClient_OrganizationsIterator(t *testing.T) {
	t.Run("iterates all organizations", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []Organization `json:"items"`
			}{
				Items: []Organization{
					{OrganizationID: "org-1", Name: "Org 1"},
					{OrganizationID: "org-2", Name: "Org 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var orgs []Organization
		for org, err := range client.Organizations(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			orgs = append(orgs, org)
		}
		if len(orgs) != 2 {
			t.Errorf("got %d organizations, want 2", len(orgs))
		}
	})
}

func TestClient_InstalledSchemaAppsIterator(t *testing.T) {
	t.Run("iterates all installed schema apps", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []InstalledSchemaApp `json:"items"`
			}{
				Items: []InstalledSchemaApp{
					{InstalledAppID: "installed-1", IsaID: "isa-1"},
					{InstalledAppID: "installed-2", IsaID: "isa-2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var apps []InstalledSchemaApp
		for app, err := range client.InstalledSchemaApps(context.Background(), "loc-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			apps = append(apps, app)
		}
		if len(apps) != 2 {
			t.Errorf("got %d apps, want 2", len(apps))
		}
	})
}

func TestClient_DeviceEventsIterator(t *testing.T) {
	t.Run("iterates device events", func(t *testing.T) {
		pageNum := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var resp PagedEvents
			switch pageNum {
			case 0:
				resp = PagedEvents{
					Items: []DeviceEvent{
						{DeviceID: "device-1", Capability: "switch", Attribute: "switch"},
						{DeviceID: "device-1", Capability: "switch", Attribute: "switch"},
					},
					Links: Links{Next: "/events?page=1"},
				}
			case 1:
				resp = PagedEvents{
					Items: []DeviceEvent{
						{DeviceID: "device-1", Capability: "switchLevel", Attribute: "level"},
					},
					Links: Links{Next: ""},
				}
			}
			pageNum++
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var events []DeviceEvent
		for ev, err := range client.DeviceEvents(context.Background(), "device-1", nil) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			events = append(events, ev)
		}
		if len(events) != 3 {
			t.Errorf("got %d events, want 3", len(events))
		}
	})
}

func TestClient_EnrolledChannelsIterator(t *testing.T) {
	t.Run("iterates all enrolled channels", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []EnrolledChannel `json:"items"`
			}{
				Items: []EnrolledChannel{
					{ChannelID: "ch-1", Name: "Channel 1"},
					{ChannelID: "ch-2", Name: "Channel 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var channels []EnrolledChannel
		for ch, err := range client.EnrolledChannels(context.Background(), "hub-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			channels = append(channels, ch)
		}
		if len(channels) != 2 {
			t.Errorf("got %d channels, want 2", len(channels))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.EnrolledChannels(context.Background(), "hub-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_InstalledDriversIterator(t *testing.T) {
	t.Run("iterates all installed drivers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []InstalledDriver `json:"items"`
			}{
				Items: []InstalledDriver{
					{DriverID: "driver-1", Name: "Driver 1"},
					{DriverID: "driver-2", Name: "Driver 2"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var drivers []InstalledDriver
		for driver, err := range client.InstalledDrivers(context.Background(), "hub-1", "") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			drivers = append(drivers, driver)
		}
		if len(drivers) != 2 {
			t.Errorf("got %d drivers, want 2", len(drivers))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.InstalledDrivers(context.Background(), "hub-1", "") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

func TestClient_AssignedDriversIterator(t *testing.T) {
	t.Run("iterates all assigned drivers", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []DriverChannelDetails `json:"items"`
			}{
				Items: []DriverChannelDetails{
					{DriverID: "driver-1", ChannelID: "ch-1"},
					{DriverID: "driver-2", ChannelID: "ch-1"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var drivers []DriverChannelDetails
		for driver, err := range client.AssignedDrivers(context.Background(), "ch-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			drivers = append(drivers, driver)
		}
		if len(drivers) != 2 {
			t.Errorf("got %d drivers, want 2", len(drivers))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.AssignedDrivers(context.Background(), "ch-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}

// Additional coverage tests for context cancellation and validation

func TestIterator_ContextCancellation(t *testing.T) {
	// All iterators should handle context cancellation
	client, _ := NewClient("token")
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	tests := []struct {
		name string
		iter func() error
	}{
		{"Locations", func() error {
			for _, err := range client.Locations(ctx) {
				return err
			}
			return nil
		}},
		{"Rooms", func() error {
			for _, err := range client.Rooms(ctx, "loc-1") {
				return err
			}
			return nil
		}},
		{"Rules", func() error {
			for _, err := range client.Rules(ctx, "loc-1") {
				return err
			}
			return nil
		}},
		{"Scenes", func() error {
			for _, err := range client.Scenes(ctx, "loc-1") {
				return err
			}
			return nil
		}},
		{"Apps", func() error {
			for _, err := range client.Apps(ctx) {
				return err
			}
			return nil
		}},
		{"DeviceProfiles", func() error {
			for _, err := range client.DeviceProfiles(ctx) {
				return err
			}
			return nil
		}},
		{"Capabilities", func() error {
			for _, err := range client.Capabilities(ctx) {
				return err
			}
			return nil
		}},
		{"InstalledApps", func() error {
			for _, err := range client.InstalledApps(ctx, "loc-1") {
				return err
			}
			return nil
		}},
		{"Subscriptions", func() error {
			for _, err := range client.Subscriptions(ctx, "app-1") {
				return err
			}
			return nil
		}},
		{"Schedules", func() error {
			for _, err := range client.Schedules(ctx, "app-1") {
				return err
			}
			return nil
		}},
		{"Modes", func() error {
			for _, err := range client.Modes(ctx, "loc-1") {
				return err
			}
			return nil
		}},
		{"Organizations", func() error {
			for _, err := range client.Organizations(ctx) {
				return err
			}
			return nil
		}},
		{"Channels", func() error {
			for _, err := range client.Channels(ctx, nil) {
				return err
			}
			return nil
		}},
		{"Drivers", func() error {
			for _, err := range client.Drivers(ctx) {
				return err
			}
			return nil
		}},
		{"SchemaApps", func() error {
			for _, err := range client.SchemaApps(ctx, false) {
				return err
			}
			return nil
		}},
		{"InstalledSchemaApps", func() error {
			for _, err := range client.InstalledSchemaApps(ctx, "loc-1") {
				return err
			}
			return nil
		}},
		{"DevicePreferences", func() error {
			for _, err := range client.DevicePreferences(ctx, "") {
				return err
			}
			return nil
		}},
		{"EnrolledChannels", func() error {
			for _, err := range client.EnrolledChannels(ctx, "hub-1") {
				return err
			}
			return nil
		}},
		{"InstalledDrivers", func() error {
			for _, err := range client.InstalledDrivers(ctx, "hub-1", "") {
				return err
			}
			return nil
		}},
		{"AssignedDrivers", func() error {
			for _, err := range client.AssignedDrivers(ctx, "ch-1") {
				return err
			}
			return nil
		}},
		{"SchemaAppInvitations", func() error {
			for _, err := range client.SchemaAppInvitations(ctx, "app-1") {
				return err
			}
			return nil
		}},
		{"DeviceEvents", func() error {
			for _, err := range client.DeviceEvents(ctx, "device-1", nil) {
				return err
			}
			return nil
		}},
		{"DevicesWithOptions", func() error {
			for _, err := range client.DevicesWithOptions(ctx, nil) {
				return err
			}
			return nil
		}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.iter()
			if err != context.Canceled {
				t.Errorf("%s: expected context.Canceled, got %v", tt.name, err)
			}
		})
	}
}

func TestIterator_EmptyIDValidation(t *testing.T) {
	client, _ := NewClient("token")
	ctx := context.Background()

	tests := []struct {
		name        string
		iter        func() error
		expectedErr error
	}{
		{"Rooms empty locationID", func() error {
			for _, err := range client.Rooms(ctx, "") {
				return err
			}
			return nil
		}, ErrEmptyLocationID},
		{"Subscriptions empty installedAppID", func() error {
			for _, err := range client.Subscriptions(ctx, "") {
				return err
			}
			return nil
		}, ErrEmptyInstalledAppID},
		{"Schedules empty installedAppID", func() error {
			for _, err := range client.Schedules(ctx, "") {
				return err
			}
			return nil
		}, ErrEmptyInstalledAppID},
		{"Modes empty locationID", func() error {
			for _, err := range client.Modes(ctx, "") {
				return err
			}
			return nil
		}, ErrEmptyLocationID},
		{"EnrolledChannels empty hubID", func() error {
			for _, err := range client.EnrolledChannels(ctx, "") {
				return err
			}
			return nil
		}, ErrEmptyHubID},
		{"InstalledDrivers empty hubID", func() error {
			for _, err := range client.InstalledDrivers(ctx, "", "") {
				return err
			}
			return nil
		}, ErrEmptyHubID},
		{"AssignedDrivers empty channelID", func() error {
			for _, err := range client.AssignedDrivers(ctx, "") {
				return err
			}
			return nil
		}, ErrEmptyChannelID},
		{"SchemaAppInvitations empty schemaAppID", func() error {
			for _, err := range client.SchemaAppInvitations(ctx, "") {
				return err
			}
			return nil
		}, ErrEmptySchemaAppID},
		{"DeviceEvents empty deviceID", func() error {
			for _, err := range client.DeviceEvents(ctx, "", nil) {
				return err
			}
			return nil
		}, ErrEmptyDeviceID},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.iter()
			if err != tt.expectedErr {
				t.Errorf("%s: expected %v, got %v", tt.name, tt.expectedErr, err)
			}
		})
	}
}

func TestIterator_EarlyTermination(t *testing.T) {
	// Test that iterators properly stop when yield returns false
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return many items - we should only process one
		resp := struct {
			Items []Device `json:"items"`
		}{
			Items: []Device{
				{DeviceID: "device-1"},
				{DeviceID: "device-2"},
				{DeviceID: "device-3"},
				{DeviceID: "device-4"},
				{DeviceID: "device-5"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))

	t.Run("Devices stops after first item", func(t *testing.T) {
		count := 0
		for device, err := range client.Devices(context.Background()) {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			count++
			if device.DeviceID == "device-1" {
				break // Stop after first
			}
		}
		if count != 1 {
			t.Errorf("expected 1 iteration, got %d", count)
		}
	})
}

func TestClient_SchemaAppInvitationsIterator(t *testing.T) {
	t.Run("iterates all invitations", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := struct {
				Items []SchemaAppInvitation `json:"items"`
			}{
				Items: []SchemaAppInvitation{
					{ID: "inv-1", SchemaAppID: "app-1"},
					{ID: "inv-2", SchemaAppID: "app-1"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var invites []SchemaAppInvitation
		for inv, err := range client.SchemaAppInvitations(context.Background(), "app-1") {
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			invites = append(invites, inv)
		}
		if len(invites) != 2 {
			t.Errorf("got %d invitations, want 2", len(invites))
		}
	})

	t.Run("handles server error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		var gotError bool
		for _, err := range client.SchemaAppInvitations(context.Background(), "app-1") {
			if err != nil {
				gotError = true
				break
			}
		}
		if !gotError {
			t.Error("expected error from iterator")
		}
	})
}
