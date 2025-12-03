package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateNotification(t *testing.T) {
	tests := []struct {
		name         string
		notification *NotificationRequest
		response     string
		statusCode   int
		wantErr      bool
	}{
		{
			name: "successful creation",
			notification: &NotificationRequest{
				LocationID: "loc1",
				Type:       NotificationTypeAlert,
				Messages: map[string]NotificationMessage{
					"en": {Title: "Test", Body: "Test message"},
				},
			},
			response:   `{"code": 0, "message": "success"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:         "nil notification",
			notification: nil,
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
		{
			name: "empty type",
			notification: &NotificationRequest{
				LocationID: "loc1",
				Type:       "",
			},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "empty messages",
			notification: &NotificationRequest{
				LocationID: "loc1",
				Type:       NotificationTypeAlert,
				Messages:   nil,
			},
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "invalid JSON response",
			notification: &NotificationRequest{
				LocationID: "loc1",
				Type:       NotificationTypeAlert,
				Messages: map[string]NotificationMessage{
					"en": {Title: "Test", Body: "Test message"},
				},
			},
			response:   `{invalid json`,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "server error",
			notification: &NotificationRequest{
				LocationID: "loc1",
				Type:       NotificationTypeAlert,
				Messages: map[string]NotificationMessage{
					"en": {Title: "Test", Body: "Test message"},
				},
			},
			response:   ``,
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
			_, err := client.CreateNotification(context.Background(), tt.notification)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
