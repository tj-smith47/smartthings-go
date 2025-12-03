package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListVirtualDevices(t *testing.T) {
	tests := []struct {
		name       string
		opts       *VirtualDeviceListOptions
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name: "successful response with location",
			opts: &VirtualDeviceListOptions{LocationID: "loc1"},
			response: `{
				"items": [
					{"deviceId": "vdev1", "name": "Virtual Switch"},
					{"deviceId": "vdev2", "name": "Virtual Motion"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "successful response without options",
			opts:       nil,
			response:   `{"items": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			devices, err := client.ListVirtualDevices(context.Background(), tt.opts)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(devices) != tt.wantCount {
				t.Errorf("expected %d devices, got %d", tt.wantCount, len(devices))
			}
		})
	}
}

func TestClient_CreateVirtualDevice(t *testing.T) {
	tests := []struct {
		name       string
		device     *VirtualDeviceCreateRequest
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "successful creation",
			device: &VirtualDeviceCreateRequest{
				Name:            "Virtual Switch",
				RoomID:          "room1",
				DeviceProfileID: "profile1",
			},
			response:   `{"deviceId": "vdev1", "name": "Virtual Switch"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil device",
			device:     nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "empty name",
			device: &VirtualDeviceCreateRequest{
				Name: "",
			},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.CreateVirtualDevice(context.Background(), tt.device)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CreateStandardVirtualDevice(t *testing.T) {
	tests := []struct {
		name       string
		device     *VirtualDeviceStandardCreateRequest
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "successful creation",
			device: &VirtualDeviceStandardCreateRequest{
				Name:      "Virtual Switch",
				RoomID:    "room1",
				Prototype: "VIRTUAL_SWITCH",
			},
			response:   `{"deviceId": "vdev1", "name": "Virtual Switch"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil device",
			device:     nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "empty name",
			device: &VirtualDeviceStandardCreateRequest{
				Name:      "",
				Prototype: "VIRTUAL_SWITCH",
			},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "empty prototype",
			device: &VirtualDeviceStandardCreateRequest{
				Name:      "Test Device",
				Prototype: "",
			},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.CreateStandardVirtualDevice(context.Background(), tt.device)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CreateVirtualDeviceEvents(t *testing.T) {
	tests := []struct {
		name       string
		deviceID   string
		events     []VirtualDeviceEvent
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:     "successful event creation",
			deviceID: "vdev1",
			events: []VirtualDeviceEvent{
				{
					Component:  "main",
					Capability: "switch",
					Attribute:  "switch",
					Value:      "on",
				},
			},
			response:   `{"stateChanges": [{"component": "main", "capability": "switch", "attribute": "switch", "value": "on", "stateChange": true}]}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty device ID",
			deviceID:   "",
			events:     []VirtualDeviceEvent{{Component: "main"}},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "nil events",
			deviceID:   "vdev1",
			events:     nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty events",
			deviceID:   "vdev1",
			events:     []VirtualDeviceEvent{},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.CreateVirtualDeviceEvents(context.Background(), tt.deviceID, tt.events)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
