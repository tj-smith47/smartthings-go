package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// SchemaApp represents a SmartThings Schema (C2C) connector.
type SchemaApp struct {
	EndpointAppID         string         `json:"endpointAppId"`
	AppName               string         `json:"appName,omitempty"`
	PartnerName           string         `json:"partnerName,omitempty"`
	OAuthClientID         string         `json:"oauthClientId,omitempty"`
	OAuthClientSecret     string         `json:"oauthClientSecret,omitempty"`
	STClientID            string         `json:"stClientId,omitempty"`
	STClientSecret        string         `json:"stClientSecret,omitempty"`
	OAuthAuthorizationURL string         `json:"oauthAuthorizationUrl,omitempty"`
	OAuthTokenURL         string         `json:"oauthTokenUrl,omitempty"`
	WebhookURL            string         `json:"webhookUrl,omitempty"`
	HostingType           string         `json:"hostingType,omitempty"`
	UserEmail             string         `json:"userEmail,omitempty"`
	IconURL               string         `json:"icon,omitempty"`
	IconX2URL             string         `json:"icon2x,omitempty"`
	IconX3URL             string         `json:"icon3x,omitempty"`
	CertificationStatus   string         `json:"certificationStatus,omitempty"`
	CreatedDate           string         `json:"createdDate,omitempty"`
	LastUpdatedDate       string         `json:"lastUpdatedDate,omitempty"`
	ViperAppLinks         *ViperAppLinks `json:"viperAppLinks,omitempty"`
}

// ViperAppLinks contains links for ST Schema app configuration.
type ViperAppLinks struct {
	Android string `json:"android,omitempty"`
	IOS     string `json:"ios,omitempty"`
	Web     string `json:"web,omitempty"`
}

// SchemaAppRequest is the request body for creating/updating a Schema app.
type SchemaAppRequest struct {
	AppName               string         `json:"appName"`
	PartnerName           string         `json:"partnerName,omitempty"`
	OAuthAuthorizationURL string         `json:"oauthAuthorizationUrl,omitempty"`
	OAuthTokenURL         string         `json:"oauthTokenUrl,omitempty"`
	OAuthClientID         string         `json:"oauthClientId,omitempty"`
	OAuthClientSecret     string         `json:"oauthClientSecret,omitempty"`
	WebhookURL            string         `json:"webhookUrl,omitempty"`
	HostingType           string         `json:"hostingType,omitempty"`
	UserEmail             string         `json:"userEmail,omitempty"`
	IconURL               string         `json:"icon,omitempty"`
	IconX2URL             string         `json:"icon2x,omitempty"`
	IconX3URL             string         `json:"icon3x,omitempty"`
	ViperAppLinks         *ViperAppLinks `json:"viperAppLinks,omitempty"`
}

// SchemaCreateResponse is the response from creating a Schema app.
type SchemaCreateResponse struct {
	EndpointAppID  string `json:"endpointAppId"`
	STClientID     string `json:"stClientId,omitempty"`
	STClientSecret string `json:"stClientSecret,omitempty"`
}

// InstalledSchemaApp represents an installed instance of a Schema app.
type InstalledSchemaApp struct {
	IsaID           string               `json:"isaId"`
	AppName         string               `json:"appName,omitempty"`
	PartnerName     string               `json:"partnerName,omitempty"`
	LocationID      string               `json:"locationId,omitempty"`
	InstalledAppID  string               `json:"installedAppId,omitempty"`
	CreatedDate     string               `json:"createdDate,omitempty"`
	LastUpdatedDate string               `json:"lastUpdatedDate,omitempty"`
	Devices         []SchemaDeviceResult `json:"devices,omitempty"`
}

// SchemaDeviceResult represents a device created by a Schema app.
type SchemaDeviceResult struct {
	DeviceID         string `json:"deviceId,omitempty"`
	ExternalDeviceID string `json:"externalDeviceId,omitempty"`
	Label            string `json:"label,omitempty"`
}

// SchemaPageType indicates the type of schema app page.
type SchemaPageType string

const (
	// SchemaPageTypeAuthorized indicates the user is authorized.
	SchemaPageTypeAuthorized SchemaPageType = "AUTHORIZED"
	// SchemaPageTypeUnauthorized indicates the user needs authorization.
	SchemaPageTypeUnauthorized SchemaPageType = "UNAUTHORIZED"
)

// SchemaPage represents a page configuration for a Schema app.
type SchemaPage struct {
	PageType SchemaPageType `json:"pageType"`
	// Authorized page fields
	PartnerName   string `json:"partnerName,omitempty"`
	AppName       string `json:"appName,omitempty"`
	ConnectedDate string `json:"connectedDate,omitempty"`
	// Unauthorized page fields
	AuthorizationURI string `json:"authorizationUri,omitempty"`
}

