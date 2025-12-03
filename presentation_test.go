package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GeneratePresentation(t *testing.T) {
	tests := []struct {
		name       string
		deviceID   string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful generation",
			deviceID:   "device1",
			response:   `{"presentationId": "pres1", "type": "default"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty device ID",
			deviceID:   "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "invalid JSON response",
			deviceID:   "device1",
			response:   `{invalid json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "server error",
			deviceID:   "device1",
			response:   `{"error": "internal error"}`,
			statusCode: http.StatusInternalServerError,
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
			_, err := client.GeneratePresentation(context.Background(), tt.deviceID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetPresentationConfig(t *testing.T) {
	tests := []struct {
		name             string
		presentationID   string
		manufacturerName string
		response         string
		statusCode       int
		wantErr          bool
	}{
		{
			name:             "successful response",
			presentationID:   "pres1",
			manufacturerName: "",
			response:         `{"id": "pres1", "mnId": "mn1", "vid": "vid1"}`,
			statusCode:       http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "with manufacturer name",
			presentationID:   "pres1",
			manufacturerName: "Samsung",
			response:         `{"id": "pres1", "manufacturerName": "Samsung"}`,
			statusCode:       http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "empty presentation ID",
			presentationID:   "",
			manufacturerName: "",
			response:         ``,
			statusCode:       http.StatusOK,
			wantErr:          true,
		},
		{
			name:             "invalid JSON response",
			presentationID:   "pres1",
			manufacturerName: "",
			response:         `{invalid json`,
			statusCode:       http.StatusOK,
			wantErr:          true,
		},
		{
			name:             "server error",
			presentationID:   "pres1",
			manufacturerName: "",
			response:         `{"error": "not found"}`,
			statusCode:       http.StatusNotFound,
			wantErr:          true,
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
			_, err := client.GetPresentationConfig(context.Background(), tt.presentationID, tt.manufacturerName)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CreatePresentationConfig(t *testing.T) {
	tests := []struct {
		name       string
		config     *PresentationDeviceConfigCreate
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "successful creation",
			config: &PresentationDeviceConfigCreate{
				Type: "profile",
			},
			response:   `{"manufacturerName": "mn1", "presentationId": "pres1"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil config",
			config:     nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "invalid JSON response",
			config: &PresentationDeviceConfigCreate{
				Type: "profile",
			},
			response:   `{invalid json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "server error",
			config: &PresentationDeviceConfigCreate{
				Type: "profile",
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
			_, err := client.CreatePresentationConfig(context.Background(), tt.config)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetDevicePresentation(t *testing.T) {
	tests := []struct {
		name             string
		presentationID   string
		manufacturerName string
		response         string
		statusCode       int
		wantErr          bool
	}{
		{
			name:             "successful response",
			presentationID:   "pres1",
			manufacturerName: "TestMfg",
			response:         `{"presentationId": "pres1", "manufacturerName": "TestMfg"}`,
			statusCode:       http.StatusOK,
			wantErr:          false,
		},
		{
			name:             "empty presentation ID",
			presentationID:   "",
			manufacturerName: "TestMfg",
			response:         ``,
			statusCode:       http.StatusOK,
			wantErr:          true,
		},
		{
			name:             "invalid JSON response",
			presentationID:   "pres1",
			manufacturerName: "TestMfg",
			response:         `{invalid json`,
			statusCode:       http.StatusOK,
			wantErr:          true,
		},
		{
			name:             "server error",
			presentationID:   "pres1",
			manufacturerName: "TestMfg",
			response:         `{"error": "not found"}`,
			statusCode:       http.StatusNotFound,
			wantErr:          true,
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
			_, err := client.GetDevicePresentation(context.Background(), tt.presentationID, tt.manufacturerName)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
