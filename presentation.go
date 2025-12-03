package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// PatchOp represents a JSON Patch operation type.
type PatchOp string

const (
	// PatchOpAdd adds a value at a path.
	PatchOpAdd PatchOp = "ADD"
	// PatchOpReplace replaces a value at a path.
	PatchOpReplace PatchOp = "REPLACE"
	// PatchOpRemove removes a value at a path.
	PatchOpRemove PatchOp = "REMOVE"
)

// PatchItem represents a JSON Patch operation.
type PatchItem struct {
	Op    PatchOp `json:"op"`
	Path  string  `json:"path"`
	Value any     `json:"value,omitempty"`
}

// PresentationConfigEntry defines how a capability is displayed.
type PresentationConfigEntry struct {
	Component        string                      `json:"component"`
	Capability       string                      `json:"capability"`
	Version          int                         `json:"version,omitempty"`
	Values           []PresentationConfigValue   `json:"values,omitempty"`
	Patch            []PatchItem                 `json:"patch,omitempty"`
	VisibleCondition *CapabilityVisibleCondition `json:"visibleCondition,omitempty"`
}

// PresentationConfigValue defines a specific value configuration.
type PresentationConfigValue struct {
	Key           string                  `json:"key"`
	EnabledValues []string                `json:"enabledValues,omitempty"`
	Range         *PresentationValueRange `json:"range,omitempty"`
	Step          float64                 `json:"step,omitempty"`
}

// PresentationValueRange defines a value range.
type PresentationValueRange struct {
	Min float64 `json:"min,omitempty"`
	Max float64 `json:"max,omitempty"`
}

// CapabilityVisibleCondition defines when a capability is visible.
type CapabilityVisibleCondition struct {
	Capability string `json:"capability,omitempty"`
	Attribute  string `json:"attribute,omitempty"`
	Value      any    `json:"value,omitempty"`
	Operator   string `json:"operator,omitempty"`
}

// PresentationDashboard defines dashboard presentation.
type PresentationDashboard struct {
	States  []PresentationConfigEntry `json:"states,omitempty"`
	Actions []PresentationConfigEntry `json:"actions,omitempty"`
}

// PresentationAutomation defines automation presentation.
type PresentationAutomation struct {
	Conditions []PresentationConfigEntry `json:"conditions,omitempty"`
	Actions    []PresentationConfigEntry `json:"actions,omitempty"`
}

// PresentationDPInfo contains device presentation info for specific OS/mode.
type PresentationDPInfo struct {
	OS            string `json:"os,omitempty"`
	DPURI         string `json:"dpUri,omitempty"`
	OperatingMode string `json:"operatingMode,omitempty"` // "easySetup" or "deviceControl"
}

// PresentationDeviceConfigCreate is the request body for creating a device configuration.
type PresentationDeviceConfigCreate struct {
	Type       string                    `json:"type,omitempty"` // "profile" or "dth"
	IconURL    string                    `json:"iconUrl,omitempty"`
	Dashboard  *PresentationDashboard    `json:"dashboard,omitempty"`
	DetailView []PresentationConfigEntry `json:"detailView,omitempty"`
	Automation *PresentationAutomation   `json:"automation,omitempty"`
}

// PresentationDeviceConfig represents a device presentation configuration.
type PresentationDeviceConfig struct {
	ManufacturerName string                    `json:"manufacturerName"`
	PresentationID   string                    `json:"presentationId"`
	Type             string                    `json:"type,omitempty"`
	IconURL          string                    `json:"iconUrl,omitempty"`
	Dashboard        *PresentationDashboard    `json:"dashboard,omitempty"`
	DetailView       []PresentationConfigEntry `json:"detailView,omitempty"`
	Automation       *PresentationAutomation   `json:"automation,omitempty"`
	DPInfo           []PresentationDPInfo      `json:"dpInfo,omitempty"`
}

// PresentationDevicePresentation represents a full device presentation.
type PresentationDevicePresentation struct {
	ManufacturerName string                    `json:"manufacturerName"`
	PresentationID   string                    `json:"presentationId"`
	IconURL          string                    `json:"iconUrl,omitempty"`
	Dashboard        *PresentationDashboard    `json:"dashboard,omitempty"`
	DetailView       []PresentationConfigEntry `json:"detailView,omitempty"`
	Automation       *PresentationAutomation   `json:"automation,omitempty"`
	DPInfo           []PresentationDPInfo      `json:"dpInfo,omitempty"`
	Language         []map[string]any          `json:"language,omitempty"`
}

// GeneratePresentation generates a device configuration from a device profile.
func (c *Client) GeneratePresentation(ctx context.Context, profileID string) (*PresentationDeviceConfig, error) {
	if profileID == "" {
		return nil, ErrEmptyProfileID
	}

	data, err := c.get(ctx, "/presentation/types/"+profileID+"/deviceconfig")
	if err != nil {
		return nil, err
	}

	var config PresentationDeviceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("GeneratePresentation: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &config, nil
}

// GetPresentationConfig returns a device configuration by presentation ID.
func (c *Client) GetPresentationConfig(ctx context.Context, presentationID, manufacturerName string) (*PresentationDeviceConfig, error) {
	if presentationID == "" {
		return nil, ErrEmptyPresentationID
	}

	params := url.Values{}
	params.Set("presentationId", presentationID)
	if manufacturerName != "" {
		params.Set("manufacturerName", manufacturerName)
	}

	data, err := c.get(ctx, "/presentation/deviceconfig?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var config PresentationDeviceConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("GetPresentationConfig: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &config, nil
}

// CreatePresentationConfig creates a new device presentation configuration.
func (c *Client) CreatePresentationConfig(ctx context.Context, config *PresentationDeviceConfigCreate) (*PresentationDeviceConfig, error) {
	if config == nil {
		return nil, ErrEmptyPresentationConfig
	}

	data, err := c.post(ctx, "/presentation/deviceconfig", config)
	if err != nil {
		return nil, err
	}

	var created PresentationDeviceConfig
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("CreatePresentationConfig: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// GetDevicePresentation returns the full device presentation for rendering.
func (c *Client) GetDevicePresentation(ctx context.Context, presentationID, manufacturerName string) (*PresentationDevicePresentation, error) {
	if presentationID == "" {
		return nil, ErrEmptyPresentationID
	}

	params := url.Values{}
	params.Set("presentationId", presentationID)
	if manufacturerName != "" {
		params.Set("manufacturerName", manufacturerName)
	}

	data, err := c.get(ctx, "/presentation?"+params.Encode())
	if err != nil {
		return nil, err
	}

	var presentation PresentationDevicePresentation
	if err := json.Unmarshal(data, &presentation); err != nil {
		return nil, fmt.Errorf("GetDevicePresentation: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &presentation, nil
}
