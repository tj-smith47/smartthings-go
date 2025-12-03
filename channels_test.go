package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListChannels(t *testing.T) {
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
					{"channelId": "chan1", "name": "Test Channel", "type": "DRIVER"},
					{"channelId": "chan2", "name": "Another Channel", "type": "DRIVER"}
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
			channels, err := client.ListChannels(context.Background(), nil)

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

func TestClient_ListChannelsWithOptions(t *testing.T) {
	t.Run("with subscriber type option", func(t *testing.T) {
		var capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.String()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items": []}`))
		}))
		defer server.Close()

		client, _ := NewClient("test-token", WithBaseURL(server.URL))
		_, err := client.ListChannels(context.Background(), &ChannelListOptions{
			SubscriberType: SubscriberTypeHub,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if capturedPath != "/channels?subscriberType=HUB" {
			t.Errorf("expected path with subscriberType=HUB, got %s", capturedPath)
		}
	})

	t.Run("with subscriber ID option", func(t *testing.T) {
		var capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.String()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items": []}`))
		}))
		defer server.Close()

		client, _ := NewClient("test-token", WithBaseURL(server.URL))
		_, err := client.ListChannels(context.Background(), &ChannelListOptions{
			SubscriberID: "hub123",
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if capturedPath != "/channels?subscriberId=hub123" {
			t.Errorf("expected path with subscriberId=hub123, got %s", capturedPath)
		}
	})

	t.Run("with include read only option", func(t *testing.T) {
		var capturedPath string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedPath = r.URL.String()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items": []}`))
		}))
		defer server.Close()

		client, _ := NewClient("test-token", WithBaseURL(server.URL))
		_, err := client.ListChannels(context.Background(), &ChannelListOptions{
			IncludeReadOnly: true,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		if capturedPath != "/channels?includeReadOnly=true" {
			t.Errorf("expected path with includeReadOnly=true, got %s", capturedPath)
		}
	})

	t.Run("with all options", func(t *testing.T) {
		var capturedURL string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedURL = r.URL.String()
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items": []}`))
		}))
		defer server.Close()

		client, _ := NewClient("test-token", WithBaseURL(server.URL))
		_, err := client.ListChannels(context.Background(), &ChannelListOptions{
			SubscriberType:  SubscriberTypeHub,
			SubscriberID:    "hub123",
			IncludeReadOnly: true,
		})

		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
		// Check that all options are in the URL (order may vary)
		if !containsSubstring(capturedURL, "subscriberType=HUB") {
			t.Errorf("expected subscriberType=HUB in URL, got %s", capturedURL)
		}
		if !containsSubstring(capturedURL, "subscriberId=hub123") {
			t.Errorf("expected subscriberId=hub123 in URL, got %s", capturedURL)
		}
		if !containsSubstring(capturedURL, "includeReadOnly=true") {
			t.Errorf("expected includeReadOnly=true in URL, got %s", capturedURL)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`not valid json`))
		}))
		defer server.Close()

		client, _ := NewClient("test-token", WithBaseURL(server.URL))
		_, err := client.ListChannels(context.Background(), nil)

		if err == nil {
			t.Error("expected error for invalid JSON")
		}
	})
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && stringContains(s, substr))
}

