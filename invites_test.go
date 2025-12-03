package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_CreateSchemaAppInvitation(t *testing.T) {
	tests := []struct {
		name       string
		invitation *SchemaAppInvitationCreate
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name: "successful creation",
			invitation: &SchemaAppInvitationCreate{
				SchemaAppID: "app1",
				Description: "Test invitation",
			},
			response:   `{"invitationId": "inv1"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "nil invitation",
			invitation: nil,
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
		{
			name: "empty app ID",
			invitation: &SchemaAppInvitationCreate{
				SchemaAppID: "",
				Description: "Test",
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
			_, err := client.CreateSchemaAppInvitation(context.Background(), tt.invitation)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_ListSchemaAppInvitations(t *testing.T) {
	tests := []struct {
		name       string
		appID      string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:  "successful response",
			appID: "app1",
			response: `{
				"items": [
					{"invitationId": "inv1"},
					{"invitationId": "inv2"}
				]
			}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  2,
		},
		{
			name:       "empty app ID",
			appID:      "",
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
			invites, err := client.ListSchemaAppInvitations(context.Background(), tt.appID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(invites) != tt.wantCount {
				t.Errorf("expected %d invites, got %d", tt.wantCount, len(invites))
			}
		})
	}
}

func TestClient_RevokeSchemaAppInvitation(t *testing.T) {
	tests := []struct {
		name         string
		invitationID string
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "successful revocation",
			invitationID: "inv1",
			statusCode:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "empty invitation ID",
			invitationID: "",
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			err := client.RevokeSchemaAppInvitation(context.Background(), tt.invitationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
