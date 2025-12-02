package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// AppType represents the type of SmartApp.
type AppType string

// App type constants.
const (
	AppTypeWebhook AppType = "WEBHOOK_SMART_APP"
	AppTypeLambda  AppType = "LAMBDA_SMART_APP"
)

// App represents a SmartThings SmartApp.
type App struct {
	AppID           string          `json:"appId"`
	AppName         string          `json:"appName"`
	AppType         AppType         `json:"appType"`
	Classifications []string        `json:"classifications,omitempty"`
	DisplayName     string          `json:"displayName,omitempty"`
	Description     string          `json:"description,omitempty"`
	IconImage       *IconImage      `json:"iconImage,omitempty"`
	SingleInstance  bool            `json:"singleInstance,omitempty"`
	WebhookSmartApp *WebhookAppInfo `json:"webhookSmartApp,omitempty"`
	LambdaSmartApp  *LambdaAppInfo  `json:"lambdaSmartApp,omitempty"`
	CreatedDate     string          `json:"createdDate,omitempty"`
	LastUpdatedDate string          `json:"lastUpdatedDate,omitempty"`
}

// IconImage represents an app icon.
type IconImage struct {
	URL string `json:"url"`
}

// WebhookAppInfo contains webhook SmartApp configuration.
type WebhookAppInfo struct {
	TargetURL        string   `json:"targetUrl"`
	TargetStatus     string   `json:"targetStatus,omitempty"` // PENDING, CONFIRMED
	PublicKey        string   `json:"publicKey,omitempty"`
	SignatureType    string   `json:"signatureType,omitempty"` // ST_PADLOCK, APP_RSA
	ConfirmationURL  string   `json:"confirmationUrl,omitempty"`
	MaxInstances     int      `json:"maxInstances,omitempty"`
	PermissionScopes []string `json:"permissionScopes,omitempty"`
}

// LambdaAppInfo contains Lambda SmartApp configuration.
type LambdaAppInfo struct {
	Functions []string `json:"functions"`
}

// AppCreate is the request body for creating a new SmartApp.
type AppCreate struct {
	AppName         string          `json:"appName"`
	AppType         AppType         `json:"appType"`
	Classifications []string        `json:"classifications,omitempty"`
	DisplayName     string          `json:"displayName,omitempty"`
	Description     string          `json:"description,omitempty"`
	SingleInstance  bool            `json:"singleInstance,omitempty"`
	IconImage       *IconImage      `json:"iconImage,omitempty"`
	WebhookSmartApp *WebhookAppInfo `json:"webhookSmartApp,omitempty"`
	LambdaSmartApp  *LambdaAppInfo  `json:"lambdaSmartApp,omitempty"`
}

// AppUpdate is the request body for updating a SmartApp.
type AppUpdate struct {
	DisplayName     string          `json:"displayName,omitempty"`
	Description     string          `json:"description,omitempty"`
	SingleInstance  bool            `json:"singleInstance,omitempty"`
	IconImage       *IconImage      `json:"iconImage,omitempty"`
	Classifications []string        `json:"classifications,omitempty"`
	WebhookSmartApp *WebhookAppInfo `json:"webhookSmartApp,omitempty"`
	LambdaSmartApp  *LambdaAppInfo  `json:"lambdaSmartApp,omitempty"`
}

// AppOAuth represents OAuth settings for a SmartApp.
type AppOAuth struct {
	ClientName   string   `json:"clientName"`
	Scope        []string `json:"scope"`
	RedirectUris []string `json:"redirectUris,omitempty"`
}

// AppOAuthGenerated contains newly generated OAuth credentials.
type AppOAuthGenerated struct {
	ClientID     string `json:"oauthClientId"`
	ClientSecret string `json:"oauthClientSecret"`
}

// appListResponse is the API response for listing apps.
type appListResponse struct {
	Items []App `json:"items"`
}

// ListApps returns all SmartApps for the authenticated account.
func (c *Client) ListApps(ctx context.Context) ([]App, error) {
	data, err := c.get(ctx, "/apps")
	if err != nil {
		return nil, err
	}

	var resp appListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse app list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetApp returns a single SmartApp by ID.
func (c *Client) GetApp(ctx context.Context, appID string) (*App, error) {
	if appID == "" {
		return nil, ErrEmptyAppID
	}

	data, err := c.get(ctx, "/apps/"+appID)
	if err != nil {
		return nil, err
	}

	var app App
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("failed to parse app: %w (body: %s)", err, truncatePreview(data))
	}

	return &app, nil
}

// CreateApp registers a new SmartApp.
func (c *Client) CreateApp(ctx context.Context, app *AppCreate) (*App, error) {
	if app == nil || app.AppName == "" {
		return nil, ErrEmptyAppName
	}

	data, err := c.post(ctx, "/apps", app)
	if err != nil {
		return nil, err
	}

	var created App
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created app: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// UpdateApp updates an existing SmartApp.
func (c *Client) UpdateApp(ctx context.Context, appID string, update *AppUpdate) (*App, error) {
	if appID == "" {
		return nil, ErrEmptyAppID
	}

	data, err := c.put(ctx, "/apps/"+appID, update)
	if err != nil {
		return nil, err
	}

	var updated App
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated app: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// DeleteApp removes a SmartApp.
func (c *Client) DeleteApp(ctx context.Context, appID string) error {
	if appID == "" {
		return ErrEmptyAppID
	}

	_, err := c.delete(ctx, "/apps/"+appID)
	return err
}

// GetAppOAuth returns the OAuth settings for a SmartApp.
func (c *Client) GetAppOAuth(ctx context.Context, appID string) (*AppOAuth, error) {
	if appID == "" {
		return nil, ErrEmptyAppID
	}

	data, err := c.get(ctx, "/apps/"+appID+"/oauth")
	if err != nil {
		return nil, err
	}

	var oauth AppOAuth
	if err := json.Unmarshal(data, &oauth); err != nil {
		return nil, fmt.Errorf("failed to parse app OAuth: %w (body: %s)", err, truncatePreview(data))
	}

	return &oauth, nil
}

// UpdateAppOAuth updates the OAuth settings for a SmartApp.
func (c *Client) UpdateAppOAuth(ctx context.Context, appID string, oauth *AppOAuth) (*AppOAuth, error) {
	if appID == "" {
		return nil, ErrEmptyAppID
	}

	data, err := c.put(ctx, "/apps/"+appID+"/oauth", oauth)
	if err != nil {
		return nil, err
	}

	var updated AppOAuth
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated OAuth: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// GenerateAppOAuth generates new OAuth client credentials for a SmartApp.
// This will invalidate any existing credentials.
func (c *Client) GenerateAppOAuth(ctx context.Context, appID string) (*AppOAuthGenerated, error) {
	if appID == "" {
		return nil, ErrEmptyAppID
	}

	data, err := c.post(ctx, "/apps/"+appID+"/oauth/generate", nil)
	if err != nil {
		return nil, err
	}

	var generated AppOAuthGenerated
	if err := json.Unmarshal(data, &generated); err != nil {
		return nil, fmt.Errorf("failed to parse generated OAuth: %w (body: %s)", err, truncatePreview(data))
	}

	return &generated, nil
}
