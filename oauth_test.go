package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetAuthorizationURL(t *testing.T) {
	tests := []struct {
		name     string
		config   *OAuthConfig
		state    string
		contains []string
	}{
		{
			name: "basic URL with default scopes",
			config: &OAuthConfig{
				ClientID:    "test-client-id",
				RedirectURL: "http://localhost:8080/callback",
			},
			state: "test-state",
			contains: []string{
				"https://api.smartthings.com/oauth/authorize",
				"client_id=test-client-id",
				"redirect_uri=http%3A%2F%2Flocalhost%3A8080%2Fcallback",
				"state=test-state",
				"response_type=code",
				"r%3Adevices%3A%2A",   // r:devices:*
				"x%3Adevices%3A%2A",   // x:devices:*
				"r%3Alocations%3A%2A", // r:locations:*
			},
		},
		{
			name: "custom scopes",
			config: &OAuthConfig{
				ClientID:    "test-client",
				RedirectURL: "https://example.com/callback",
				Scopes:      []string{"r:devices:*"},
			},
			state: "",
			contains: []string{
				"client_id=test-client",
				"r%3Adevices%3A%2A",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			url := GetAuthorizationURL(tt.config, tt.state)
			for _, expected := range tt.contains {
				if !contains(url, expected) {
					t.Errorf("URL missing expected substring %q\nURL: %s", expected, url)
				}
			}
		})
	}
}

func TestExchangeCode(t *testing.T) {
	t.Run("successful exchange", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("expected application/x-www-form-urlencoded content type")
			}

			r.ParseForm()
			if r.FormValue("grant_type") != "authorization_code" {
				t.Errorf("expected grant_type=authorization_code, got %s", r.FormValue("grant_type"))
			}
			if r.FormValue("code") != "test-code" {
				t.Errorf("expected code=test-code, got %s", r.FormValue("code"))
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "test-access-token",
				"refresh_token": "test-refresh-token",
				"expires_in":    3600,
				"token_type":    "Bearer",
				"scope":         "r:devices:* x:devices:*",
			})
		}))
		defer server.Close()

		// Override token endpoint for testing
		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
			RedirectURL:  "http://localhost/callback",
		}

		tokens, err := ExchangeCode(context.Background(), cfg, "test-code")
		if err != nil {
			t.Fatalf("ExchangeCode failed: %v", err)
		}

		if tokens.AccessToken != "test-access-token" {
			t.Errorf("expected access_token 'test-access-token', got %q", tokens.AccessToken)
		}
		if tokens.RefreshToken != "test-refresh-token" {
			t.Errorf("expected refresh_token 'test-refresh-token', got %q", tokens.RefreshToken)
		}
		if tokens.ExpiresIn != 3600 {
			t.Errorf("expected expires_in 3600, got %d", tokens.ExpiresIn)
		}
		if tokens.ExpiresAt.IsZero() {
			t.Error("ExpiresAt should be set")
		}
	})

	t.Run("empty code returns error", func(t *testing.T) {
		cfg := &OAuthConfig{
			ClientID:     "test",
			ClientSecret: "test",
			RedirectURL:  "http://localhost/callback",
		}
		_, err := ExchangeCode(context.Background(), cfg, "")
		if err == nil {
			t.Error("expected error for empty code")
		}
	})

	t.Run("server error with OAuth error response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Authorization code expired",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "test",
			ClientSecret: "test",
			RedirectURL:  "http://localhost/callback",
		}

		_, err := ExchangeCode(context.Background(), cfg, "expired-code")
		if err == nil {
			t.Fatal("expected error for OAuth error response")
		}
		if !containsHelper(err.Error(), "invalid_grant") {
			t.Errorf("error should contain 'invalid_grant', got: %v", err)
		}
	})

	t.Run("server error with non-JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Internal Server Error"))
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "test",
			ClientSecret: "test",
			RedirectURL:  "http://localhost/callback",
		}

		_, err := ExchangeCode(context.Background(), cfg, "some-code")
		if err == nil {
			t.Fatal("expected error for server error")
		}
		if !containsHelper(err.Error(), "500") {
			t.Errorf("error should contain status code, got: %v", err)
		}
	})

	t.Run("invalid JSON response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("{invalid json"))
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "test",
			ClientSecret: "test",
			RedirectURL:  "http://localhost/callback",
		}

		_, err := ExchangeCode(context.Background(), cfg, "some-code")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestRefreshTokens(t *testing.T) {
	t.Run("successful refresh", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodPost {
				t.Errorf("expected POST, got %s", r.Method)
			}

			r.ParseForm()
			if r.FormValue("grant_type") != "refresh_token" {
				t.Errorf("expected grant_type=refresh_token, got %s", r.FormValue("grant_type"))
			}
			if r.FormValue("refresh_token") != "old-refresh-token" {
				t.Errorf("expected refresh_token=old-refresh-token, got %s", r.FormValue("refresh_token"))
			}

			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "new-access-token",
				"refresh_token": "new-refresh-token",
				"expires_in":    7200,
				"token_type":    "Bearer",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "test-client-id",
			ClientSecret: "test-client-secret",
		}

		tokens, err := RefreshTokens(context.Background(), cfg, "old-refresh-token")
		if err != nil {
			t.Fatalf("RefreshTokens failed: %v", err)
		}

		if tokens.AccessToken != "new-access-token" {
			t.Errorf("expected access_token 'new-access-token', got %q", tokens.AccessToken)
		}
		if tokens.RefreshToken != "new-refresh-token" {
			t.Errorf("expected refresh_token 'new-refresh-token', got %q", tokens.RefreshToken)
		}
		if tokens.ExpiresIn != 7200 {
			t.Errorf("expected expires_in 7200, got %d", tokens.ExpiresIn)
		}
	})

	t.Run("empty refresh token returns error", func(t *testing.T) {
		cfg := &OAuthConfig{
			ClientID:     "test",
			ClientSecret: "test",
		}
		_, err := RefreshTokens(context.Background(), cfg, "")
		if err == nil {
			t.Error("expected error for empty refresh token")
		}
	})

	t.Run("server returns invalid_grant for expired refresh token", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Refresh token expired",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "test",
			ClientSecret: "test",
		}

		_, err := RefreshTokens(context.Background(), cfg, "expired-refresh-token")
		if err == nil {
			t.Fatal("expected error for expired refresh token")
		}
		if !containsHelper(err.Error(), "invalid_grant") {
			t.Errorf("error should contain 'invalid_grant', got: %v", err)
		}
	})

	t.Run("context cancellation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			time.Sleep(100 * time.Millisecond)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		cfg := &OAuthConfig{
			ClientID:     "test",
			ClientSecret: "test",
		}

		_, err := RefreshTokens(ctx, cfg, "some-refresh-token")
		if err == nil {
			t.Error("expected error for cancelled context")
		}
	})
}

