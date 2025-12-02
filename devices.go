package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
)

// truncatePreview returns a truncated string for error messages.
func truncatePreview(data []byte) string {
	s := string(data)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

// ListDevices returns all devices associated with the account.
// For pagination support, use ListDevicesWithOptions instead.
func (c *Client) ListDevices(ctx context.Context) ([]Device, error) {
	data, err := c.get(ctx, "/devices")
	if err != nil {
		return nil, err
	}

	var resp DeviceListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// ListDevicesWithOptions returns devices with pagination and filtering options.
func (c *Client) ListDevicesWithOptions(ctx context.Context, opts *ListDevicesOptions) (*PagedDevices, error) {
	path := "/devices"
	if opts != nil {
		params := url.Values{}
		for _, cap := range opts.Capability {
			params.Add("capability", cap)
		}
		for _, loc := range opts.LocationID {
			params.Add("locationId", loc)
		}
		for _, id := range opts.DeviceID {
			params.Add("deviceId", id)
		}
		if opts.Type != "" {
			params.Set("type", opts.Type)
		}
		if opts.Max > 0 {
			params.Set("max", strconv.Itoa(opts.Max))
		}
		if opts.Page > 0 {
			params.Set("page", strconv.Itoa(opts.Page))
		}
		if opts.IncludeRestricted {
			params.Set("includeRestricted", "true")
		}
		if encoded := params.Encode(); encoded != "" {
			path += "?" + encoded
		}
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp PagedDevices
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device list: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}

// ListAllDevices retrieves all devices by automatically handling pagination.
func (c *Client) ListAllDevices(ctx context.Context) ([]Device, error) {
	var allDevices []Device
	page := 0

	for {
		resp, err := c.ListDevicesWithOptions(ctx, &ListDevicesOptions{
			Max:  200,
			Page: page,
		})
		if err != nil {
			return nil, err
		}

		allDevices = append(allDevices, resp.Items...)

		if resp.Links.Next == "" || len(resp.Items) == 0 {
			break
		}
		page++
	}

	return allDevices, nil
}

// GetDevice returns a single device by ID.
func (c *Client) GetDevice(ctx context.Context, deviceID string) (*Device, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}
	data, err := c.get(ctx, "/devices/"+deviceID)
	if err != nil {
		return nil, err
	}

	var device Device
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("failed to parse device: %w (body: %s)", err, truncatePreview(data))
	}

	return &device, nil
}

// GetDeviceStatus returns the status of the main component of a device.
func (c *Client) GetDeviceStatus(ctx context.Context, deviceID string) (Status, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}
	data, err := c.get(ctx, "/devices/"+deviceID+"/components/main/status")
	if err != nil {
		return nil, err
	}

	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse device status: %w (body: %s)", err, truncatePreview(data))
	}

	return status, nil
}

