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

// ConfigValueType represents the type of configuration value.
type ConfigValueType string

// Configuration value type constants.
const (
	ConfigValueTypeString     ConfigValueType = "STRING"
	ConfigValueTypeDevice     ConfigValueType = "DEVICE"
	ConfigValueTypePermission ConfigValueType = "PERMISSION"
	ConfigValueTypeMode       ConfigValueType = "MODE"
	ConfigValueTypeScene      ConfigValueType = "SCENE"
	ConfigValueTypeMessage    ConfigValueType = "MESSAGE"
)

// StringConfig represents a string configuration value.
type StringConfig struct {
	Value string `json:"value"`
}

// DeviceConfig represents a device configuration value.
type DeviceConfig struct {
	DeviceID    string   `json:"deviceId"`
	ComponentID string   `json:"componentId,omitempty"`
	Permissions []string `json:"permissions,omitempty"`
}

// PermissionConfig represents a permission configuration value.
type PermissionConfig struct {
	Permissions []string `json:"permissions"`
}

// ModeConfig represents a mode configuration value.
type ModeConfig struct {
	ModeID string `json:"modeId"`
}

// SceneConfig represents a scene configuration value.
type SceneConfig struct {
	SceneID     string   `json:"sceneId"`
	Permissions []string `json:"permissions,omitempty"`
}

// MessageConfig represents a message group configuration value.
type MessageConfig struct {
	MessageGroupKey string `json:"messageGroupKey"`
}

// ConfigEntry represents a single configuration entry in an installed app.
type ConfigEntry struct {
	ValueType        ConfigValueType   `json:"valueType"`
	StringConfig     *StringConfig     `json:"stringConfig,omitempty"`
	DeviceConfig     *DeviceConfig     `json:"deviceConfig,omitempty"`
	PermissionConfig *PermissionConfig `json:"permissionConfig,omitempty"`
	ModeConfig       *ModeConfig       `json:"modeConfig,omitempty"`
	SceneConfig      *SceneConfig      `json:"sceneConfig,omitempty"`
	MessageConfig    *MessageConfig    `json:"messageConfig,omitempty"`
}

// InstalledAppConfigItem represents a configuration item in the list response.
type InstalledAppConfigItem struct {
	ConfigurationID     string `json:"configurationId"`
	ConfigurationStatus string `json:"configurationStatus"` // STAGED, AUTHORIZED, REVOKED
	CreatedDate         string `json:"createdDate,omitempty"`
	LastUpdatedDate     string `json:"lastUpdatedDate,omitempty"`
}

// InstalledAppConfiguration represents the full configuration of an installed app.
type InstalledAppConfiguration struct {
	InstalledAppID      string                   `json:"installedAppId"`
	ConfigurationID     string                   `json:"configurationId"`
	ConfigurationStatus string                   `json:"configurationStatus"` // STAGED, AUTHORIZED, REVOKED
	CreatedDate         string                   `json:"createdDate,omitempty"`
	LastUpdatedDate     string                   `json:"lastUpdatedDate,omitempty"`
	Config              map[string][]ConfigEntry `json:"config"`
}

// installedAppConfigListResponse is the API response for listing configs.
type installedAppConfigListResponse struct {
	Items []InstalledAppConfigItem `json:"items"`
}

// ListInstalledAppConfigs returns all configurations for an installed app.
func (c *Client) ListInstalledAppConfigs(ctx context.Context, installedAppID string) ([]InstalledAppConfigItem, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}

	data, err := c.get(ctx, "/installedapps/"+installedAppID+"/configs")
	if err != nil {
		return nil, err
	}

	var resp installedAppConfigListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse installed app config list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetInstalledAppConfig returns a specific configuration by ID.
func (c *Client) GetInstalledAppConfig(ctx context.Context, installedAppID, configID string) (*InstalledAppConfiguration, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}
	if configID == "" {
		return nil, fmt.Errorf("smartthings: configuration ID cannot be empty")
	}

	data, err := c.get(ctx, "/installedapps/"+installedAppID+"/configs/"+configID)
	if err != nil {
		return nil, err
	}

	var config InstalledAppConfiguration
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse installed app config: %w (body: %s)", err, truncatePreview(data))
	}

	return &config, nil
}

// GetCurrentInstalledAppConfig returns the current (latest authorized) configuration.
// Returns nil if no authorized configuration exists.
func (c *Client) GetCurrentInstalledAppConfig(ctx context.Context, installedAppID string) (*InstalledAppConfiguration, error) {
	configs, err := c.ListInstalledAppConfigs(ctx, installedAppID)
	if err != nil {
		return nil, err
	}

	// Find the latest authorized configuration
	for _, cfg := range configs {
		if cfg.ConfigurationStatus == "AUTHORIZED" {
			return c.GetInstalledAppConfig(ctx, installedAppID, cfg.ConfigurationID)
		}
	}

	return nil, nil
}