func TestTokenResponse(t *testing.T) {
	t.Run("IsValid", func(t *testing.T) {
		tests := []struct {
			name     string
			token    *TokenResponse
			expected bool
		}{
			{
				name:     "nil token",
				token:    nil,
				expected: false,
			},
			{
				name: "empty access token",
				token: &TokenResponse{
					AccessToken: "",
					ExpiresAt:   time.Now().Add(time.Hour),
				},
				expected: false,
			},
			{
				name: "expired token",
				token: &TokenResponse{
					AccessToken: "test",
					ExpiresAt:   time.Now().Add(-time.Hour),
				},
				expected: false,
			},
			{
				name: "expires within buffer",
				token: &TokenResponse{
					AccessToken: "test",
					ExpiresAt:   time.Now().Add(3 * time.Minute), // Less than 5 minute buffer
				},
				expected: false,
			},
			{
				name: "valid token",
				token: &TokenResponse{
					AccessToken: "test",
					ExpiresAt:   time.Now().Add(time.Hour),
				},
				expected: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.token.IsValid(); got != tt.expected {
					t.Errorf("IsValid() = %v, want %v", got, tt.expected)
				}
			})
		}
	})

	t.Run("IsRefreshTokenValid", func(t *testing.T) {
		tests := []struct {
			name     string
			token    *TokenResponse
			expected bool
		}{
			{
				name:     "nil token",
				token:    nil,
				expected: false,
			},
			{
				name: "empty refresh token",
				token: &TokenResponse{
					RefreshToken: "",
				},
				expected: false,
			},
			{
				name: "no expiry set (never expires)",
				token: &TokenResponse{
					RefreshToken: "test",
				},
				expected: true,
			},
			{
				name: "expired refresh token",
				token: &TokenResponse{
					RefreshToken:          "test",
					RefreshTokenExpiresAt: time.Now().Add(-time.Hour),
				},
				expected: false,
			},
			{
				name: "valid refresh token",
				token: &TokenResponse{
					RefreshToken:          "test",
					RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
				},
				expected: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.token.IsRefreshTokenValid(); got != tt.expected {
					t.Errorf("IsRefreshTokenValid() = %v, want %v", got, tt.expected)
				}
			})
		}
	})

	t.Run("NeedsRefresh", func(t *testing.T) {
		tests := []struct {
			name     string
			token    *TokenResponse
			expected bool
		}{
			{
				name:     "nil token",
				token:    nil,
				expected: true,
			},
			{
				name: "expired token",
				token: &TokenResponse{
					AccessToken: "test",
					ExpiresAt:   time.Now().Add(-time.Hour),
				},
				expected: true,
			},
			{
				name: "valid token",
				token: &TokenResponse{
					AccessToken: "test",
					ExpiresAt:   time.Now().Add(time.Hour),
				},
				expected: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				if got := tt.token.NeedsRefresh(); got != tt.expected {
					t.Errorf("NeedsRefresh() = %v, want %v", got, tt.expected)
				}
			})
		}
	})
}

