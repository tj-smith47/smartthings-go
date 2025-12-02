package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListApps(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := appListResponse{
				Items: []App{
					{AppID: "app-1", AppName: "my-webhook-app", AppType: AppTypeWebhook},
					{AppID: "app-2", AppName: "my-lambda-app", AppType: AppTypeLambda},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		apps, err := client.ListApps(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 2 {
			t.Errorf("got %d apps, want 2", len(apps))
		}
		if apps[0].AppName != "my-webhook-app" {
			t.Errorf("AppName = %q, want %q", apps[0].AppName, "my-webhook-app")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(appListResponse{Items: []App{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		apps, err := client.ListApps(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(apps) != 0 {
			t.Errorf("got %d apps, want 0", len(apps))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListApps(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestClient_GetApp(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps/app-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps/app-123")
			}
			app := App{
				AppID:       "app-123",
				AppName:     "test-app",
				AppType:     AppTypeWebhook,
				DisplayName: "Test Application",
				Description: "A test SmartApp",
				WebhookSmartApp: &WebhookAppInfo{
					TargetURL:    "https://example.com/webhook",
					TargetStatus: "CONFIRMED",
				},
			}
			json.NewEncoder(w).Encode(app)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		app, err := client.GetApp(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if app.AppID != "app-123" {
			t.Errorf("AppID = %q, want %q", app.AppID, "app-123")
		}
		if app.WebhookSmartApp == nil {
			t.Fatal("expected WebhookSmartApp")
		}
		if app.WebhookSmartApp.TargetURL != "https://example.com/webhook" {
			t.Errorf("TargetURL = %q, want %q", app.WebhookSmartApp.TargetURL, "https://example.com/webhook")
		}
	})

	t.Run("empty app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetApp(context.Background(), "")
		if err != ErrEmptyAppID {
			t.Errorf("expected ErrEmptyAppID, got %v", err)
		}
	})

	t.Run("app not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetApp(context.Background(), "missing-app")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_CreateApp(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req AppCreate
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Errorf("failed to decode request: %v", err)
			}
			if req.AppName != "new-app" {
				t.Errorf("AppName = %q, want %q", req.AppName, "new-app")
			}
			if req.AppType != AppTypeWebhook {
				t.Errorf("AppType = %q, want %q", req.AppType, AppTypeWebhook)
			}

			app := App{
				AppID:   "app-new",
				AppName: req.AppName,
				AppType: req.AppType,
			}
			json.NewEncoder(w).Encode(app)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		app, err := client.CreateApp(context.Background(), &AppCreate{
			AppName: "new-app",
			AppType: AppTypeWebhook,
			WebhookSmartApp: &WebhookAppInfo{
				TargetURL: "https://example.com/webhook",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if app.AppID != "app-new" {
			t.Errorf("AppID = %q, want %q", app.AppID, "app-new")
		}
	})

	t.Run("nil app", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateApp(context.Background(), nil)
		if err != ErrEmptyAppName {
			t.Errorf("expected ErrEmptyAppName, got %v", err)
		}
	})

	t.Run("empty app name", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateApp(context.Background(), &AppCreate{})
		if err != ErrEmptyAppName {
			t.Errorf("expected ErrEmptyAppName, got %v", err)
		}
	})
}

func TestClient_UpdateApp(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps/app-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps/app-123")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req AppUpdate
			json.NewDecoder(r.Body).Decode(&req)

			app := App{
				AppID:       "app-123",
				AppName:     "existing-app",
				DisplayName: req.DisplayName,
			}
			json.NewEncoder(w).Encode(app)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		app, err := client.UpdateApp(context.Background(), "app-123", &AppUpdate{
			DisplayName: "Updated Name",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if app.DisplayName != "Updated Name" {
			t.Errorf("DisplayName = %q, want %q", app.DisplayName, "Updated Name")
		}
	})

	t.Run("empty app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateApp(context.Background(), "", &AppUpdate{})
		if err != ErrEmptyAppID {
			t.Errorf("expected ErrEmptyAppID, got %v", err)
		}
	})
}

func TestClient_DeleteApp(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps/app-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps/app-123")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteApp(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteApp(context.Background(), "")
		if err != ErrEmptyAppID {
			t.Errorf("expected ErrEmptyAppID, got %v", err)
		}
	})

	t.Run("app not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteApp(context.Background(), "missing-app")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_GetAppOAuth(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps/app-123/oauth" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps/app-123/oauth")
			}
			oauth := AppOAuth{
				ClientName:   "My SmartApp",
				Scope:        []string{"r:devices:*", "x:devices:*"},
				RedirectUris: []string{"https://example.com/callback"},
			}
			json.NewEncoder(w).Encode(oauth)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		oauth, err := client.GetAppOAuth(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if oauth.ClientName != "My SmartApp" {
			t.Errorf("ClientName = %q, want %q", oauth.ClientName, "My SmartApp")
		}
		if len(oauth.Scope) != 2 {
			t.Errorf("got %d scopes, want 2", len(oauth.Scope))
		}
	})

	t.Run("empty app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetAppOAuth(context.Background(), "")
		if err != ErrEmptyAppID {
			t.Errorf("expected ErrEmptyAppID, got %v", err)
		}
	})
}

func TestClient_UpdateAppOAuth(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps/app-123/oauth" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps/app-123/oauth")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req AppOAuth
			json.NewDecoder(r.Body).Decode(&req)

			oauth := AppOAuth{
				ClientName:   req.ClientName,
				Scope:        req.Scope,
				RedirectUris: req.RedirectUris,
			}
			json.NewEncoder(w).Encode(oauth)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		oauth, err := client.UpdateAppOAuth(context.Background(), "app-123", &AppOAuth{
			ClientName: "Updated App",
			Scope:      []string{"r:devices:*"},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if oauth.ClientName != "Updated App" {
			t.Errorf("ClientName = %q, want %q", oauth.ClientName, "Updated App")
		}
	})

	t.Run("empty app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateAppOAuth(context.Background(), "", &AppOAuth{})
		if err != ErrEmptyAppID {
			t.Errorf("expected ErrEmptyAppID, got %v", err)
		}
	})
}

func TestClient_GenerateAppOAuth(t *testing.T) {
	t.Run("successful generation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/apps/app-123/oauth/generate" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/apps/app-123/oauth/generate")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			generated := AppOAuthGenerated{
				ClientID:     "client-id-123",
				ClientSecret: "client-secret-456",
			}
			json.NewEncoder(w).Encode(generated)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		generated, err := client.GenerateAppOAuth(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if generated.ClientID != "client-id-123" {
			t.Errorf("ClientID = %q, want %q", generated.ClientID, "client-id-123")
		}
		if generated.ClientSecret != "client-secret-456" {
			t.Errorf("ClientSecret = %q, want %q", generated.ClientSecret, "client-secret-456")
		}
	})

	t.Run("empty app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GenerateAppOAuth(context.Background(), "")
		if err != ErrEmptyAppID {
			t.Errorf("expected ErrEmptyAppID, got %v", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GenerateAppOAuth(context.Background(), "app-123")
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}
