package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListSubscriptions(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/subscriptions" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/subscriptions")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := subscriptionListResponse{
				Items: []Subscription{
					{ID: "sub-1", InstalledAppID: "app-123", SourceType: "DEVICE"},
					{ID: "sub-2", InstalledAppID: "app-123", SourceType: "CAPABILITY"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		subs, err := client.ListSubscriptions(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(subs) != 2 {
			t.Errorf("got %d subscriptions, want 2", len(subs))
		}
		if subs[0].SourceType != "DEVICE" {
			t.Errorf("subs[0].SourceType = %q, want %q", subs[0].SourceType, "DEVICE")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.ListSubscriptions(context.Background(), "")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(subscriptionListResponse{Items: []Subscription{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		subs, err := client.ListSubscriptions(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(subs) != 0 {
			t.Errorf("got %d subscriptions, want 0", len(subs))
		}
	})
}

func TestClient_CreateSubscription(t *testing.T) {
	t.Run("successful device subscription", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/subscriptions" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/subscriptions")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req SubscriptionCreate
			json.NewDecoder(r.Body).Decode(&req)
			if req.SourceType != "DEVICE" {
				t.Errorf("SourceType = %q, want %q", req.SourceType, "DEVICE")
			}
			if req.Device == nil {
				t.Fatal("Device is nil")
			}
			if req.Device.DeviceID != "device-456" {
				t.Errorf("Device.DeviceID = %q, want %q", req.Device.DeviceID, "device-456")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Subscription{
				ID:             "new-sub-123",
				InstalledAppID: "app-123",
				SourceType:     req.SourceType,
				Device:         req.Device,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		sub, err := client.CreateSubscription(context.Background(), "app-123", &SubscriptionCreate{
			SourceType: "DEVICE",
			Device: &DeviceSubscription{
				DeviceID:   "device-456",
				Capability: "switch",
				Attribute:  "switch",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sub.ID != "new-sub-123" {
			t.Errorf("ID = %q, want %q", sub.ID, "new-sub-123")
		}
	})

	t.Run("successful capability subscription", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req SubscriptionCreate
			json.NewDecoder(r.Body).Decode(&req)
			if req.SourceType != "CAPABILITY" {
				t.Errorf("SourceType = %q, want %q", req.SourceType, "CAPABILITY")
			}
			if req.Capability == nil {
				t.Fatal("Capability is nil")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Subscription{
				ID:         "cap-sub-123",
				SourceType: req.SourceType,
				Capability: req.Capability,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		sub, err := client.CreateSubscription(context.Background(), "app-123", &SubscriptionCreate{
			SourceType: "CAPABILITY",
			Capability: &CapabilitySubscription{
				LocationID:      "loc-123",
				Capability:      "motionSensor",
				Attribute:       "motion",
				StateChangeOnly: true,
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sub.ID != "cap-sub-123" {
			t.Errorf("ID = %q, want %q", sub.ID, "cap-sub-123")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateSubscription(context.Background(), "", &SubscriptionCreate{
			SourceType: "DEVICE",
			Device:     &DeviceSubscription{DeviceID: "d1"},
		})
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("nil subscription", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateSubscription(context.Background(), "app-123", nil)
		if err != ErrInvalidSubscription {
			t.Errorf("expected ErrInvalidSubscription, got %v", err)
		}
	})

	t.Run("empty source type", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateSubscription(context.Background(), "app-123", &SubscriptionCreate{
			SourceType: "",
		})
		if err != ErrInvalidSubscription {
			t.Errorf("expected ErrInvalidSubscription, got %v", err)
		}
	})
}

func TestClient_DeleteSubscription(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/subscriptions/sub-456" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/subscriptions/sub-456")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteSubscription(context.Background(), "app-123", "sub-456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteSubscription(context.Background(), "", "sub-456")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("empty subscription ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteSubscription(context.Background(), "app-123", "")
		if err != ErrEmptySubscriptionID {
			t.Errorf("expected ErrEmptySubscriptionID, got %v", err)
		}
	})
}

func TestClient_DeleteAllSubscriptions(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/subscriptions" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/subscriptions")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteAllSubscriptions(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteAllSubscriptions(context.Background(), "")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})
}
