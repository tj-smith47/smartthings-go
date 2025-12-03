package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListDrivers(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name: "successful response",
			response: `{
				"items": [
					{"driverId": "driver1", "name": "Test Driver", "version": "1.0"},
					{"driverId": "driver2", "name": "Another Driver", "version": "2.0"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty list",
			response:   `{"items": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "server error",
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
			drivers, err := client.ListDrivers(context.Background())

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

func TestClient_ListDefaultDrivers(t *testing.T) {
	tests := []struct {
		name       string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name: "successful response",
			response: `{
				"items": [
					{"driverId": "default1", "name": "Default Driver 1"},
					{"driverId": "default2", "name": "Default Driver 2"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty list",
			response:   `{"items": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
		{
			name:       "invalid JSON response",
			response:   `{invalid json`,
			statusCode: http.StatusOK,
			wantErr:    true,
			wantCount:  0,
		},
		{
			name:       "server error",
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
			drivers, err := client.ListDefaultDrivers(context.Background())

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

func TestClient_GetDriver(t *testing.T) {
	tests := []struct {
		name       string
		driverID   string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			driverID:   "driver1",
			response:   `{"driverId": "driver1", "name": "Test Driver", "version": "1.0"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty driver ID",
			driverID:   "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "not found",
			driverID:   "nonexistent",
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
			_, err := client.GetDriver(context.Background(), tt.driverID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetDriverRevision(t *testing.T) {
	tests := []struct {
		name       string
		driverID   string
		version    string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			driverID:   "driver1",
			version:    "1.0",
			response:   `{"driverId": "driver1", "version": "1.0", "name": "Test Driver"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty driver ID",
			driverID:   "",
			version:    "1.0",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty version",
			driverID:   "driver1",
			version:    "",
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
			_, err := client.GetDriverRevision(context.Background(), tt.driverID, tt.version)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_DeleteDriver(t *testing.T) {
	tests := []struct {
		name       string
		driverID   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful deletion",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty driver ID",
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
			err := client.DeleteDriver(context.Background(), tt.driverID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_UploadDriver(t *testing.T) {
	tests := []struct {
		name       string
		driverData []byte
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful upload",
			driverData: []byte("driver package data"),
			response:   `{"driverId": "driver1", "version": "1.0"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil driver data",
			driverData: nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver data",
			driverData: []byte{},
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
			_, err := client.UploadDriver(context.Background(), tt.driverData)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
