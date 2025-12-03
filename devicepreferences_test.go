package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListDevicePreferences(t *testing.T) {
	tests := []struct {
		name       string
		namespace  string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:      "successful response",
			namespace: "",
			response: `{
				"items": [
					{"preferenceId": "pref1", "name": "Preference 1"},
					{"preferenceId": "pref2", "name": "Preference 2"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty list",
			namespace:  "test-ns",
			response:   `{"items": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "invalid JSON response",
			namespace:  "",
			response:   `{invalid json`,
			statusCode: http.StatusOK,
			wantErr:    true,
			wantCount:  0,
		},
		{
			name:       "server error",
			namespace:  "",
			response:   `{"error": "internal error"}`,
			statusCode: http.StatusInternalServerError,
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
			prefs, err := client.ListDevicePreferences(context.Background(), tt.namespace)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(prefs) != tt.wantCount {
				t.Errorf("expected %d preferences, got %d", tt.wantCount, len(prefs))
			}
		})
	}
}

func TestClient_GetDevicePreference(t *testing.T) {
	tests := []struct {
		name         string
		preferenceID string
		response     string
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "successful response",
			preferenceID: "pref1",
			response:     `{"preferenceId": "pref1", "name": "Test Preference", "preferenceType": "integer"}`,
			statusCode:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "empty preference ID",
			preferenceID: "",
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "invalid JSON response",
			preferenceID: "pref1",
			response:     `{invalid json`,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			preferenceID: "pref1",
			response:     `{"error": "not found"}`,
			statusCode:   http.StatusNotFound,
			wantErr:      true,
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
			_, err := client.GetDevicePreference(context.Background(), tt.preferenceID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CreateDevicePreference(t *testing.T) {
	tests := []struct {
		name       string
		pref       *DevicePreferenceCreate
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "successful creation",
			pref: &DevicePreferenceCreate{
				Name:           "Test Preference",
				PreferenceType: PreferenceTypeInteger,
			},
			response:   `{"preferenceId": "pref1", "name": "Test Preference"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil preference",
			pref:       nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "empty name",
			pref: &DevicePreferenceCreate{
				Name:           "",
				PreferenceType: PreferenceTypeInteger,
			},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "invalid JSON response",
			pref: &DevicePreferenceCreate{
				Name:           "Test Preference",
				PreferenceType: PreferenceTypeInteger,
			},
			response:   `{invalid json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "server error",
			pref: &DevicePreferenceCreate{
				Name:           "Test Preference",
				PreferenceType: PreferenceTypeInteger,
			},
			response:   `{"error": "bad request"}`,
			statusCode: http.StatusBadRequest,
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
			_, err := client.CreateDevicePreference(context.Background(), tt.pref)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_UpdateDevicePreference(t *testing.T) {
	tests := []struct {
		name         string
		preferenceID string
		update       *DevicePreference
		response     string
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "successful update",
			preferenceID: "pref1",
			update: &DevicePreference{
				Name: "Updated Preference",
			},
			response:   `{"preferenceId": "pref1", "name": "Updated Preference"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:         "empty preference ID",
			preferenceID: "",
			update:       &DevicePreference{Name: "Test"},
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "nil update",
			preferenceID: "pref1",
			update:       nil,
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "invalid JSON response",
			preferenceID: "pref1",
			update:       &DevicePreference{Name: "Test"},
			response:     `{invalid json`,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			preferenceID: "pref1",
			update:       &DevicePreference{Name: "Test"},
			response:     `{"error": "not found"}`,
			statusCode:   http.StatusNotFound,
			wantErr:      true,
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
			_, err := client.UpdateDevicePreference(context.Background(), tt.preferenceID, tt.update)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CreatePreferenceTranslations(t *testing.T) {
	tests := []struct {
		name         string
		preferenceID string
		localization *PreferenceLocalization
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "successful creation",
			preferenceID: "pref1",
			localization: &PreferenceLocalization{
				Tag:         "en",
				Label:       "Test",
				Description: "A test preference",
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:         "empty preference ID",
			preferenceID: "",
			localization: &PreferenceLocalization{Tag: "en", Label: "Test"},
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "empty locale tag",
			preferenceID: "pref1",
			localization: &PreferenceLocalization{Tag: "", Label: "Test"},
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "nil localization",
			preferenceID: "pref1",
			localization: nil,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			preferenceID: "pref1",
			localization: &PreferenceLocalization{Tag: "en", Label: "Test"},
			statusCode:   http.StatusBadRequest,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"tag": "en", "label": "Test"}`))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.CreatePreferenceTranslations(context.Background(), tt.preferenceID, tt.localization)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetPreferenceTranslations(t *testing.T) {
	tests := []struct {
		name         string
		preferenceID string
		locale       string
		response     string
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "successful response",
			preferenceID: "pref1",
			locale:       "en",
			response:     `{"tag": "en", "label": "Test"}`,
			statusCode:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "empty preference ID",
			preferenceID: "",
			locale:       "en",
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "empty locale",
			preferenceID: "pref1",
			locale:       "",
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "invalid JSON response",
			preferenceID: "pref1",
			locale:       "en",
			response:     `{invalid json`,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			preferenceID: "pref1",
			locale:       "en",
			response:     `{"error": "not found"}`,
			statusCode:   http.StatusNotFound,
			wantErr:      true,
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
			_, err := client.GetPreferenceTranslations(context.Background(), tt.preferenceID, tt.locale)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_ListPreferenceTranslations(t *testing.T) {
	tests := []struct {
		name         string
		preferenceID string
		response     string
		statusCode   int
		wantErr      bool
		wantCount    int
	}{
		{
			name:         "successful response",
			preferenceID: "pref1",
			response: `{
				"items": [
					{"tag": "en", "label": "Test"},
					{"tag": "es", "label": "Prueba"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:         "empty preference ID",
			preferenceID: "",
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
			wantCount:    0,
		},
		{
			name:         "invalid JSON response",
			preferenceID: "pref1",
			response:     `{invalid json`,
			statusCode:   http.StatusOK,
			wantErr:      true,
			wantCount:    0,
		},
		{
			name:         "server error",
			preferenceID: "pref1",
			response:     `{"error": "not found"}`,
			statusCode:   http.StatusNotFound,
			wantErr:      true,
			wantCount:    0,
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
			translations, err := client.ListPreferenceTranslations(context.Background(), tt.preferenceID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(translations) != tt.wantCount {
				t.Errorf("expected %d translations, got %d", tt.wantCount, len(translations))
			}
		})
	}
}

func TestClient_UpdatePreferenceTranslations(t *testing.T) {
	tests := []struct {
		name         string
		preferenceID string
		localization *PreferenceLocalization
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "successful update",
			preferenceID: "pref1",
			localization: &PreferenceLocalization{
				Tag:   "en",
				Label: "Updated Test",
			},
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:         "empty preference ID",
			preferenceID: "",
			localization: &PreferenceLocalization{Tag: "en", Label: "Test"},
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "empty locale tag",
			preferenceID: "pref1",
			localization: &PreferenceLocalization{Tag: "", Label: "Test"},
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "nil localization",
			preferenceID: "pref1",
			localization: nil,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name:         "server error",
			preferenceID: "pref1",
			localization: &PreferenceLocalization{Tag: "en", Label: "Test"},
			statusCode:   http.StatusBadRequest,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"tag": "en", "label": "Updated Test"}`))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.UpdatePreferenceTranslations(context.Background(), tt.preferenceID, tt.localization)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
