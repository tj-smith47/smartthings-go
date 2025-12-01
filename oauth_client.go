package smartthings

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// OAuthClient wraps a Client and manages OAuth token lifecycle
type OAuthClient struct {
	*Client
	config     *OAuthConfig
	tokenStore TokenStore
	tokens     *TokenResponse
	mu         sync.RWMutex
}

// NewOAuthClient creates a new OAuth-enabled SmartThings client.
// It attempts to load existing tokens from the store.
// If no tokens are available, the client will be created but API calls will fail
// until tokens are set via SetTokens or obtained through the OAuth flow.
func NewOAuthClient(cfg *OAuthConfig, store TokenStore, opts ...Option) (*OAuthClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("OAuth config is required")
	}
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("client ID is required")
	}
	if cfg.ClientSecret == "" {
		return nil, fmt.Errorf("client secret is required")
	}
	if store == nil {
		return nil, fmt.Errorf("token store is required")
	}

	// Create base client with a placeholder token
	// The token will be set from the store or OAuth flow
	client, err := NewClient("placeholder", opts...)
	if err != nil && err != ErrEmptyToken {
		return nil, fmt.Errorf("failed to create base client: %w", err)
	}

	// Override the placeholder - we need to create manually since NewClient requires token
	if client == nil {
		client = &Client{
			baseURL: DefaultBaseURL,
			token:   "",
			httpClient: defaultHTTPClient(),
		}
		for _, opt := range opts {
			opt(client)
		}
	}

	oc := &OAuthClient{
		Client:     client,
		config:     cfg,
		tokenStore: store,
	}

	// Try to load existing tokens
	ctx := context.Background()
	if tokens, err := store.LoadTokens(ctx); err == nil {
		oc.tokens = tokens
		oc.Client.SetToken(tokens.AccessToken)
	}

	return oc, nil
}

// defaultHTTPClient returns the default HTTP client configuration
func defaultHTTPClient() *http.Client {
	return &http.Client{
		Timeout: DefaultTimeout,
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}
}

// SetTokens sets the OAuth tokens and updates the underlying client
func (c *OAuthClient) SetTokens(ctx context.Context, tokens *TokenResponse) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens = tokens
	c.Client.SetToken(tokens.AccessToken)

	// Persist the tokens
	if err := c.tokenStore.SaveTokens(ctx, tokens); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	return nil
}

// GetTokens returns the current tokens (read-only copy)
func (c *OAuthClient) GetTokens() *TokenResponse {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.tokens == nil {
		return nil
	}

	// Return a copy to prevent external modification
	tokensCopy := *c.tokens
	return &tokensCopy
}

// IsAuthenticated returns true if valid tokens are available
func (c *OAuthClient) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.tokens != nil && c.tokens.IsValid()
}

// NeedsReauthentication returns true if the refresh token is invalid/expired
// and the user needs to go through the OAuth flow again
func (c *OAuthClient) NeedsReauthentication() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.tokens == nil {
		return true
	}
	return !c.tokens.IsRefreshTokenValid()
}

// EnsureValidToken checks if the access token is valid and refreshes if needed.
// This should be called before making API requests.
func (c *OAuthClient) EnsureValidToken(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tokens == nil {
		return fmt.Errorf("no tokens available - OAuth authentication required")
	}

	// Token is still valid
	if c.tokens.IsValid() {
		return nil
	}

	// Check if refresh token is still valid
	if !c.tokens.IsRefreshTokenValid() {
		return fmt.Errorf("refresh token expired - OAuth re-authentication required")
	}

	// Refresh the token
	newTokens, err := RefreshTokens(ctx, c.config, c.tokens.RefreshToken)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %w", err)
	}

	// Update tokens
	c.tokens = newTokens
	c.Client.SetToken(newTokens.AccessToken)

	// Persist the new tokens
	if err := c.tokenStore.SaveTokens(ctx, newTokens); err != nil {
		// Log but don't fail - we have valid tokens in memory
		// The next call will try to persist again
		return nil
	}

	return nil
}

// GetAuthorizationURL returns the URL to start the OAuth flow
func (c *OAuthClient) GetAuthorizationURL(state string) string {
	return GetAuthorizationURL(c.config, state)
}

// ExchangeCode exchanges an authorization code for tokens
func (c *OAuthClient) ExchangeCode(ctx context.Context, code string) error {
	tokens, err := ExchangeCode(ctx, c.config, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	return c.SetTokens(ctx, tokens)
}

// ClearTokens removes all stored tokens
func (c *OAuthClient) ClearTokens(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens = nil
	c.Client.SetToken("")

	// Try to delete from store if it supports deletion
	if deleter, ok := c.tokenStore.(interface{ Delete(context.Context) error }); ok {
		return deleter.Delete(ctx)
	}

	// Otherwise just save nil/empty
	return c.tokenStore.SaveTokens(ctx, &TokenResponse{})
}

// --- Override Client methods to ensure valid token before requests ---

// ListDevices returns all devices, ensuring valid token first
func (c *OAuthClient) ListDevices(ctx context.Context) ([]Device, error) {
	if err := c.EnsureValidToken(ctx); err != nil {
		return nil, err
	}
	return c.Client.ListDevices(ctx)
}

// GetDevice returns a device by ID, ensuring valid token first
func (c *OAuthClient) GetDevice(ctx context.Context, deviceID string) (*Device, error) {
	if err := c.EnsureValidToken(ctx); err != nil {
		return nil, err
	}
	return c.Client.GetDevice(ctx, deviceID)
}

// GetDeviceStatus returns the status of a device, ensuring valid token first
func (c *OAuthClient) GetDeviceStatus(ctx context.Context, deviceID string) (Status, error) {
	if err := c.EnsureValidToken(ctx); err != nil {
		return nil, err
	}
	return c.Client.GetDeviceStatus(ctx, deviceID)
}

// GetDeviceStatusAllComponents returns status for all components, ensuring valid token first
func (c *OAuthClient) GetDeviceStatusAllComponents(ctx context.Context, deviceID string) (Status, error) {
	if err := c.EnsureValidToken(ctx); err != nil {
		return nil, err
	}
	return c.Client.GetDeviceStatusAllComponents(ctx, deviceID)
}

// ExecuteCommand executes a command on a device, ensuring valid token first
func (c *OAuthClient) ExecuteCommand(ctx context.Context, deviceID string, cmd Command) error {
	if err := c.EnsureValidToken(ctx); err != nil {
		return err
	}
	return c.Client.ExecuteCommand(ctx, deviceID, cmd)
}

// ExecuteCommands executes multiple commands on a device, ensuring valid token first
func (c *OAuthClient) ExecuteCommands(ctx context.Context, deviceID string, cmds []Command) error {
	if err := c.EnsureValidToken(ctx); err != nil {
		return err
	}
	return c.Client.ExecuteCommands(ctx, deviceID, cmds)
}
