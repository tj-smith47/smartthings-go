package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// VirtualDeviceOwnerType represents the type of virtual device owner.
type VirtualDeviceOwnerType string

const (
	// VirtualDeviceOwnerUser indicates the device is owned by a user.
	VirtualDeviceOwnerUser VirtualDeviceOwnerType = "USER"
	// VirtualDeviceOwnerLocation indicates the device is owned by a location.
	VirtualDeviceOwnerLocation VirtualDeviceOwnerType = "LOCATION"
)

// ExecutionTarget represents where virtual device commands execute.
type ExecutionTarget string

const (
	// ExecutionTargetCloud executes commands in the cloud.
	ExecutionTargetCloud ExecutionTarget = "CLOUD"
	// ExecutionTargetLocal executes commands locally on the hub.
	ExecutionTargetLocal ExecutionTarget = "LOCAL"
)

// VirtualDeviceOwner specifies the owner of a virtual device.
type VirtualDeviceOwner struct {
	OwnerType VirtualDeviceOwnerType `json:"ownerType"`
	OwnerID   string                 `json:"ownerId"`
}

// VirtualDeviceCreateRequest is the request body for creating a virtual device.
type VirtualDeviceCreateRequest struct {
	Name            string              `json:"name"`
	Owner           *VirtualDeviceOwner `json:"owner,omitempty"`
	RoomID          string              `json:"roomId,omitempty"`
	DeviceProfileID string              `json:"deviceProfileId,omitempty"`
	ExecutionTarget ExecutionTarget     `json:"executionTarget,omitempty"`
	HubID           string              `json:"hubId,omitempty"`
}

// VirtualDeviceStandardCreateRequest is the request body for creating a standard virtual device.
type VirtualDeviceStandardCreateRequest struct {
	Name            string              `json:"name"`
	Owner           *VirtualDeviceOwner `json:"owner,omitempty"`
	RoomID          string              `json:"roomId,omitempty"`
	Prototype       string              `json:"prototype"` // e.g., "VIRTUAL_SWITCH", "VIRTUAL_DIMMER"
	ExecutionTarget ExecutionTarget     `json:"executionTarget,omitempty"`
	HubID           string              `json:"hubId,omitempty"`
}

// VirtualDeviceListOptions contains options for listing virtual devices.
type VirtualDeviceListOptions struct {
	LocationID string // Filter by location ID
}

// VirtualDeviceEvent represents an event to create on a virtual device.
type VirtualDeviceEvent struct {
	Component  string `json:"component,omitempty"`
	Capability string `json:"capability"`
	Attribute  string `json:"attribute"`
	Value      any    `json:"value"`
	Unit       string `json:"unit,omitempty"`
}

// VirtualDeviceEventResponse represents a single event response.
type VirtualDeviceEventResponse struct {
	Component   string `json:"component,omitempty"`
	Capability  string `json:"capability,omitempty"`
	Attribute   string `json:"attribute,omitempty"`
	Value       any    `json:"value,omitempty"`
	StateChange bool   `json:"stateChange,omitempty"`
}

// VirtualDeviceEventsResponse is the response from creating virtual device events.
type VirtualDeviceEventsResponse struct {
	StateChanges []VirtualDeviceEventResponse `json:"stateChanges,omitempty"`
}

// ListVirtualDevices returns all virtual devices.
func (c *Client) ListVirtualDevices(ctx context.Context, opts *VirtualDeviceListOptions) ([]Device, error) {
	path := "/virtualdevices"

	if opts != nil {
		params := url.Values{}
		if opts.LocationID != "" {
			params.Set("locationId", opts.LocationID)
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp DeviceListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListVirtualDevices: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// CreateVirtualDevice creates a new virtual device from a device profile.
func (c *Client) CreateVirtualDevice(ctx context.Context, req *VirtualDeviceCreateRequest) (*Device, error) {
	if req == nil || req.Name == "" {
		return nil, ErrEmptyVirtualDeviceName
	}

	data, err := c.post(ctx, "/virtualdevices", req)
	if err != nil {
		return nil, err
	}

	var device Device
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("CreateVirtualDevice: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &device, nil
}

// CreateStandardVirtualDevice creates a new virtual device from a standard prototype.
// Common prototypes include: VIRTUAL_SWITCH, VIRTUAL_DIMMER, VIRTUAL_LOCK, etc.
func (c *Client) CreateStandardVirtualDevice(ctx context.Context, req *VirtualDeviceStandardCreateRequest) (*Device, error) {
	if req == nil || req.Name == "" {
		return nil, ErrEmptyVirtualDeviceName
	}
	if req.Prototype == "" {
		return nil, ErrEmptyPrototype
	}

	data, err := c.post(ctx, "/virtualdevices/prototypes/"+req.Prototype+"/create", req)
	if err != nil {
		return nil, err
	}

	var device Device
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("CreateStandardVirtualDevice: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &device, nil
}

// CreateVirtualDeviceEvents creates events on a virtual device.
// This is used to simulate state changes on virtual devices.
func (c *Client) CreateVirtualDeviceEvents(ctx context.Context, deviceID string, events []VirtualDeviceEvent) (*VirtualDeviceEventsResponse, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}
	if len(events) == 0 {
		return nil, ErrEmptyEvents
	}

	body := map[string][]VirtualDeviceEvent{
		"deviceEvents": events,
	}

	data, err := c.post(ctx, "/virtualdevices/"+deviceID+"/events", body)
	if err != nil {
		return nil, err
	}

	var resp VirtualDeviceEventsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreateVirtualDeviceEvents: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}
