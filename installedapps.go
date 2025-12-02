package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// InstalledApp represents an installed SmartApp.
type InstalledApp struct {
	InstalledAppID     string   `json:"installedAppId"`
	InstalledAppType   string   `json:"installedAppType"`   // WEBHOOK_SMART_APP, LAMBDA_SMART_APP
	InstalledAppStatus string   `json:"installedAppStatus"` // PENDING, AUTHORIZED, REVOKED, DISABLED
	DisplayName        string   `json:"displayName"`
	AppID              string   `json:"appId"`
	ReferenceID        string   `json:"referenceId,omitempty"`
	LocationID         string   `json:"locationId"`
	Permissions        []string `json:"permissions,omitempty"`
	CreatedDate        string   `json:"createdDate,omitempty"`
	LastUpdatedDate    string   `json:"lastUpdatedDate,omitempty"`
	Classifications    []string `json:"classifications,omitempty"`
}

// installedAppListResponse is the API response for listing installed apps.
type installedAppListResponse struct {
	Items []InstalledApp `json:"items"`
}

// ListInstalledApps returns all installed apps for a location.
// If locationID is empty, returns installed apps for all locations.
func (c *Client) ListInstalledApps(ctx context.Context, locationID string) ([]InstalledApp, error) {
	path := "/installedapps"
	if locationID != "" {
		path += "?locationId=" + locationID
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp installedAppListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse installed app list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetInstalledApp returns a single installed app by ID.
func (c *Client) GetInstalledApp(ctx context.Context, installedAppID string) (*InstalledApp, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}

	data, err := c.get(ctx, "/installedapps/"+installedAppID)
	if err != nil {
		return nil, err
	}

	var app InstalledApp
	if err := json.Unmarshal(data, &app); err != nil {
		return nil, fmt.Errorf("failed to parse installed app: %w (body: %s)", err, truncatePreview(data))
	}

	return &app, nil
}

// DeleteInstalledApp uninstalls an app.
func (c *Client) DeleteInstalledApp(ctx context.Context, installedAppID string) error {
	if installedAppID == "" {
		return ErrEmptyInstalledAppID
	}

	_, err := c.delete(ctx, "/installedapps/"+installedAppID)
	return err
}
