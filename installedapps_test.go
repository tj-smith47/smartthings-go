package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListInstalledApps(t *testing.T) {
	t.Run("successful response with location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps")
			}
			if r.URL.Query().Get("locationId") != "loc-123" {
				t.Errorf("locationId query = %q, want %q", r.URL.Query().Get("locationId"), "loc-123")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := installedAppListResponse{
				Items: []InstalledApp{
					{
						InstalledAppID:     "app-1",
						DisplayName:        "My Smart App",
						InstalledAppType:   "WEBHOOK_SMART_APP",
						InstalledAppStatus: "AUTHORIZED",
						LocationID:         "loc-123",
					},
					{
						InstalledAppID:     "app-2",
						DisplayName:        "Another App",
						InstalledAppType:   "LAMBDA_SMART_APP",
						InstalledAppStatus: "AUTHORIZED",
						LocationID:         "loc-123",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		apps, err := client.ListInstalledApps(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 2 {
			t.Errorf("got %d apps, want 2", len(apps))
		}
		if apps[0].DisplayName != "My Smart App" {
			t.Errorf("apps[0].DisplayName = %q, want %q", apps[0].DisplayName, "My Smart App")
		}
	})

	t.Run("successful response without location", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("locationId") != "" {
				t.Errorf("locationId query should be empty, got %q", r.URL.Query().Get("locationId"))
			}
			resp := installedAppListResponse{
				Items: []InstalledApp{
					{InstalledAppID: "app-1", DisplayName: "App 1"},
					{InstalledAppID: "app-2", DisplayName: "App 2"},
					{InstalledAppID: "app-3", DisplayName: "App 3"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		apps, err := client.ListInstalledApps(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 3 {
			t.Errorf("got %d apps, want 3", len(apps))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(installedAppListResponse{Items: []InstalledApp{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		apps, err := client.ListInstalledApps(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 0 {
			t.Errorf("got %d apps, want 0", len(apps))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListInstalledApps(context.Background(), "loc-123")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestClient_GetInstalledApp(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123")
			}
			app := InstalledApp{
				InstalledAppID:     "app-123",
				DisplayName:        "My Smart App",
				InstalledAppType:   "WEBHOOK_SMART_APP",
				InstalledAppStatus: "AUTHORIZED",
				AppID:              "original-app-id",
				LocationID:         "loc-456",
				CreatedDate:        "2024-01-01T00:00:00Z",
				LastUpdatedDate:    "2024-06-01T00:00:00Z",
			}
			json.NewEncoder(w).Encode(app)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		app, err := client.GetInstalledApp(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if app.InstalledAppID != "app-123" {
			t.Errorf("InstalledAppID = %q, want %q", app.InstalledAppID, "app-123")
		}
		if app.DisplayName != "My Smart App" {
			t.Errorf("DisplayName = %q, want %q", app.DisplayName, "My Smart App")
		}
		if app.InstalledAppStatus != "AUTHORIZED" {
			t.Errorf("InstalledAppStatus = %q, want %q", app.InstalledAppStatus, "AUTHORIZED")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetInstalledApp(context.Background(), "")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("app not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetInstalledApp(context.Background(), "missing-app")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_DeleteInstalledApp(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteInstalledApp(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteInstalledApp(context.Background(), "")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("app not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteInstalledApp(context.Background(), "missing-app")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_ListInstalledAppConfigs(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/configs" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/configs")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := installedAppConfigListResponse{
				Items: []InstalledAppConfigItem{
					{
						ConfigurationID:     "config-1",
						ConfigurationStatus: "AUTHORIZED",
						CreatedDate:         "2024-01-01T00:00:00Z",
						LastUpdatedDate:     "2024-06-01T00:00:00Z",
					},
					{
						ConfigurationID:     "config-2",
						ConfigurationStatus: "STAGED",
						CreatedDate:         "2024-06-02T00:00:00Z",
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		configs, err := client.ListInstalledAppConfigs(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(configs) != 2 {
			t.Errorf("got %d configs, want 2", len(configs))
		}
		if configs[0].ConfigurationID != "config-1" {
			t.Errorf("configs[0].ConfigurationID = %q, want %q", configs[0].ConfigurationID, "config-1")
		}
		if configs[0].ConfigurationStatus != "AUTHORIZED" {
			t.Errorf("configs[0].ConfigurationStatus = %q, want %q", configs[0].ConfigurationStatus, "AUTHORIZED")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.ListInstalledAppConfigs(context.Background(), "")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})
}

func TestClient_GetInstalledAppConfig(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/configs/config-456" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/configs/config-456")
			}
			config := InstalledAppConfiguration{
				InstalledAppID:      "app-123",
				ConfigurationID:     "config-456",
				ConfigurationStatus: "AUTHORIZED",
				CreatedDate:         "2024-01-01T00:00:00Z",
				Config: map[string][]ConfigEntry{
					"devices": {
						{
							ValueType: ConfigValueTypeDevice,
							DeviceConfig: &DeviceConfig{
								DeviceID:    "device-789",
								ComponentID: "main",
								Permissions: []string{"r:devices:*", "x:devices:*"},
							},
						},
					},
					"name": {
						{
							ValueType:    ConfigValueTypeString,
							StringConfig: &StringConfig{Value: "My App Name"},
						},
					},
				},
			}
			json.NewEncoder(w).Encode(config)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		config, err := client.GetInstalledAppConfig(context.Background(), "app-123", "config-456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config.InstalledAppID != "app-123" {
			t.Errorf("InstalledAppID = %q, want %q", config.InstalledAppID, "app-123")
		}
		if config.ConfigurationID != "config-456" {
			t.Errorf("ConfigurationID = %q, want %q", config.ConfigurationID, "config-456")
		}
		if config.ConfigurationStatus != "AUTHORIZED" {
			t.Errorf("ConfigurationStatus = %q, want %q", config.ConfigurationStatus, "AUTHORIZED")
		}
		if len(config.Config) != 2 {
			t.Errorf("got %d config entries, want 2", len(config.Config))
		}

		// Verify device config
		devices := config.Config["devices"]
		if len(devices) != 1 {
			t.Fatalf("got %d device entries, want 1", len(devices))
		}
		if devices[0].ValueType != ConfigValueTypeDevice {
			t.Errorf("devices[0].ValueType = %q, want %q", devices[0].ValueType, ConfigValueTypeDevice)
		}
		if devices[0].DeviceConfig == nil {
			t.Fatal("expected DeviceConfig to be set")
		}
		if devices[0].DeviceConfig.DeviceID != "device-789" {
			t.Errorf("DeviceID = %q, want %q", devices[0].DeviceConfig.DeviceID, "device-789")
		}

		// Verify string config
		names := config.Config["name"]
		if len(names) != 1 {
			t.Fatalf("got %d name entries, want 1", len(names))
		}
		if names[0].StringConfig == nil {
			t.Fatal("expected StringConfig to be set")
		}
		if names[0].StringConfig.Value != "My App Name" {
			t.Errorf("StringConfig.Value = %q, want %q", names[0].StringConfig.Value, "My App Name")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetInstalledAppConfig(context.Background(), "", "config-123")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("empty config ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetInstalledAppConfig(context.Background(), "app-123", "")
		if err == nil {
			t.Fatal("expected error for empty config ID")
		}
	})

	t.Run("config not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetInstalledAppConfig(context.Background(), "app-123", "missing-config")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_GetCurrentInstalledAppConfig(t *testing.T) {
	t.Run("returns authorized config", func(t *testing.T) {
		callCount := 0
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.URL.Path == "/installedapps/app-123/configs" {
				resp := installedAppConfigListResponse{
					Items: []InstalledAppConfigItem{
						{ConfigurationID: "config-old", ConfigurationStatus: "REVOKED"},
						{ConfigurationID: "config-current", ConfigurationStatus: "AUTHORIZED"},
						{ConfigurationID: "config-staged", ConfigurationStatus: "STAGED"},
					},
				}
				json.NewEncoder(w).Encode(resp)
			} else if r.URL.Path == "/installedapps/app-123/configs/config-current" {
				config := InstalledAppConfiguration{
					InstalledAppID:      "app-123",
					ConfigurationID:     "config-current",
					ConfigurationStatus: "AUTHORIZED",
					Config:              map[string][]ConfigEntry{},
				}
				json.NewEncoder(w).Encode(config)
			} else {
				t.Errorf("unexpected path: %s", r.URL.Path)
				w.WriteHeader(http.StatusNotFound)
			}
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		config, err := client.GetCurrentInstalledAppConfig(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config == nil {
			t.Fatal("expected config, got nil")
		}
		if config.ConfigurationID != "config-current" {
			t.Errorf("ConfigurationID = %q, want %q", config.ConfigurationID, "config-current")
		}
		if callCount != 2 {
			t.Errorf("expected 2 API calls, got %d", callCount)
		}
	})

	t.Run("returns nil when no authorized config", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := installedAppConfigListResponse{
				Items: []InstalledAppConfigItem{
					{ConfigurationID: "config-staged", ConfigurationStatus: "STAGED"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		config, err := client.GetCurrentInstalledAppConfig(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if config != nil {
			t.Errorf("expected nil config, got %+v", config)
		}
	})
}