func TestFileTokenStore(t *testing.T) {
	t.Run("SaveTokens nil returns error", func(t *testing.T) {
		tmpDir := t.TempDir()
		filepath := filepath.Join(tmpDir, "tokens.json")

		store := NewFileTokenStore(filepath)
		err := store.SaveTokens(context.Background(), nil)
		if err == nil {
			t.Error("expected error for nil tokens")
		}
	})

	t.Run("SaveTokens and LoadTokens", func(t *testing.T) {
		tmpDir := t.TempDir()
		filepath := filepath.Join(tmpDir, "tokens.json")

		store := NewFileTokenStore(filepath)
		ctx := context.Background()

		// Save tokens
		tokens := &TokenResponse{
			AccessToken:  "test-access",
			RefreshToken: "test-refresh",
			ExpiresIn:    3600,
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().Add(time.Hour),
		}

		err := store.SaveTokens(ctx, tokens)
		if err != nil {
			t.Fatalf("SaveTokens failed: %v", err)
		}

		// Load tokens
		loaded, err := store.LoadTokens(ctx)
		if err != nil {
			t.Fatalf("LoadTokens failed: %v", err)
		}

		if loaded.AccessToken != tokens.AccessToken {
			t.Errorf("AccessToken mismatch: got %q, want %q", loaded.AccessToken, tokens.AccessToken)
		}
		if loaded.RefreshToken != tokens.RefreshToken {
			t.Errorf("RefreshToken mismatch: got %q, want %q", loaded.RefreshToken, tokens.RefreshToken)
		}
	})

	t.Run("LoadTokens returns error for missing file", func(t *testing.T) {
		store := NewFileTokenStore("/nonexistent/path/tokens.json")
		_, err := store.LoadTokens(context.Background())
		if err == nil {
			t.Error("expected error for missing file")
		}
	})

	t.Run("SaveTokens creates directory", func(t *testing.T) {
		tmpDir := t.TempDir()
		filepath := filepath.Join(tmpDir, "subdir", "tokens.json")

		store := NewFileTokenStore(filepath)
		err := store.SaveTokens(context.Background(), &TokenResponse{
			AccessToken: "test",
		})
		if err != nil {
			t.Fatalf("SaveTokens failed: %v", err)
		}

		// Verify file exists
		if _, err := os.Stat(filepath); os.IsNotExist(err) {
			t.Error("token file was not created")
		}
	})

	t.Run("Exists", func(t *testing.T) {
		tmpDir := t.TempDir()
		filepath := filepath.Join(tmpDir, "tokens.json")

		store := NewFileTokenStore(filepath)

		// File doesn't exist yet
		if store.Exists() {
			t.Error("Exists() should return false for nonexistent file")
		}

		// Create the file
		store.SaveTokens(context.Background(), &TokenResponse{AccessToken: "test"})

		// Now it should exist
		if !store.Exists() {
			t.Error("Exists() should return true after saving")
		}
	})

	t.Run("Delete", func(t *testing.T) {
		tmpDir := t.TempDir()
		path := filepath.Join(tmpDir, "tokens.json")

		store := NewFileTokenStore(path)
		ctx := context.Background()

		// Create the file
		store.SaveTokens(ctx, &TokenResponse{AccessToken: "test"})

		// Delete it
		err := store.Delete(ctx)
		if err != nil {
			t.Fatalf("Delete failed: %v", err)
		}

		// Verify it's gone
		if store.Exists() {
			t.Error("file should be deleted")
		}
	})
}

