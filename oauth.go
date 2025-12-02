package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	// OAuth endpoints
	authorizationEndpoint = "https://api.smartthings.com/oauth/authorize"
	tokenEndpoint         = "https://api.smartthings.com/oauth/token"

	// Default scopes for SmartThings OAuth
	defaultScopeDevicesRead     = "r:devices:*"
	defaultScopeDevicesExecute  = "x:devices:*"
	defaultScopeLocationsRead   = "r:locations:*"

	// tokenRefreshBuffer is how long before expiry we should refresh the token
	tokenRefreshBuffer = 5 * time.Minute
)

// DefaultScopes returns the default OAuth scopes for SmartThings
func DefaultScopes() []string {
	return []string{
		defaultScopeDevicesRead,
		defaultScopeDevicesExecute,
		defaultScopeLocationsRead,
	}
}

// OAuthConfig holds the configuration for OAuth authentication
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

// TokenResponse represents the response from the OAuth token endpoint
type TokenResponse struct {
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	ExpiresIn             int       `json:"expires_in"`
	RefreshTokenExpiresIn int       `json:"refresh_token_expires_in,omitempty"`
	TokenType             string    `json:"token_type"`
	Scope                 string    `json:"scope"`
	InstalledAppID        string    `json:"installed_app_id,omitempty"`
	ExpiresAt             time.Time `json:"expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at,omitempty"`
}

// IsValid checks if the access token is still valid (with buffer)
func (t *TokenResponse) IsValid() bool {
	if t == nil || t.AccessToken == "" {
		return false
	}
	return time.Now().Add(tokenRefreshBuffer).Before(t.ExpiresAt)
}

// IsRefreshTokenValid checks if the refresh token is still valid
func (t *TokenResponse) IsRefreshTokenValid() bool {
	if t == nil || t.RefreshToken == "" {
		return false
	}
	// If RefreshTokenExpiresAt is zero, assume refresh token doesn't expire
	if t.RefreshTokenExpiresAt.IsZero() {
		return true
	}
	return time.Now().Before(t.RefreshTokenExpiresAt)
}

// NeedsRefresh returns true if the access token should be refreshed
func (t *TokenResponse) NeedsRefresh() bool {
	if t == nil || t.AccessToken == "" {
		return true
	}
	return time.Now().Add(tokenRefreshBuffer).After(t.ExpiresAt)
}

// TokenStore is the interface for persisting OAuth tokens
type TokenStore interface {
	SaveTokens(ctx context.Context, tokens *TokenResponse) error
	LoadTokens(ctx context.Context) (*TokenResponse, error)
}

// GetAuthorizationURL returns the URL to redirect users to for OAuth authorization
func GetAuthorizationURL(cfg *OAuthConfig, state string) string {
	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = DefaultScopes()
	}

	params := url.Values{}
	params.Set("response_type", "code")
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURL)
	params.Set("scope", strings.Join(scopes, " "))
	if state != "" {
		params.Set("state", state)
	}

	return authorizationEndpoint + "?" + params.Encode()
}

// ExchangeCode exchanges an authorization code for access and refresh tokens
func ExchangeCode(ctx context.Context, cfg *OAuthConfig, code string) (*TokenResponse, error) {
	if code == "" {
		return nil, fmt.Errorf("authorization code is required")
	}

	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", cfg.RedirectURL)
	data.Set("code", code)

	return doTokenRequestWithAuth(ctx, cfg.ClientID, cfg.ClientSecret, data)
}

// RefreshTokens refreshes the access token using a refresh token
func RefreshTokens(ctx context.Context, cfg *OAuthConfig, refreshToken string) (*TokenResponse, error) {
	if refreshToken == "" {
		return nil, fmt.Errorf("refresh token is required")
	}

	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)

	return doTokenRequestWithAuth(ctx, cfg.ClientID, cfg.ClientSecret, data)
}

// doTokenRequestWithAuth performs a token request using HTTP Basic Auth
func doTokenRequestWithAuth(ctx context.Context, clientID, clientSecret string, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth(clientID, clientSecret)

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("OAuth error: %s - %s", errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Set expiry time if not provided
	if tokens.ExpiresAt.IsZero() && tokens.ExpiresIn > 0 {
		tokens.ExpiresAt = time.Now().Add(time.Duration(tokens.ExpiresIn) * time.Second)
	}

	return &tokens, nil
}

// doTokenRequest performs a token request to the OAuth token endpoint (deprecated, use doTokenRequestWithAuth)
func doTokenRequest(ctx context.Context, data url.Values) (*TokenResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, tokenEndpoint, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp struct {
			Error            string `json:"error"`
			ErrorDescription string `json:"error_description"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error != "" {
			return nil, fmt.Errorf("OAuth error: %s - %s", errResp.Error, errResp.ErrorDescription)
		}
		return nil, fmt.Errorf("token request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokens TokenResponse
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	// Calculate expiry times
	now := time.Now()
	if tokens.ExpiresIn > 0 {
		tokens.ExpiresAt = now.Add(time.Duration(tokens.ExpiresIn) * time.Second)
	}
	if tokens.RefreshTokenExpiresIn > 0 {
		tokens.RefreshTokenExpiresAt = now.Add(time.Duration(tokens.RefreshTokenExpiresIn) * time.Second)
	}

	return &tokens, nil
}