// GetDeviceFullStatus returns the status of all components of a device.
// The returned map contains component IDs as keys and their status as values.
func (c *Client) GetDeviceFullStatus(ctx context.Context, deviceID string) (map[string]Status, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}
	data, err := c.get(ctx, "/devices/"+deviceID+"/status")
	if err != nil {
		return nil, err
	}

	var resp struct {
		Components map[string]Status `json:"components"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device full status: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Components, nil
}

// GetDeviceStatusAllComponents returns a merged status from all components.
// This is useful for devices like refrigerators where data is split across components.
func (c *Client) GetDeviceStatusAllComponents(ctx context.Context, deviceID string) (Status, error) {
	components, err := c.GetDeviceFullStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}

	// Merge all components into a single status map
	merged := make(Status)
	for componentID, componentStatus := range components {
		merged[componentID] = componentStatus
	}

	return merged, nil
}

// GetComponentStatus returns the status of a specific component.
func (c *Client) GetComponentStatus(ctx context.Context, deviceID, componentID string) (Status, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}
	if componentID == "" {
		return nil, ErrEmptyComponentID
	}
	data, err := c.get(ctx, "/devices/"+deviceID+"/components/"+componentID+"/status")
	if err != nil {
		return nil, err
	}

	var status Status
	if err := json.Unmarshal(data, &status); err != nil {
		return nil, fmt.Errorf("failed to parse component status: %w (body: %s)", err, truncatePreview(data))
	}

	return status, nil
}

// ExecuteCommand sends a single command to a device.
func (c *Client) ExecuteCommand(ctx context.Context, deviceID string, cmd Command) error {
	return c.ExecuteCommands(ctx, deviceID, []Command{cmd})
}

// ExecuteCommands sends multiple commands to a device.
func (c *Client) ExecuteCommands(ctx context.Context, deviceID string, cmds []Command) error {
	if deviceID == "" {
		return ErrEmptyDeviceID
	}
	// Ensure each command has a component (default to "main")
	for i := range cmds {
		if cmds[i].Component == "" {
			cmds[i].Component = "main"
		}
	}

	req := CommandRequest{Commands: cmds}
	_, err := c.post(ctx, "/devices/"+deviceID+"/commands", req)
	return err
}

// DeleteDevice deletes a device.
func (c *Client) DeleteDevice(ctx context.Context, deviceID string) error {
	if deviceID == "" {
		return ErrEmptyDeviceID
	}
	_, err := c.delete(ctx, "/devices/"+deviceID)
	return err
}

// UpdateDevice updates a device (currently only the label can be updated).
func (c *Client) UpdateDevice(ctx context.Context, deviceID string, update *DeviceUpdate) (*Device, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}

	data, err := c.put(ctx, "/devices/"+deviceID, update)
	if err != nil {
		return nil, err
	}

	var device Device
	if err := json.Unmarshal(data, &device); err != nil {
		return nil, fmt.Errorf("failed to parse updated device: %w (body: %s)", err, truncatePreview(data))
	}

	return &device, nil
}

// GetDeviceHealth returns the health status of a device.
func (c *Client) GetDeviceHealth(ctx context.Context, deviceID string) (*DeviceHealth, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}

	data, err := c.get(ctx, "/devices/"+deviceID+"/health")
	if err != nil {
		return nil, err
	}

	var health DeviceHealth
	if err := json.Unmarshal(data, &health); err != nil {
		return nil, fmt.Errorf("failed to parse device health: %w (body: %s)", err, truncatePreview(data))
	}

	return &health, nil
}

// NewCommand creates a command with the given capability and command name.
func NewCommand(capability, command string, args ...any) Command {
	return Command{
		Component:  "main",
		Capability: capability,
		Command:    command,
		Arguments:  args,
	}
}

// NewComponentCommand creates a command for a specific component.
func NewComponentCommand(component, capability, command string, args ...any) Command {
	return Command{
		Component:  component,
		Capability: capability,
		Command:    command,
		Arguments:  args,
	}
}

// FilterDevices returns devices matching the given filter function.
func FilterDevices(devices []Device, filter func(Device) bool) []Device {
	result := make([]Device, 0, len(devices))
	for _, d := range devices {
		if filter(d) {
			result = append(result, d)
		}
	}
	return result
}

// FilterByManufacturer returns devices from a specific manufacturer.
func FilterByManufacturer(devices []Device, manufacturer string) []Device {
	return FilterDevices(devices, func(d Device) bool {
		return d.ManufacturerName == manufacturer
	})
}

// FindDeviceByLabel returns the first device matching the given label.
// Returns a pointer to the device in the slice, or nil if not found.
func FindDeviceByLabel(devices []Device, label string) *Device {
	for i := range devices {
		if devices[i].Label == label {
			return &devices[i]
		}
	}
	return nil
}

// FindDeviceByID returns the device with the given ID.
// Returns a pointer to the device in the slice, or nil if not found.
func FindDeviceByID(devices []Device, deviceID string) *Device {
	for i := range devices {
		if devices[i].DeviceID == deviceID {
			return &devices[i]
		}
	}
	return nil
}
