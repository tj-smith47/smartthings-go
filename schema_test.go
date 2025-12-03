package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListSchemaApps(t *testing.T) {
	tests := []struct {
		name                    string
		includeAllOrganizations bool
		response                string
		statusCode              int
		wantErr                 bool
		wantCount               int
	}{
		{
			name:                    "successful response",
			includeAllOrganizations: false,
			response: `{
				"items": [
					{"appId": "app1", "appName": "Test App"},
					{"appId": "app2", "appName": "Another App"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:                    "empty list",
			includeAllOrganizations: true,
			response:                `{"items": []}`,
			statusCode:              http.StatusOK,
			wantErr:                 false,
			wantCount:               0,
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
			apps, err := client.ListSchemaApps(context.Background(), tt.includeAllOrganizations)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(apps) != tt.wantCount {
				t.Errorf("expected %d apps, got %d", tt.wantCount, len(apps))
			}
		})
	}
}

func TestClient_GetSchemaApp(t *testing.T) {
	tests := []struct {
		name       string
		appID      string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			appID:      "app1",
			response:   `{"appId": "app1", "appName": "Test App", "partnerName": "Partner"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty app ID",
			appID:      "",
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
			_, err := client.GetSchemaApp(context.Background(), tt.appID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CreateSchemaApp(t *testing.T) {
	tests := []struct {
		name           string
		app            *SchemaAppRequest
		organizationID string
		response       string
		statusCode     int
		wantErr        bool
	}{
		{
			name: "successful creation",
			app: &SchemaAppRequest{
				AppName:     "Test App",
				PartnerName: "Partner",
				HostingType: "webhook",
			},
			organizationID: "",
			response:       `{"appId": "app1", "appName": "Test App"}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "nil app",
			app:            nil,
			organizationID: "",
			response:       ``,
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
		{
			name: "empty app name",
			app: &SchemaAppRequest{
				AppName: "",
			},
			organizationID: "",
			response:       ``,
			statusCode:     http.StatusOK,
			wantErr:        true,
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
			_, err := client.CreateSchemaApp(context.Background(), tt.app, tt.organizationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_UpdateSchemaApp(t *testing.T) {
	tests := []struct {
		name           string
		appID          string
		update         *SchemaAppRequest
		organizationID string
		statusCode     int
		wantErr        bool
	}{
		{
			name:  "successful update",
			appID: "app1",
			update: &SchemaAppRequest{
				AppName:     "Updated App",
				PartnerName: "Updated Partner",
			},
			organizationID: "",
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty app ID",
			appID:          "",
			update:         &SchemaAppRequest{AppName: "Test"},
			organizationID: "",
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			err := client.UpdateSchemaApp(context.Background(), tt.appID, tt.update, tt.organizationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_DeleteSchemaApp(t *testing.T) {
	tests := []struct {
		name       string
		appID      string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful deletion",
			appID:      "app1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty app ID",
			appID:      "",
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
			err := client.DeleteSchemaApp(context.Background(), tt.appID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_RegenerateSchemaAppOAuth(t *testing.T) {
	tests := []struct {
		name       string
		appID      string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful regeneration",
			appID:      "app1",
			response:   `{"clientId": "newclient", "clientSecret": "newsecret"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty app ID",
			appID:      "",
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
			_, err := client.RegenerateSchemaAppOAuth(context.Background(), tt.appID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetSchemaAppPage(t *testing.T) {
	tests := []struct {
		name       string
		appID      string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			appID:      "app1",
			response:   `{"pageId": "page1", "sections": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty app ID",
			appID:      "",
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
			_, err := client.GetSchemaAppPage(context.Background(), tt.appID, "loc1")

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_ListInstalledSchemaApps(t *testing.T) {
	tests := []struct {
		name       string
		locationID string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "successful response",
			locationID: "loc1",
			response: `{
				"items": [
					{"installedAppId": "installed1", "appId": "app1"},
					{"installedAppId": "installed2", "appId": "app2"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty list",
			locationID: "loc1",
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
			apps, err := client.ListInstalledSchemaApps(context.Background(), tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(apps) != tt.wantCount {
				t.Errorf("expected %d apps, got %d", tt.wantCount, len(apps))
			}
		})
	}
}

func TestClient_GetInstalledSchemaApp(t *testing.T) {
	tests := []struct {
		name           string
		installedAppID string
		response       string
		statusCode     int
		wantErr        bool
	}{
		{
			name:           "successful response",
			installedAppID: "installed1",
			response:       `{"installedAppId": "installed1", "appId": "app1"}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty installed app ID",
			installedAppID: "",
			response:       ``,
			statusCode:     http.StatusOK,
			wantErr:        true,
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
			_, err := client.GetInstalledSchemaApp(context.Background(), tt.installedAppID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_DeleteInstalledSchemaApp(t *testing.T) {
	tests := []struct {
		name           string
		installedAppID string
		statusCode     int
		wantErr        bool
	}{
		{
			name:           "successful deletion",
			installedAppID: "installed1",
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty installed app ID",
			installedAppID: "",
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			err := client.DeleteInstalledSchemaApp(context.Background(), tt.installedAppID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
