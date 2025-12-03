package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetHub(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:  "successful response",
			hubID: "hub1",
			response: `{
				"deviceId": "hub1",
				"name": "SmartThings Hub",
				"type": "HUB"
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "not found",
			hubID:      "nonexistent",
			response:   `{"error": "not found"}`,
			statusCode: http.StatusNotFound,
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
			_, err := client.GetHub(context.Background(), tt.hubID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetHubCharacteristics(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:  "successful response",
			hubID: "hub1",
			response: `{
				"hubEUI": "1234567890ABCDEF",
				"zigbeeChannel": 15,
				"zigbeeEui": "00:11:22:33:44:55:66:77",
				"localIP": "192.168.1.100"
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
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
			_, err := client.GetHubCharacteristics(context.Background(), tt.hubID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_ListEnrolledChannels(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:  "successful response",
			hubID: "hub1",
			response: `{
				"items": [
					{"channelId": "chan1", "name": "Channel 1"},
					{"channelId": "chan2", "name": "Channel 2"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
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
			channels, err := client.ListEnrolledChannels(context.Background(), tt.hubID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(channels) != tt.wantCount {
				t.Errorf("expected %d channels, got %d", tt.wantCount, len(channels))
			}
		})
	}
}

func TestClient_ListInstalledDrivers(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:  "successful response",
			hubID: "hub1",
			response: `{
				"items": [
					{"driverId": "driver1", "name": "Driver 1", "version": "1.0"},
					{"driverId": "driver2", "name": "Driver 2", "version": "2.0"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
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
			drivers, err := client.ListInstalledDrivers(context.Background(), tt.hubID, "")

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(drivers) != tt.wantCount {
				t.Errorf("expected %d drivers, got %d", tt.wantCount, len(drivers))
			}
		})
	}
}

func TestClient_GetInstalledDriver(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		driverID   string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			hubID:      "hub1",
			driverID:   "driver1",
			response:   `{"driverId": "driver1", "name": "Test Driver", "version": "1.0"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
			driverID:   "driver1",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver ID",
			hubID:      "hub1",
			driverID:   "",
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
			_, err := client.GetInstalledDriver(context.Background(), tt.hubID, tt.driverID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_InstallDriver(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		channelID  string
		driverID   string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful installation",
			hubID:      "hub1",
			channelID:  "chan1",
			driverID:   "driver1",
			response:   `{"driverId": "driver1", "hubId": "hub1"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
			channelID:  "chan1",
			driverID:   "driver1",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty channel ID",
			hubID:      "hub1",
			channelID:  "",
			driverID:   "driver1",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver ID",
			hubID:      "hub1",
			channelID:  "chan1",
			driverID:   "",
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
			err := client.InstallDriver(context.Background(), tt.hubID, tt.channelID, tt.driverID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_UninstallDriver(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		driverID   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful uninstallation",
			hubID:      "hub1",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver ID",
			hubID:      "hub1",
			driverID:   "",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			err := client.UninstallDriver(context.Background(), tt.hubID, tt.driverID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_SwitchDriver(t *testing.T) {
	tests := []struct {
		name       string
		hubID      string
		deviceID   string
		driverID   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful switch",
			hubID:      "hub1",
			deviceID:   "device1",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty hub ID",
			hubID:      "",
			deviceID:   "device1",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty device ID",
			hubID:      "hub1",
			deviceID:   "",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver ID",
			hubID:      "hub1",
			deviceID:   "device1",
			driverID:   "",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			err := client.SwitchDriver(context.Background(), tt.driverID, tt.hubID, tt.deviceID, false)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestExtractHubData(t *testing.T) {
	tests := []struct {
		name      string
		status    Status
		want      *HubData
		wantEmpty bool
	}{
		{
			name: "full hub data in main component",
			status: Status{
				"main": map[string]any{
					"hubData": map[string]any{
						"localIP":                 "192.168.1.100",
						"macAddress":              "AA:BB:CC:DD:EE:FF",
						"hardwareType":            "HubV3",
						"hubLocalApiAvailability": "available",
						"zigbeeEui":               "00:11:22:33:44:55:66:77",
						"zigbeeChannel":           float64(15),
						"zwaveRegion":             "US",
						"zwaveSucId":              float64(1),
						"zwaveHomeId":             "12345678",
					},
				},
			},
			want: &HubData{
				LocalIP:                 "192.168.1.100",
				MacAddress:              "AA:BB:CC:DD:EE:FF",
				HardwareType:            "HubV3",
				HubLocalAPIAvailability: "available",
				ZigbeeEUI:               "00:11:22:33:44:55:66:77",
				ZigbeeChannel:           15,
				ZWaveRegion:             "US",
				ZWaveSUCID:              1,
				ZWaveHomeID:             "12345678",
			},
		},
		{
			name: "hub data at top level",
			status: Status{
				"hubData": map[string]any{
					"localIP":    "10.0.0.1",
					"macAddress": "11:22:33:44:55:66",
				},
			},
			want: &HubData{
				LocalIP:    "10.0.0.1",
				MacAddress: "11:22:33:44:55:66",
			},
		},
		{
			name:      "nil status",
			status:    nil,
			wantEmpty: true,
		},
		{
			name:      "empty status",
			status:    Status{},
			wantEmpty: true,
		},
		{
			name: "missing hubData key",
			status: Status{
				"main": map[string]any{},
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractHubData(tt.status)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}
			if tt.wantEmpty {
				// Function returns empty struct for nil/empty/missing hubData
				if got == nil {
					t.Error("expected empty struct, got nil")
				}
				return
			}
			if got == nil {
				t.Error("expected hub data, got nil")
				return
			}
			if got.LocalIP != tt.want.LocalIP {
				t.Errorf("LocalIP = %s, want %s", got.LocalIP, tt.want.LocalIP)
			}
			if got.MacAddress != tt.want.MacAddress {
				t.Errorf("MacAddress = %s, want %s", got.MacAddress, tt.want.MacAddress)
			}
			if got.ZigbeeChannel != tt.want.ZigbeeChannel {
				t.Errorf("ZigbeeChannel = %d, want %d", got.ZigbeeChannel, tt.want.ZigbeeChannel)
			}
		})
	}
}