func TestMemoryTokenStore(t *testing.T) {
	t.Run("SaveTokens and LoadTokens", func(t *testing.T) {
		store := NewMemoryTokenStore()
		ctx := context.Background()

		tokens := &TokenResponse{
			AccessToken:  "test-access",
			RefreshToken: "test-refresh",
		}

		err := store.SaveTokens(ctx, tokens)
		if err != nil {
			t.Fatalf("SaveTokens failed: %v", err)
		}

		loaded, err := store.LoadTokens(ctx)
		if err != nil {
			t.Fatalf("LoadTokens failed: %v", err)
		}

		if loaded.AccessToken != tokens.AccessToken {
			t.Errorf("AccessToken mismatch")
		}
	})

	t.Run("LoadTokens returns error when empty", func(t *testing.T) {
		store := NewMemoryTokenStore()
		_, err := store.LoadTokens(context.Background())
		if err == nil {
			t.Error("expected error for empty store")
		}
	})

	t.Run("Clear", func(t *testing.T) {
		store := NewMemoryTokenStore()
		ctx := context.Background()

		store.SaveTokens(ctx, &TokenResponse{AccessToken: "test"})
		store.Clear()

		_, err := store.LoadTokens(ctx)
		if err == nil {
			t.Error("expected error after clear")
		}
	})
}

func TestNewOAuthClient(t *testing.T) {
	t.Run("requires config", func(t *testing.T) {
		_, err := NewOAuthClient(nil, NewMemoryTokenStore())
		if err == nil {
			t.Error("expected error for nil config")
		}
	})

	t.Run("requires client ID", func(t *testing.T) {
		_, err := NewOAuthClient(&OAuthConfig{
			ClientSecret: "secret",
		}, NewMemoryTokenStore())
		if err == nil {
			t.Error("expected error for empty client ID")
		}
	})

	t.Run("requires client secret", func(t *testing.T) {
		_, err := NewOAuthClient(&OAuthConfig{
			ClientID: "id",
		}, NewMemoryTokenStore())
		if err == nil {
			t.Error("expected error for empty client secret")
		}
	})

	t.Run("requires token store", func(t *testing.T) {
		_, err := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, nil)
		if err == nil {
			t.Error("expected error for nil token store")
		}
	})

	t.Run("creates client successfully", func(t *testing.T) {
		client, err := NewOAuthClient(&OAuthConfig{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
			RedirectURL:  "http://localhost/callback",
		}, NewMemoryTokenStore())

		if err != nil {
			t.Fatalf("NewOAuthClient failed: %v", err)
		}
		if client == nil {
			t.Error("expected non-nil client")
		}
	})

	t.Run("loads existing tokens from store", func(t *testing.T) {
		store := NewMemoryTokenStore()
		ctx := context.Background()

		// Pre-populate the store
		store.SaveTokens(ctx, &TokenResponse{
			AccessToken:  "existing-token",
			RefreshToken: "existing-refresh",
			ExpiresAt:    time.Now().Add(time.Hour),
		})

		client, err := NewOAuthClient(&OAuthConfig{
			ClientID:     "test-id",
			ClientSecret: "test-secret",
		}, store)

		if err != nil {
			t.Fatalf("NewOAuthClient failed: %v", err)
		}

		if !client.IsAuthenticated() {
			t.Error("client should be authenticated with existing tokens")
		}
	})
}

