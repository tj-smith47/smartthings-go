package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
)

// Capability represents a SmartThings capability definition.
type Capability struct {
	ID         string                         `json:"id"`
	Version    int                            `json:"version"`
	Status     string                         `json:"status"` // "live", "proposed", "deprecated"
	Name       string                         `json:"name,omitempty"`
	Attributes map[string]CapabilityAttribute `json:"attributes,omitempty"`
	Commands   map[string]CapabilityCommand   `json:"commands,omitempty"`
}

// CapabilityReference is a lightweight reference to a capability.
type CapabilityReference struct {
	ID      string `json:"id"`
	Version int    `json:"version"`
	Status  string `json:"status,omitempty"`
}

// CapabilityAttribute describes a capability attribute.
type CapabilityAttribute struct {
	Schema       AttributeSchema `json:"schema,omitempty"`
	Setter       string          `json:"setter,omitempty"`
	EnumCommands []string        `json:"enumCommands,omitempty"`
}

// AttributeSchema describes the schema for an attribute value.
type AttributeSchema struct {
	Type       string                 `json:"type,omitempty"` // "string", "integer", "number", "boolean", "object", "array"
	Properties map[string]interface{} `json:"properties,omitempty"`
	Required   []string               `json:"required,omitempty"`
	Minimum    *float64               `json:"minimum,omitempty"`
	Maximum    *float64               `json:"maximum,omitempty"`
	Enum       []interface{}          `json:"enum,omitempty"`
}

// CapabilityCommand describes a capability command.
type CapabilityCommand struct {
	Name      string                      `json:"name,omitempty"`
	Arguments []CapabilityCommandArgument `json:"arguments,omitempty"`
}

// CapabilityCommandArgument describes a command argument.
type CapabilityCommandArgument struct {
	Name     string          `json:"name"`
	Optional bool            `json:"optional,omitempty"`
	Schema   AttributeSchema `json:"schema,omitempty"`
}

// capabilityListResponse is the API response for listing capabilities.
type capabilityListResponse struct {
	Items []CapabilityReference `json:"items"`
}

// ListCapabilities returns all available capabilities.
func (c *Client) ListCapabilities(ctx context.Context) ([]CapabilityReference, error) {
	data, err := c.get(ctx, "/capabilities")
	if err != nil {
		return nil, err
	}

	var resp capabilityListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse capability list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetCapability returns a specific capability definition.
// If version is 0, returns the latest version.
func (c *Client) GetCapability(ctx context.Context, capabilityID string, version int) (*Capability, error) {
	if capabilityID == "" {
		return nil, ErrEmptyCapabilityID
	}

	path := "/capabilities/" + capabilityID
	if version > 0 {
		path += "/" + strconv.Itoa(version)
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var cap Capability
	if err := json.Unmarshal(data, &cap); err != nil {
		return nil, fmt.Errorf("failed to parse capability: %w (body: %s)", err, truncatePreview(data))
	}

	return &cap, nil
}
