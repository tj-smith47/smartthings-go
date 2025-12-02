package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListDeviceProfiles(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/deviceprofiles" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/deviceprofiles")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := profileListResponse{
				Items: []DeviceProfileFull{
					{
						ID:     "profile-1",
						Name:   "custom-switch",
						Status: ProfileStatusDevelopment,
						Components: []ProfileComponent{
							{
								ID: "main",
								Capabilities: []CapabilityRef{
									{ID: "switch", Version: 1},
								},
							},
						},
					},
					{
						ID:     "profile-2",
						Name:   "temperature-sensor",
						Status: ProfileStatusPublished,
					},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		profiles, err := client.ListDeviceProfiles(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 2 {
			t.Errorf("got %d profiles, want 2", len(profiles))
		}
		if profiles[0].Name != "custom-switch" {
			t.Errorf("Name = %q, want %q", profiles[0].Name, "custom-switch")
		}
		if profiles[0].Status != ProfileStatusDevelopment {
			t.Errorf("Status = %q, want %q", profiles[0].Status, ProfileStatusDevelopment)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(profileListResponse{Items: []DeviceProfileFull{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		profiles, err := client.ListDeviceProfiles(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(profiles) != 0 {
			t.Errorf("got %d profiles, want 0", len(profiles))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListDeviceProfiles(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListDeviceProfiles(context.Background())
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}

func TestClient_GetDeviceProfile(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/deviceprofiles/profile-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/deviceprofiles/profile-123")
			}
			profile := DeviceProfileFull{
				ID:     "profile-123",
				Name:   "smart-dimmer",
				Status: ProfileStatusPublished,
				Components: []ProfileComponent{
					{
						ID:    "main",
						Label: "Main Component",
						Capabilities: []CapabilityRef{
							{ID: "switch", Version: 1},
							{ID: "switchLevel", Version: 1},
						},
						Categories: []string{"Light"},
					},
				},
				Metadata: map[string]string{
					"vid":  "custom-vid",
					"mnmn": "manufacturer",
				},
				Preferences: []ProfilePreference{
					{
						Name:        "fadeSpeed",
						Title:       "Fade Speed",
						Description: "How fast the light fades",
						Type:        "integer",
						Required:    false,
						Default:     5,
					},
				},
			}
			json.NewEncoder(w).Encode(profile)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		profile, err := client.GetDeviceProfile(context.Background(), "profile-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.ID != "profile-123" {
			t.Errorf("ID = %q, want %q", profile.ID, "profile-123")
		}
		if profile.Name != "smart-dimmer" {
			t.Errorf("Name = %q, want %q", profile.Name, "smart-dimmer")
		}
		if len(profile.Components) != 1 {
			t.Fatalf("got %d components, want 1", len(profile.Components))
		}
		if len(profile.Components[0].Capabilities) != 2 {
			t.Errorf("got %d capabilities, want 2", len(profile.Components[0].Capabilities))
		}
		if len(profile.Preferences) != 1 {
			t.Errorf("got %d preferences, want 1", len(profile.Preferences))
		}
	})

	t.Run("empty profile ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetDeviceProfile(context.Background(), "")
		if err != ErrEmptyProfileID {
			t.Errorf("expected ErrEmptyProfileID, got %v", err)
		}
	})

	t.Run("profile not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetDeviceProfile(context.Background(), "missing-profile")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_CreateDeviceProfile(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/deviceprofiles" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/deviceprofiles")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req DeviceProfileCreate
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}
			if req.Name != "new-profile" {
				t.Errorf("Name = %q, want %q", req.Name, "new-profile")
			}

			profile := DeviceProfileFull{
				ID:         "profile-new",
				Name:       req.Name,
				Status:     ProfileStatusDevelopment,
				Components: req.Components,
			}
			json.NewEncoder(w).Encode(profile)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		profile, err := client.CreateDeviceProfile(context.Background(), &DeviceProfileCreate{
			Name: "new-profile",
			Components: []ProfileComponent{
				{
					ID: "main",
					Capabilities: []CapabilityRef{
						{ID: "switch", Version: 1},
					},
				},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.ID != "profile-new" {
			t.Errorf("ID = %q, want %q", profile.ID, "profile-new")
		}
		if profile.Status != ProfileStatusDevelopment {
			t.Errorf("Status = %q, want %q", profile.Status, ProfileStatusDevelopment)
		}
	})

	t.Run("nil profile", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateDeviceProfile(context.Background(), nil)
		if err != ErrEmptyProfileName {
			t.Errorf("expected ErrEmptyProfileName, got %v", err)
		}
	})

	t.Run("empty profile name", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateDeviceProfile(context.Background(), &DeviceProfileCreate{})
		if err != ErrEmptyProfileName {
			t.Errorf("expected ErrEmptyProfileName, got %v", err)
		}
	})
}

func TestClient_UpdateDeviceProfile(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/deviceprofiles/profile-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/deviceprofiles/profile-123")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req DeviceProfileUpdate
			json.NewDecoder(r.Body).Decode(&req)

			profile := DeviceProfileFull{
				ID:         "profile-123",
				Name:       "updated-profile",
				Status:     ProfileStatusDevelopment,
				Components: req.Components,
				Metadata:   req.Metadata,
			}
			json.NewEncoder(w).Encode(profile)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		profile, err := client.UpdateDeviceProfile(context.Background(), "profile-123", &DeviceProfileUpdate{
			Components: []ProfileComponent{
				{
					ID: "main",
					Capabilities: []CapabilityRef{
						{ID: "switch", Version: 1},
						{ID: "switchLevel", Version: 1},
					},
				},
			},
			Metadata: map[string]string{"version": "2.0"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if profile.ID != "profile-123" {
			t.Errorf("ID = %q, want %q", profile.ID, "profile-123")
		}
		if len(profile.Components) != 1 {
			t.Errorf("got %d components, want 1", len(profile.Components))
		}
	})

	t.Run("empty profile ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateDeviceProfile(context.Background(), "", &DeviceProfileUpdate{})
		if err != ErrEmptyProfileID {
			t.Errorf("expected ErrEmptyProfileID, got %v", err)
		}
	})
}

func TestClient_DeleteDeviceProfile(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/deviceprofiles/profile-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/deviceprofiles/profile-123")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteDeviceProfile(context.Background(), "profile-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty profile ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteDeviceProfile(context.Background(), "")
		if err != ErrEmptyProfileID {
			t.Errorf("expected ErrEmptyProfileID, got %v", err)
		}
	})

	t.Run("profile not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteDeviceProfile(context.Background(), "missing-profile")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteDeviceProfile(context.Background(), "profile-123")
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}