func TestOAuthClient_TokenManagement(t *testing.T) {
	t.Run("SetTokens and GetTokens", func(t *testing.T) {
		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		tokens := &TokenResponse{
			AccessToken:  "new-access",
			RefreshToken: "new-refresh",
			ExpiresAt:    time.Now().Add(time.Hour),
		}

		err := client.SetTokens(context.Background(), tokens)
		if err != nil {
			t.Fatalf("SetTokens failed: %v", err)
		}

		got := client.GetTokens()
		if got.AccessToken != tokens.AccessToken {
			t.Errorf("AccessToken mismatch")
		}

		// Verify underlying client has the token
		if client.Client.Token() != tokens.AccessToken {
			t.Error("underlying client token not updated")
		}

		// Verify tokens were persisted
		loaded, _ := store.LoadTokens(context.Background())
		if loaded.AccessToken != tokens.AccessToken {
			t.Error("tokens not persisted to store")
		}
	})

	t.Run("IsAuthenticated", func(t *testing.T) {
		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		// Initially not authenticated
		if client.IsAuthenticated() {
			t.Error("should not be authenticated initially")
		}

		// Set valid tokens
		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken: "token",
			ExpiresAt:   time.Now().Add(time.Hour),
		})

		if !client.IsAuthenticated() {
			t.Error("should be authenticated after setting tokens")
		}
	})

	t.Run("NeedsReauthentication", func(t *testing.T) {
		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		// No tokens - needs reauth
		if !client.NeedsReauthentication() {
			t.Error("should need reauth with no tokens")
		}

		// Valid refresh token
		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken:           "token",
			RefreshToken:          "refresh",
			ExpiresAt:             time.Now().Add(time.Hour),
			RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
		})

		if client.NeedsReauthentication() {
			t.Error("should not need reauth with valid refresh token")
		}

		// Expired refresh token
		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken:           "token",
			RefreshToken:          "refresh",
			ExpiresAt:             time.Now().Add(time.Hour),
			RefreshTokenExpiresAt: time.Now().Add(-time.Hour),
		})

		if !client.NeedsReauthentication() {
			t.Error("should need reauth with expired refresh token")
		}
	})

	t.Run("ClearTokens", func(t *testing.T) {
		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken: "token",
			ExpiresAt:   time.Now().Add(time.Hour),
		})

		err := client.ClearTokens(context.Background())
		if err != nil {
			t.Fatalf("ClearTokens failed: %v", err)
		}

		if client.IsAuthenticated() {
			t.Error("should not be authenticated after clearing")
		}
	})
}

func TestOAuthClient_GetAuthorizationURL(t *testing.T) {
	client, _ := NewOAuthClient(&OAuthConfig{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/callback",
	}, NewMemoryTokenStore())

	url := client.GetAuthorizationURL("test-state")

	if !contains(url, "client_id=test-id") {
		t.Error("URL missing client_id")
	}
	if !contains(url, "state=test-state") {
		t.Error("URL missing state")
	}
}

func TestOAuthClient_Config(t *testing.T) {
	cfg := &OAuthConfig{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		RedirectURL:  "http://localhost/callback",
	}
	client, _ := NewOAuthClient(cfg, NewMemoryTokenStore())

	got := client.Config()
	if got != cfg {
		t.Error("Config() should return the same config")
	}
	if got.ClientID != "test-id" {
		t.Error("Config() ClientID mismatch")
	}
}

func TestOAuthClient_TokenStore(t *testing.T) {
	store := NewMemoryTokenStore()
	client, _ := NewOAuthClient(&OAuthConfig{
		ClientID:     "test-id",
		ClientSecret: "test-secret",
	}, store)

	got := client.TokenStore()
	if got != store {
		t.Error("TokenStore() should return the same store")
	}
}

