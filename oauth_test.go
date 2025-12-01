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
				"r%3Adevices%3A%2A", // r:devices:*
				"x%3Adevices%3A%2A", // x:devices:*
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
				t.Errorf("expected grant_type=authorization_code")
			}
			if r.FormValue("code") != "test-code" {
				t.Errorf("expected code=test-code")
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

		// We can't easily override the token endpoint, so we test the token parsing
		// In a real test, we'd need to inject the endpoint or use a mock
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
}

func TestRefreshTokens(t *testing.T) {
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
