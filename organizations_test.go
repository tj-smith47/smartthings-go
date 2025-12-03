package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListOrganizations(t *testing.T) {
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
					{"organizationId": "org1", "name": "Org One"},
					{"organizationId": "org2", "name": "Org Two"}
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
			orgs, err := client.ListOrganizations(context.Background())

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(orgs) != tt.wantCount {
				t.Errorf("expected %d orgs, got %d", tt.wantCount, len(orgs))
			}
		})
	}
}

func TestClient_GetOrganization(t *testing.T) {
	tests := []struct {
		name           string
		organizationID string
		response       string
		statusCode     int
		wantErr        bool
	}{
		{
			name:           "successful response",
			organizationID: "org1",
			response:       `{"organizationId": "org1", "name": "Org One", "warehouseGroupId": "wg1"}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty organization ID",
			organizationID: "",
			response:       ``,
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
		{
			name:           "not found",
			organizationID: "nonexistent",
			response:       `{"error": "not found"}`,
			statusCode:     http.StatusNotFound,
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
			_, err := client.GetOrganization(context.Background(), tt.organizationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