// schemaAppListResponse is the API response for listing Schema apps.
type schemaAppListResponse struct {
	Items []SchemaApp `json:"items"`
}

// installedSchemaAppListResponse is the API response for listing installed Schema apps.
type installedSchemaAppListResponse struct {
	Items []InstalledSchemaApp `json:"items"`
}

// ListSchemaApps returns all ST Schema connectors for the user.
func (c *Client) ListSchemaApps(ctx context.Context, includeAllOrganizations bool) ([]SchemaApp, error) {
	path := "/schema/apps"
	if includeAllOrganizations {
		path += "?includeAllOrganizations=true"
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp schemaAppListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListSchemaApps: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetSchemaApp returns a single Schema app by ID.
func (c *Client) GetSchemaApp(ctx context.Context, appID string) (*SchemaApp, error) {
	if appID == "" {
		return nil, ErrEmptySchemaAppID
	}

	data, err := c.get(ctx, "/schema/apps/"+appID)
	if err != nil {
		return nil, err
	}

	var app SchemaApp
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("GetSchemaApp: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &app, nil
}

// CreateSchemaApp creates a new ST Schema connector.
func (c *Client) CreateSchemaApp(ctx context.Context, req *SchemaAppRequest, organizationID string) (*SchemaCreateResponse, error) {
	if req == nil || req.AppName == "" {
		return nil, ErrEmptySchemaAppName
	}

	path := "/schema/apps"
	if organizationID != "" {
		path += "?organizationId=" + url.QueryEscape(organizationID)
	}

	data, err := c.post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var resp SchemaCreateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreateSchemaApp: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}

// UpdateSchemaApp updates an existing ST Schema connector.
func (c *Client) UpdateSchemaApp(ctx context.Context, appID string, req *SchemaAppRequest, organizationID string) error {
	if appID == "" {
		return ErrEmptySchemaAppID
	}

	path := "/schema/apps/" + appID
	if organizationID != "" {
		path += "?organizationId=" + url.QueryEscape(organizationID)
	}

	_, err := c.put(ctx, path, req)
	return err
}

// DeleteSchemaApp deletes a ST Schema connector.
func (c *Client) DeleteSchemaApp(ctx context.Context, appID string) error {
	if appID == "" {
		return ErrEmptySchemaAppID
	}

	_, err := c.delete(ctx, "/schema/apps/"+appID)
	return err
}

// RegenerateSchemaAppOAuth regenerates OAuth client credentials for a Schema app.
func (c *Client) RegenerateSchemaAppOAuth(ctx context.Context, appID string) (*SchemaCreateResponse, error) {
	if appID == "" {
		return nil, ErrEmptySchemaAppID
	}

	data, err := c.post(ctx, "/schema/apps/"+appID+"/oauth/generate", nil)
	if err != nil {
		return nil, err
	}

	var resp SchemaCreateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("RegenerateSchemaAppOAuth: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}

// GetSchemaAppPage returns the page configuration for an installed Schema app.
func (c *Client) GetSchemaAppPage(ctx context.Context, appID, locationID string) (*SchemaPage, error) {
	if appID == "" {
		return nil, ErrEmptySchemaAppID
	}
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}

	data, err := c.get(ctx, "/schema/apps/"+appID+"/page/"+locationID)
	if err != nil {
		return nil, err
	}

	var page SchemaPage
	if err := json.Unmarshal(data, &page); err != nil {
		return nil, fmt.Errorf("GetSchemaAppPage: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &page, nil
}

// ListInstalledSchemaApps returns all installed Schema app instances.
func (c *Client) ListInstalledSchemaApps(ctx context.Context, locationID string) ([]InstalledSchemaApp, error) {
	path := "/schema/installedapps"
	if locationID != "" {
		path += "?locationId=" + url.QueryEscape(locationID)
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp installedSchemaAppListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListInstalledSchemaApps: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetInstalledSchemaApp returns a single installed Schema app instance.
func (c *Client) GetInstalledSchemaApp(ctx context.Context, isaID string) (*InstalledSchemaApp, error) {
	if isaID == "" {
		return nil, ErrEmptyInstalledSchemaAppID
	}

	data, err := c.get(ctx, "/schema/installedapps/"+isaID)
	if err != nil {
		return nil, err
	}

	var app InstalledSchemaApp
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("GetInstalledSchemaApp: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &app, nil
}

// DeleteInstalledSchemaApp deletes an installed Schema app instance and its devices.
func (c *Client) DeleteInstalledSchemaApp(ctx context.Context, isaID string) error {
	if isaID == "" {
		return ErrEmptyInstalledSchemaAppID
	}

	_, err := c.delete(ctx, "/schema/installedapps/"+isaID)
	return err
}