func TestOAuthClient_EnsureValidToken(t *testing.T) {
	t.Run("returns error when no tokens", func(t *testing.T) {
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, NewMemoryTokenStore())

		err := client.EnsureValidToken(context.Background())
		if err == nil {
			t.Error("expected error when no tokens")
		}
	})

	t.Run("succeeds when token is valid", func(t *testing.T) {
		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken:  "valid-token",
			RefreshToken: "refresh",
			ExpiresAt:    time.Now().Add(time.Hour),
		})

		err := client.EnsureValidToken(context.Background())
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
	})

	t.Run("returns error when refresh token expired", func(t *testing.T) {
		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken:           "expired-token",
			RefreshToken:          "expired-refresh",
			ExpiresAt:             time.Now().Add(-time.Hour),     // Access token expired
			RefreshTokenExpiresAt: time.Now().Add(-time.Hour * 2), // Refresh token also expired
		})

		err := client.EnsureValidToken(context.Background())
		if err == nil {
			t.Error("expected error when refresh token expired")
		}
	})

	t.Run("refreshes token when access token expired but refresh token valid", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "new-access-token",
				"refresh_token": "new-refresh-token",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		// Set tokens with expired access token but valid refresh token
		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken:           "expired-access-token",
			RefreshToken:          "valid-refresh-token",
			ExpiresAt:             time.Now().Add(-time.Hour), // Access token expired
			RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
		})

		err := client.EnsureValidToken(context.Background())
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}

		// Verify the token was refreshed
		tokens := client.GetTokens()
		if tokens.AccessToken != "new-access-token" {
			t.Errorf("expected new access token, got: %s", tokens.AccessToken)
		}
		if tokens.RefreshToken != "new-refresh-token" {
			t.Errorf("expected new refresh token, got: %s", tokens.RefreshToken)
		}
	})

	t.Run("returns error when token refresh fails", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Token has been revoked",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		// Set tokens with expired access token but valid refresh token
		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken:           "expired-access-token",
			RefreshToken:          "revoked-refresh-token",
			ExpiresAt:             time.Now().Add(-time.Hour), // Access token expired
			RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
		})

		err := client.EnsureValidToken(context.Background())
		if err == nil {
			t.Error("expected error when refresh fails")
		}
		if !containsHelper(err.Error(), "invalid_grant") {
			t.Errorf("error should contain 'invalid_grant', got: %v", err)
		}
	})

	t.Run("persists refreshed tokens to store", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "persisted-access-token",
				"refresh_token": "persisted-refresh-token",
				"expires_in":    3600,
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store)

		// Set expired tokens
		client.SetTokens(context.Background(), &TokenResponse{
			AccessToken:           "expired-access-token",
			RefreshToken:          "valid-refresh-token",
			ExpiresAt:             time.Now().Add(-time.Hour),
			RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
		})

		err := client.EnsureValidToken(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify tokens were persisted to store
		storedTokens, err := store.LoadTokens(context.Background())
		if err != nil {
			t.Fatalf("failed to load tokens from store: %v", err)
		}
		if storedTokens.AccessToken != "persisted-access-token" {
			t.Errorf("expected persisted access token, got: %s", storedTokens.AccessToken)
		}
	})
}

func TestOAuthClient_ExchangeCode(t *testing.T) {
	t.Run("returns error for empty code", func(t *testing.T) {
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, NewMemoryTokenStore())

		err := client.ExchangeCode(context.Background(), "")
		if err == nil {
			t.Error("expected error for empty code")
		}
	})

	t.Run("successful exchange sets tokens", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token":  "exchanged-access-token",
				"refresh_token": "exchanged-refresh-token",
				"expires_in":    3600,
				"token_type":    "Bearer",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		store := NewMemoryTokenStore()
		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		}, store)

		err := client.ExchangeCode(context.Background(), "valid-code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify tokens were set
		tokens := client.GetTokens()
		if tokens == nil {
			t.Fatal("expected tokens after successful exchange")
		}
		if tokens.AccessToken != "exchanged-access-token" {
			t.Errorf("expected access token 'exchanged-access-token', got: %s", tokens.AccessToken)
		}
		if tokens.RefreshToken != "exchanged-refresh-token" {
			t.Errorf("expected refresh token 'exchanged-refresh-token', got: %s", tokens.RefreshToken)
		}

		// Verify client is authenticated
		if !client.IsAuthenticated() {
			t.Error("client should be authenticated after exchange")
		}

		// Verify tokens were persisted to store
		storedTokens, _ := store.LoadTokens(context.Background())
		if storedTokens.AccessToken != "exchanged-access-token" {
			t.Errorf("tokens should be persisted to store")
		}
	})

	t.Run("exchange failure returns wrapped error", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error":             "invalid_grant",
				"error_description": "Invalid authorization code",
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		}, NewMemoryTokenStore())

		err := client.ExchangeCode(context.Background(), "invalid-code")
		if err == nil {
			t.Fatal("expected error for invalid code")
		}
		if !containsHelper(err.Error(), "ExchangeCode") {
			t.Errorf("error should contain 'ExchangeCode', got: %v", err)
		}
	})
}

func TestTokenRefreshTransport_RoundTrip(t *testing.T) {
	t.Run("adds authorization header when token exists", func(t *testing.T) {
		var capturedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`{"items":[]}`))
		}))
		defer server.Close()

		store := NewMemoryTokenStore()
		store.SaveTokens(context.Background(), &TokenResponse{
			AccessToken:  "test-token-123",
			RefreshToken: "refresh",
			ExpiresAt:    time.Now().Add(time.Hour),
		})

		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, store, WithBaseURL(server.URL))

		// Make a request
		_, err := client.ListDevices(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if capturedAuth != "Bearer test-token-123" {
			t.Errorf("expected Authorization header 'Bearer test-token-123', got %q", capturedAuth)
		}
	})

	t.Run("returns error when no tokens available", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewOAuthClient(&OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
		}, NewMemoryTokenStore(), WithBaseURL(server.URL))

		// Make a request without tokens - should fail
		_, err := client.ListDevices(context.Background())
		if err == nil {
			t.Error("expected error when no tokens available")
		}
	})
}

