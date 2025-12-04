package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"
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
// Results are cached if caching is enabled.
func (c *Client) GetCapability(ctx context.Context, capabilityID string, version int) (*Capability, error) {
	if capabilityID == "" {
		return nil, ErrEmptyCapabilityID
	}

	path := "/capabilities/" + capabilityID
	if version > 0 {
		path += "/" + strconv.Itoa(version)
	}

	// Use cache if enabled
	ttl := c.getCapabilityTTL()
	if ttl > 0 {
		key := cacheKey("capability", capabilityID, strconv.Itoa(version))
		result, err := c.getCached(key, ttl, func() (any, error) {
			return c.fetchCapability(ctx, path)
		})
		if err != nil {
			return nil, err
		}
		return result.(*Capability), nil
	}

	return c.fetchCapability(ctx, path)
}

// fetchCapability performs the actual API call for GetCapability.
func (c *Client) fetchCapability(ctx context.Context, path string) (*Capability, error) {
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

// getCapabilityTTL returns the TTL for capability caching, or 0 if caching is disabled.
func (c *Client) getCapabilityTTL() time.Duration {
	if c.cacheConfig == nil {
		return 0
	}
	return c.cacheConfig.CapabilityTTL
}

// CapabilityNamespace represents the namespace of a capability.
type CapabilityNamespace string

// Capability namespace constants.
const (
	CapabilityNamespaceSmartThings CapabilityNamespace = "st"
	CapabilityNamespaceCustom      CapabilityNamespace = "custom"
)

// ListCapabilitiesOptions contains options for listing capabilities.
type ListCapabilitiesOptions struct {
	// Namespace filters capabilities by namespace ("st" for standard, "custom" for custom).
	Namespace CapabilityNamespace
}

// ListCapabilitiesWithOptions returns capabilities with filtering options.
func (c *Client) ListCapabilitiesWithOptions(ctx context.Context, opts *ListCapabilitiesOptions) ([]CapabilityReference, error) {
	path := "/capabilities"
	if opts != nil && opts.Namespace != "" {
		path += "?namespace=" + string(opts.Namespace)
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp capabilityListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse capability list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

