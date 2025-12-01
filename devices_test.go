package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListDevices(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices")
			}
			resp := DeviceListResponse{
				Items: []Device{
					{DeviceID: "device-1", Label: "Living Room Light"},
					{DeviceID: "device-2", Label: "Kitchen Light"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		devices, err := client.ListDevices(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(devices) != 2 {
			t.Errorf("got %d devices, want 2", len(devices))
		}
		if devices[0].Label != "Living Room Light" {
			t.Errorf("devices[0].Label = %q, want %q", devices[0].Label, "Living Room Light")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(DeviceListResponse{Items: []Device{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		devices, err := client.ListDevices(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(devices) != 0 {
			t.Errorf("got %d devices, want 0", len(devices))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListDevices(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestClient_GetDevice(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123")
			}
			device := Device{
				DeviceID:         "device-123",
				Label:            "My Device",
				ManufacturerName: "Samsung",
			}
			json.NewEncoder(w).Encode(device)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		device, err := client.GetDevice(context.Background(), "device-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if device.DeviceID != "device-123" {
			t.Errorf("DeviceID = %q, want %q", device.DeviceID, "device-123")
		}
		if device.Label != "My Device" {
			t.Errorf("Label = %q, want %q", device.Label, "My Device")
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetDevice(context.Background(), "")
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})

	t.Run("device not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDevice(context.Background(), "missing-device")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_GetDeviceStatus(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123/components/main/status" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123/components/main/status")
			}
			status := Status{
				"switch": map[string]any{
					"switch": map[string]any{
						"value": "on",
					},
				},
			}
			json.NewEncoder(w).Encode(status)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		status, err := client.GetDeviceStatus(context.Background(), "device-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		power, ok := GetString(status, "switch", "switch", "value")
		if !ok || power != "on" {
			t.Errorf("switch value = %q, want %q", power, "on")
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetDeviceStatus(context.Background(), "")
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})
}

func TestClient_GetDeviceFullStatus(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123/status" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123/status")
			}
			resp := map[string]any{
				"components": map[string]any{
					"main": map[string]any{
						"switch": map[string]any{
							"switch": map[string]any{"value": "on"},
						},
					},
					"cooler": map[string]any{
						"temperatureMeasurement": map[string]any{
							"temperature": map[string]any{"value": 4.0},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		components, err := client.GetDeviceFullStatus(context.Background(), "device-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(components) != 2 {
			t.Errorf("got %d components, want 2", len(components))
		}
		if _, ok := components["main"]; !ok {
			t.Error("missing 'main' component")
		}
		if _, ok := components["cooler"]; !ok {
			t.Error("missing 'cooler' component")
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetDeviceFullStatus(context.Background(), "")
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})
}

func TestClient_GetDeviceStatusAllComponents(t *testing.T) {
	t.Run("merges components", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"components": map[string]any{
					"main":    map[string]any{"key1": "value1"},
					"cooler":  map[string]any{"key2": "value2"},
					"freezer": map[string]any{"key3": "value3"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		merged, err := client.GetDeviceStatusAllComponents(context.Background(), "device-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(merged) != 3 {
			t.Errorf("got %d merged entries, want 3", len(merged))
		}
	})
}

func TestClient_GetComponentStatus(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123/components/cooler/status" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123/components/cooler/status")
			}
			status := Status{
				"temperatureMeasurement": map[string]any{
					"temperature": map[string]any{"value": 4.0},
				},
			}
			json.NewEncoder(w).Encode(status)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		status, err := client.GetComponentStatus(context.Background(), "device-123", "cooler")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if status == nil {
			t.Fatal("status is nil")
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetComponentStatus(context.Background(), "", "main")
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})

	t.Run("empty component ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetComponentStatus(context.Background(), "device-123", "")
		if err != ErrEmptyComponentID {
			t.Errorf("expected ErrEmptyComponentID, got %v", err)
		}
	})
}

func TestClient_ExecuteCommand(t *testing.T) {
	t.Run("successful command", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/devices/device-123/commands" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/devices/device-123/commands")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if len(req.Commands) != 1 {
				t.Errorf("got %d commands, want 1", len(req.Commands))
			}
			if req.Commands[0].Capability != "switch" {
				t.Errorf("capability = %q, want %q", req.Commands[0].Capability, "switch")
			}
			if req.Commands[0].Command != "on" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "on")
			}

			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.ExecuteCommand(context.Background(), "device-123", NewCommand("switch", "on"))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty device ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.ExecuteCommand(context.Background(), "", NewCommand("switch", "on"))
		if err != ErrEmptyDeviceID {
			t.Errorf("expected ErrEmptyDeviceID, got %v", err)
		}
	})
}

func TestClient_ExecuteCommands(t *testing.T) {
	t.Run("multiple commands", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if len(req.Commands) != 2 {
				t.Errorf("got %d commands, want 2", len(req.Commands))
			}
			// Verify default component is set
			for _, cmd := range req.Commands {
				if cmd.Component == "" {
					t.Error("component should default to 'main'")
				}
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		cmds := []Command{
			NewCommand("switch", "on"),
			NewCommand("audioVolume", "setVolume", 50),
		}
		err := client.ExecuteCommands(context.Background(), "device-123", cmds)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("preserves explicit component", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Component != "cooler" {
				t.Errorf("component = %q, want %q", req.Commands[0].Component, "cooler")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		cmds := []Command{
			NewComponentCommand("cooler", "temperatureSetpoint", "setTemperature", 4),
		}
		err := client.ExecuteCommands(context.Background(), "device-123", cmds)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestNewCommand(t *testing.T) {
	t.Run("without arguments", func(t *testing.T) {
		cmd := NewCommand("switch", "on")
		if cmd.Component != "main" {
			t.Errorf("Component = %q, want %q", cmd.Component, "main")
		}
		if cmd.Capability != "switch" {
			t.Errorf("Capability = %q, want %q", cmd.Capability, "switch")
		}
		if cmd.Command != "on" {
			t.Errorf("Command = %q, want %q", cmd.Command, "on")
		}
		if len(cmd.Arguments) != 0 {
			t.Errorf("Arguments length = %d, want 0", len(cmd.Arguments))
		}
	})

	t.Run("with arguments", func(t *testing.T) {
		cmd := NewCommand("audioVolume", "setVolume", 50)
		if len(cmd.Arguments) != 1 {
			t.Fatalf("Arguments length = %d, want 1", len(cmd.Arguments))
		}
		if cmd.Arguments[0] != 50 {
			t.Errorf("Arguments[0] = %v, want 50", cmd.Arguments[0])
		}
	})
}

func TestNewComponentCommand(t *testing.T) {
	cmd := NewComponentCommand("cooler", "temperatureSetpoint", "setTemperature", 4)
	if cmd.Component != "cooler" {
		t.Errorf("Component = %q, want %q", cmd.Component, "cooler")
	}
	if cmd.Capability != "temperatureSetpoint" {
		t.Errorf("Capability = %q, want %q", cmd.Capability, "temperatureSetpoint")
	}
}

func TestFilterDevices(t *testing.T) {
	devices := []Device{
		{DeviceID: "1", ManufacturerName: "Samsung", Label: "TV"},
		{DeviceID: "2", ManufacturerName: "Philips", Label: "Light"},
		{DeviceID: "3", ManufacturerName: "Samsung", Label: "Fridge"},
	}

	t.Run("filter by manufacturer", func(t *testing.T) {
		samsung := FilterDevices(devices, func(d Device) bool {
			return d.ManufacturerName == "Samsung"
		})
		if len(samsung) != 2 {
			t.Errorf("got %d devices, want 2", len(samsung))
		}
	})

	t.Run("no matches", func(t *testing.T) {
		result := FilterDevices(devices, func(d Device) bool {
			return d.ManufacturerName == "LG"
		})
		if len(result) != 0 {
			t.Errorf("got %d devices, want 0", len(result))
		}
	})
}

func TestFilterByManufacturer(t *testing.T) {
	devices := []Device{
		{DeviceID: "1", ManufacturerName: "Samsung"},
		{DeviceID: "2", ManufacturerName: "Philips"},
		{DeviceID: "3", ManufacturerName: "Samsung"},
	}

	result := FilterByManufacturer(devices, "Samsung")
	if len(result) != 2 {
		t.Errorf("got %d devices, want 2", len(result))
	}
}

func TestFindDeviceByLabel(t *testing.T) {
	devices := []Device{
		{DeviceID: "1", Label: "Living Room"},
		{DeviceID: "2", Label: "Kitchen"},
	}

	t.Run("found", func(t *testing.T) {
		device := FindDeviceByLabel(devices, "Kitchen")
		if device == nil {
			t.Fatal("device is nil")
		}
		if device.DeviceID != "2" {
			t.Errorf("DeviceID = %q, want %q", device.DeviceID, "2")
		}
	})

	t.Run("not found", func(t *testing.T) {
		device := FindDeviceByLabel(devices, "Bedroom")
		if device != nil {
			t.Errorf("expected nil, got device with ID %q", device.DeviceID)
		}
	})
}

func TestFindDeviceByID(t *testing.T) {
	devices := []Device{
		{DeviceID: "device-1", Label: "Device 1"},
		{DeviceID: "device-2", Label: "Device 2"},
	}

	t.Run("found", func(t *testing.T) {
		device := FindDeviceByID(devices, "device-2")
		if device == nil {
			t.Fatal("device is nil")
		}
		if device.Label != "Device 2" {
			t.Errorf("Label = %q, want %q", device.Label, "Device 2")
		}
	})

	t.Run("not found", func(t *testing.T) {
		device := FindDeviceByID(devices, "missing")
		if device != nil {
			t.Errorf("expected nil, got device with ID %q", device.DeviceID)
		}
	})
}