func TestDoTokenRequestWithAuth(t *testing.T) {
	t.Run("verifies basic auth header is set", func(t *testing.T) {
		var capturedAuth string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			capturedAuth = r.Header.Get("Authorization")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "my-client-id",
			ClientSecret: "my-client-secret",
			RedirectURL:  "http://localhost/callback",
		}

		_, err := ExchangeCode(context.Background(), cfg, "code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Basic auth should be "my-client-id:my-client-secret" base64 encoded
		if capturedAuth == "" {
			t.Error("Authorization header should be set")
		}
		if !containsHelper(capturedAuth, "Basic") {
			t.Errorf("expected Basic auth, got: %s", capturedAuth)
		}
	})

	t.Run("sets correct content type and accept headers", func(t *testing.T) {
		var contentType, accept string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			contentType = r.Header.Get("Content-Type")
			accept = r.Header.Get("Accept")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		}

		_, err := ExchangeCode(context.Background(), cfg, "code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if contentType != "application/x-www-form-urlencoded" {
			t.Errorf("expected Content-Type 'application/x-www-form-urlencoded', got: %s", contentType)
		}
		if accept != "application/json" {
			t.Errorf("expected Accept 'application/json', got: %s", accept)
		}
	})

	t.Run("includes client credentials in body", func(t *testing.T) {
		var clientID, clientSecret string
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.ParseForm()
			clientID = r.FormValue("client_id")
			clientSecret = r.FormValue("client_secret")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "body-client-id",
			ClientSecret: "body-client-secret",
			RedirectURL:  "http://localhost/callback",
		}

		_, err := ExchangeCode(context.Background(), cfg, "code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if clientID != "body-client-id" {
			t.Errorf("expected client_id 'body-client-id', got: %s", clientID)
		}
		if clientSecret != "body-client-secret" {
			t.Errorf("expected client_secret 'body-client-secret', got: %s", clientSecret)
		}
	})

	t.Run("sets ExpiresAt from expires_in", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token",
				"expires_in":   3600,
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		}

		tokens, err := ExchangeCode(context.Background(), cfg, "code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if tokens.ExpiresAt.IsZero() {
			t.Error("ExpiresAt should be set")
		}

		// Should be approximately 1 hour from now
		expected := time.Now().Add(3600 * time.Second)
		diff := tokens.ExpiresAt.Sub(expected)
		if diff < -time.Second || diff > time.Second {
			t.Errorf("ExpiresAt should be ~1 hour from now, got diff: %v", diff)
		}
	})

	t.Run("preserves ExpiresAt if already set in response", func(t *testing.T) {
		expiresAt := time.Now().Add(2 * time.Hour)
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"access_token": "token",
				"expires_in":   3600,
				"expires_at":   expiresAt.Format(time.RFC3339),
			})
		}))
		defer server.Close()

		originalEndpoint := tokenEndpoint
		tokenEndpoint = server.URL
		defer func() { tokenEndpoint = originalEndpoint }()

		cfg := &OAuthConfig{
			ClientID:     "id",
			ClientSecret: "secret",
			RedirectURL:  "http://localhost/callback",
		}

		tokens, err := ExchangeCode(context.Background(), cfg, "code")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// When ExpiresAt is provided in JSON, it should be used as-is
		if tokens.ExpiresAt.IsZero() {
			t.Error("ExpiresAt should be set")
		}
	})
}

func TestDefaultScopes(t *testing.T) {
	scopes := DefaultScopes()

	expected := []string{
		"r:devices:*",
		"x:devices:*",
		"r:locations:*",
	}

	if len(scopes) != len(expected) {
		t.Errorf("expected %d scopes, got %d", len(expected), len(scopes))
	}

	for i, scope := range expected {
		if scopes[i] != scope {
			t.Errorf("scope[%d] = %q, want %q", i, scopes[i], scope)
		}
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
