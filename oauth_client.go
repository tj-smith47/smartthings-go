package smartthings

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"
)

// OAuthClient wraps a Client and manages OAuth token lifecycle.
// It automatically refreshes tokens before they expire.
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

	oc := &OAuthClient{
		config:     cfg,
		tokenStore: store,
	}

	// Try to load existing tokens
	ctx := context.Background()
	if tokens, err := store.LoadTokens(ctx); err == nil {
		oc.tokens = tokens
	}

	// Create a custom transport that refreshes tokens automatically
	transport := &tokenRefreshTransport{
		oauth: oc,
		base: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
		},
	}

	httpClient := &http.Client{
		Timeout:   DefaultTimeout,
		Transport: transport,
	}

	// Create the base client with empty token (will be set by transport)
	client := &Client{
		baseURL:    DefaultBaseURL,
		token:      "",
		httpClient: httpClient,
	}

	// Apply options
	for _, opt := range opts {
		opt(client)
	}

	// If options replaced the http client, wrap its transport
	if client.httpClient != httpClient {
		transport.base = client.httpClient.Transport
		client.httpClient.Transport = transport
	}

	// Set initial token if available
	if oc.tokens != nil {
		client.token = oc.tokens.AccessToken
	}

	oc.Client = client
	return oc, nil
}

// tokenRefreshTransport is an http.RoundTripper that ensures the OAuth token
// is valid before each request. This eliminates the need to override every
// API method in OAuthClient.
type tokenRefreshTransport struct {
	oauth *OAuthClient
	base  http.RoundTripper
}

// RoundTrip implements http.RoundTripper.
func (t *tokenRefreshTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Ensure we have a valid token before making the request
	if err := t.oauth.ensureValidTokenInternal(req.Context()); err != nil {
		return nil, err
	}

	// Update the Authorization header with the current token
	t.oauth.mu.RLock()
	token := ""
	if t.oauth.tokens != nil {
		token = t.oauth.tokens.AccessToken
	}
	t.oauth.mu.RUnlock()

	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	return t.base.RoundTrip(req)
}

// ensureValidTokenInternal checks if the access token is valid and refreshes if needed.
// This is the internal version that doesn't acquire a read lock first.
func (c *OAuthClient) ensureValidTokenInternal(ctx context.Context) error {
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
	c.Client.token = newTokens.AccessToken

	// Persist the new tokens (ignore errors - we have valid tokens in memory)
	_ = c.tokenStore.SaveTokens(ctx, newTokens)

	return nil
}

// EnsureValidToken checks if the access token is valid and refreshes if needed.
// This is the public version for manual token validation.
func (c *OAuthClient) EnsureValidToken(ctx context.Context) error {
	return c.ensureValidTokenInternal(ctx)
}

// SetTokens sets the OAuth tokens and updates the underlying client.
func (c *OAuthClient) SetTokens(ctx context.Context, tokens *TokenResponse) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens = tokens
	c.Client.token = tokens.AccessToken

	// Persist the tokens
	if err := c.tokenStore.SaveTokens(ctx, tokens); err != nil {
		return fmt.Errorf("failed to save tokens: %w", err)
	}

	return nil
}

// GetTokens returns the current tokens (read-only copy).
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

// IsAuthenticated returns true if valid tokens are available.
func (c *OAuthClient) IsAuthenticated() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return c.tokens != nil && c.tokens.IsValid()
}

// NeedsReauthentication returns true if the refresh token is invalid/expired
// and the user needs to go through the OAuth flow again.
func (c *OAuthClient) NeedsReauthentication() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if c.tokens == nil {
		return true
	}
	return !c.tokens.IsRefreshTokenValid()
}

// GetAuthorizationURL returns the URL to start the OAuth flow.
func (c *OAuthClient) GetAuthorizationURL(state string) string {
	return GetAuthorizationURL(c.config, state)
}

// ExchangeCode exchanges an authorization code for tokens.
func (c *OAuthClient) ExchangeCode(ctx context.Context, code string) error {
	tokens, err := ExchangeCode(ctx, c.config, code)
	if err != nil {
		return fmt.Errorf("failed to exchange code: %w", err)
	}

	return c.SetTokens(ctx, tokens)
}

// ClearTokens removes all stored tokens.
func (c *OAuthClient) ClearTokens(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.tokens = nil
	c.Client.token = ""

	// Try to delete from store if it supports deletion
	if deleter, ok := c.tokenStore.(interface{ Delete(context.Context) error }); ok {
		return deleter.Delete(ctx)
	}

	// Otherwise just save empty tokens
	return c.tokenStore.SaveTokens(ctx, &TokenResponse{})
}

// Config returns the OAuth configuration.
func (c *OAuthClient) Config() *OAuthConfig {
	return c.config
}

// TokenStore returns the token store.
func (c *OAuthClient) TokenStore() TokenStore {
	return c.tokenStore
}