func stringContains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestClient_GetChannel(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			channelID:  "chan1",
			response:   `{"channelId": "chan1", "name": "Test Channel", "type": "DRIVER"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "not found",
			channelID:  "nonexistent",
			response:   `{"error": "not found"}`,
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "invalid JSON response",
			channelID:  "chan1",
			response:   `not json`,
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
			_, err := client.GetChannel(context.Background(), tt.channelID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_CreateChannel(t *testing.T) {
	tests := []struct {
		name       string
		channel    *ChannelCreate
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "successful creation",
			channel: &ChannelCreate{
				Name:        "New Channel",
				Description: "A test channel",
				Type:        "DRIVER",
			},
			response:   `{"channelId": "chan1", "name": "New Channel", "type": "DRIVER"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil channel",
			channel:    nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "empty name",
			channel: &ChannelCreate{
				Name: "",
				Type: "DRIVER",
			},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "invalid JSON response",
			channel: &ChannelCreate{
				Name: "Test",
				Type: "DRIVER",
			},
			response:   `not json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "server error",
			channel: &ChannelCreate{
				Name: "Test",
				Type: "DRIVER",
			},
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
			_, err := client.CreateChannel(context.Background(), tt.channel)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_UpdateChannel(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		update     *ChannelUpdate
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:      "successful update",
			channelID: "chan1",
			update: &ChannelUpdate{
				Name:        "Updated Channel",
				Description: "Updated description",
			},
			response:   `{"channelId": "chan1", "name": "Updated Channel"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
			update:     &ChannelUpdate{Name: "Test"},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:      "invalid JSON response",
			channelID: "chan1",
			update:    &ChannelUpdate{Name: "Test"},
			response:  `not json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "server error",
			channelID:  "chan1",
			update:     &ChannelUpdate{Name: "Test"},
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
			_, err := client.UpdateChannel(context.Background(), tt.channelID, tt.update)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_DeleteChannel(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful deletion",
			channelID:  "chan1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
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
			err := client.DeleteChannel(context.Background(), tt.channelID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_ListAssignedDrivers(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:      "successful response",
			channelID: "chan1",
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
			name:       "empty channel ID",
			channelID:  "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
			wantCount:  0,
		},
		{
			name:       "invalid JSON response",
			channelID:  "chan1",
			response:   `not json`,
			statusCode: http.StatusOK,
			wantErr:    true,
			wantCount:  0,
		},
		{
			name:       "server error",
			channelID:  "chan1",
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
			drivers, err := client.ListAssignedDrivers(context.Background(), tt.channelID)

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

func TestClient_AssignDriver(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		driverID   string
		version    string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful assignment",
			channelID:  "chan1",
			driverID:   "driver1",
			version:    "1.0",
			response:   `{"driverId": "driver1", "channelId": "chan1"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
			driverID:   "driver1",
			version:    "1.0",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver ID",
			channelID:  "chan1",
			driverID:   "",
			version:    "1.0",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "invalid JSON response",
			channelID:  "chan1",
			driverID:   "driver1",
			version:    "1.0",
			response:   `not json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "server error",
			channelID:  "chan1",
			driverID:   "driver1",
			version:    "1.0",
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
			_, err := client.AssignDriver(context.Background(), tt.channelID, tt.driverID, tt.version)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_UnassignDriver(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		driverID   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful unassignment",
			channelID:  "chan1",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
			driverID:   "driver1",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver ID",
			channelID:  "chan1",
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
			err := client.UnassignDriver(context.Background(), tt.channelID, tt.driverID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_EnrollHub(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		hubID      string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful enrollment",
			channelID:  "chan1",
			hubID:      "hub1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
			hubID:      "hub1",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty hub ID",
			channelID:  "chan1",
			hubID:      "",
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
			err := client.EnrollHub(context.Background(), tt.channelID, tt.hubID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_UnenrollHub(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		hubID      string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful unenrollment",
			channelID:  "chan1",
			hubID:      "hub1",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
			hubID:      "hub1",
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty hub ID",
			channelID:  "chan1",
			hubID:      "",
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
			err := client.UnenrollHub(context.Background(), tt.channelID, tt.hubID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetDriverChannelMetaInfo(t *testing.T) {
	tests := []struct {
		name       string
		channelID  string
		driverID   string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			channelID:  "chan1",
			driverID:   "driver1",
			response:   `{"driverId": "driver1", "channelId": "chan1", "version": "1.0.0"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty channel ID",
			channelID:  "",
			driverID:   "driver1",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "empty driver ID",
			channelID:  "chan1",
			driverID:   "",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "invalid JSON response",
			channelID:  "chan1",
			driverID:   "driver1",
			response:   `not json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name:       "server error",
			channelID:  "chan1",
			driverID:   "driver1",
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
			_, err := client.GetDriverChannelMetaInfo(context.Background(), tt.channelID, tt.driverID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
