package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// Hub represents a SmartThings hub device.
type Hub struct {
	HubID           string `json:"hubId"`
	Name            string `json:"name"`
	EUI             string `json:"eui,omitempty"`
	Owner           string `json:"owner,omitempty"`
	SerialNumber    string `json:"serialNumber,omitempty"`
	FirmwareVersion string `json:"firmwareVersion,omitempty"`
}

// HubCharacteristics contains detailed hub characteristics.
type HubCharacteristics map[string]any

// HubData contains hub-specific data extracted from device status.
// This includes network information like local IP, MAC address, and radio topology.
type HubData struct {
	LocalIP                 string `json:"localIP,omitempty"`
	MacAddress              string `json:"macAddress,omitempty"`
	HardwareType            string `json:"hardwareType,omitempty"`
	HubLocalAPIAvailability string `json:"hubLocalApiAvailability,omitempty"`
	ZigbeeEUI               string `json:"zigbeeEui,omitempty"`
	ZigbeeChannel           int    `json:"zigbeeChannel,omitempty"`
	ZigbeeNodeID            string `json:"zigbeeNodeId,omitempty"`
	ZigbeePanID             string `json:"zigbeePanId,omitempty"`
	ZWaveRegion             string `json:"zwaveRegion,omitempty"`
	ZWaveSUCID              int    `json:"zwaveSucId,omitempty"`
	ZWaveHomeID             string `json:"zwaveHomeId,omitempty"`
	ZWaveNodeID             string `json:"zwaveNodeId,omitempty"`
}

// EnrolledChannel represents a driver channel that a hub is enrolled in.
type EnrolledChannel struct {
	ChannelID        string `json:"channelId"`
	Name             string `json:"name"`
	Description      string `json:"description,omitempty"`
	CreatedDate      string `json:"createdDate,omitempty"`
	LastModifiedDate string `json:"lastModifiedDate,omitempty"`
	SubscriptionURL  string `json:"subscriptionUrl,omitempty"`
}

// InstalledDriver represents a driver installed on a hub.
type InstalledDriver struct {
	DriverID                 string         `json:"driverId"`
	Name                     string         `json:"name"`
	Description              string         `json:"description,omitempty"`
	Version                  string         `json:"version"`
	ChannelID                string         `json:"channelId,omitempty"`
	Developer                string         `json:"developer,omitempty"`
	VendorSupportInformation string         `json:"vendorSupportInformation,omitempty"`
	Permissions              map[string]any `json:"permissions,omitempty"`
}

// enrolledChannelListResponse is the API response for listing enrolled channels.
type enrolledChannelListResponse struct {
	Items []EnrolledChannel `json:"items"`
}

// installedDriverListResponse is the API response for listing installed drivers.
type installedDriverListResponse struct {
	Items []InstalledDriver `json:"items"`
}

// GetHub returns a hub by ID.
// This uses the standard devices endpoint since /hubdevices/{id} is not available.
func (c *Client) GetHub(ctx context.Context, hubID string) (*Hub, error) {
	if hubID == "" {
		return nil, ErrEmptyHubID
	}

	// Use GetDevice since /hubdevices/{id} doesn't exist
	device, err := c.GetDevice(ctx, hubID)
	if err != nil {
		return nil, fmt.Errorf("GetHub: get device: %w", err)
	}

	// Verify it's actually a hub
	if device.Type != DeviceTypeHUB {
		return nil, fmt.Errorf("GetHub: device %s is not a hub (type: %s)", hubID, device.Type)
	}

	hub := &Hub{
		HubID: device.DeviceID,
		Name:  device.Label,
	}
	if hub.Name == "" {
		hub.Name = device.Name
	}
	if device.Hub != nil {
		hub.EUI = device.Hub.HubEUI
		hub.FirmwareVersion = device.Hub.FirmwareVersion
	}

	return hub, nil
}

// GetHubCharacteristics returns detailed characteristics for a hub.
func (c *Client) GetHubCharacteristics(ctx context.Context, hubID string) (HubCharacteristics, error) {
	if hubID == "" {
		return nil, ErrEmptyHubID
	}

	data, err := c.get(ctx, "/hubdevices/"+hubID+"/characteristics")
	if err != nil {
		return nil, err
	}

	var chars HubCharacteristics
	if err := json.Unmarshal(data, &chars); err != nil {
		return nil, fmt.Errorf("GetHubCharacteristics: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return chars, nil
}

// ListEnrolledChannels returns all driver channels a hub is enrolled in.
func (c *Client) ListEnrolledChannels(ctx context.Context, hubID string) ([]EnrolledChannel, error) {
	if hubID == "" {
		return nil, ErrEmptyHubID
	}

	data, err := c.get(ctx, "/hubdevices/"+hubID+"/channels")
	if err != nil {
		return nil, err
	}

	var resp enrolledChannelListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListEnrolledChannels: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// ListInstalledDrivers returns all drivers installed on a hub.
// If deviceID is provided, only returns drivers associated with that device.
func (c *Client) ListInstalledDrivers(ctx context.Context, hubID string, deviceID string) ([]InstalledDriver, error) {
	if hubID == "" {
		return nil, ErrEmptyHubID
	}

	path := "/hubdevices/" + hubID + "/drivers"
	if deviceID != "" {
		path += "?deviceId=" + url.QueryEscape(deviceID)
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp installedDriverListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListInstalledDrivers: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetInstalledDriver returns a specific installed driver on a hub.
func (c *Client) GetInstalledDriver(ctx context.Context, hubID, driverID string) (*InstalledDriver, error) {
	if hubID == "" {
		return nil, ErrEmptyHubID
	}
	if driverID == "" {
		return nil, ErrEmptyDriverID
	}

	data, err := c.get(ctx, "/hubdevices/"+hubID+"/drivers/"+driverID)
	if err != nil {
		return nil, err
	}

	var driver InstalledDriver
	if err := json.Unmarshal(data, &driver); err != nil {
		return nil, fmt.Errorf("GetInstalledDriver: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &driver, nil
}

// InstallDriver installs a driver from a channel onto a hub.
func (c *Client) InstallDriver(ctx context.Context, driverID, hubID, channelID string) error {
	if driverID == "" {
		return ErrEmptyDriverID
	}
	if hubID == "" {
		return ErrEmptyHubID
	}
	if channelID == "" {
		return ErrEmptyChannelID
	}

	body := map[string]string{
		"driverId":  driverID,
		"channelId": channelID,
	}

	_, err := c.post(ctx, "/hubdevices/"+hubID+"/drivers", body)
	return err
}

// UninstallDriver removes a driver from a hub.
func (c *Client) UninstallDriver(ctx context.Context, driverID, hubID string) error {
	if driverID == "" {
		return ErrEmptyDriverID
	}
	if hubID == "" {
		return ErrEmptyHubID
	}

	_, err := c.delete(ctx, "/hubdevices/"+hubID+"/drivers/"+driverID)
	return err
}

// SwitchDriver changes the driver for a device on a hub.
// If forceUpdate is true, forces the driver switch even if versions match.
func (c *Client) SwitchDriver(ctx context.Context, driverID, hubID, deviceID string, forceUpdate bool) error {
	if driverID == "" {
		return ErrEmptyDriverID
	}
	if hubID == "" {
		return ErrEmptyHubID
	}
	if deviceID == "" {
		return ErrEmptyDeviceID
	}

	path := "/hubdevices/" + hubID + "/drivers/" + driverID + "/switch"
	if forceUpdate {
		path += "?forceUpdate=true"
	}

	body := map[string]string{
		"deviceId": deviceID,
	}

	_, err := c.post(ctx, path, body)
	return err
}

// ExtractHubData extracts hub-specific data from a device status response.
// This is useful for getting network information like localIP, macAddress, and radio topology.
// Pass the status map from GetDeviceStatus or GetDeviceStatusAllComponents.
func ExtractHubData(status map[string]any) (*HubData, error) {
	hubData := &HubData{}

	// Look for hubData in the main component or directly in status
	var data map[string]any

	if main, ok := status["main"].(map[string]any); ok {
		if hd, ok := main["hubData"].(map[string]any); ok {
			data = hd
		}
	}

	if data == nil {
		if hd, ok := status["hubData"].(map[string]any); ok {
			data = hd
		}
	}

	if data == nil {
		return hubData, nil // Return empty struct if no hubData found
	}

	// Extract fields using type assertions
	if v, ok := data["localIP"].(string); ok {
		hubData.LocalIP = v
	}
	if v, ok := data["macAddress"].(string); ok {
		hubData.MacAddress = v
	}
	if v, ok := data["hardwareType"].(string); ok {
		hubData.HardwareType = v
	}
	if v, ok := data["hubLocalApiAvailability"].(string); ok {
		hubData.HubLocalAPIAvailability = v
	}
	if v, ok := data["zigbeeEui"].(string); ok {
		hubData.ZigbeeEUI = v
	}
	if v, ok := data["zigbeeChannel"].(float64); ok {
		hubData.ZigbeeChannel = int(v)
	}
	if v, ok := data["zigbeeNodeId"].(string); ok {
		hubData.ZigbeeNodeID = v
	}
	if v, ok := data["zigbeePanId"].(string); ok {
		hubData.ZigbeePanID = v
	}
	if v, ok := data["zwaveRegion"].(string); ok {
		hubData.ZWaveRegion = v
	}
	if v, ok := data["zwaveSucId"].(float64); ok {
		hubData.ZWaveSUCID = int(v)
	}
	if v, ok := data["zwaveHomeId"].(string); ok {
		hubData.ZWaveHomeID = v
	}
	if v, ok := data["zwaveNodeId"].(string); ok {
		hubData.ZWaveNodeID = v
	}

	return hubData, nil
}

// HubLocalConnectionInfo holds configuration for connecting to a hub's local API.
type HubLocalConnectionInfo struct {
	HubID   string
	Name    string
	LocalIP string
	Port    int // Default: 39500
}

// hubListResponse is the API response for listing hubs.
type hubListResponse struct {
	Items []Hub `json:"items"`
}

// ListHubs returns all hubs associated with the account.
// This works by listing all devices and filtering for those with type "HUB".
func (c *Client) ListHubs(ctx context.Context) ([]Hub, error) {
	// Use ListDevicesWithOptions to filter by HUB type
	paged, err := c.ListDevicesWithOptions(ctx, &ListDevicesOptions{
		Type: string(DeviceTypeHUB),
	})
	if err != nil {
		return nil, fmt.Errorf("ListHubs: list devices: %w", err)
	}

	// Convert devices to Hub structs
	hubs := make([]Hub, 0, len(paged.Items))
	for _, device := range paged.Items {
		hub := Hub{
			HubID: device.DeviceID,
			Name:  device.Label,
		}
		if hub.Name == "" {
			hub.Name = device.Name
		}
		// Extract additional info from the Hub field if available
		if device.Hub != nil {
			hub.EUI = device.Hub.HubEUI
			hub.FirmwareVersion = device.Hub.FirmwareVersion
		}
		hubs = append(hubs, hub)
	}

	return hubs, nil
}

// GetHubWithLocalIP returns a hub with its local IP address by fetching the device details.
// The hub device ID can be found by listing devices and filtering for type "HUB".
// The local IP is in device.hub.hubData.localIP, NOT in the device status.
func (c *Client) GetHubWithLocalIP(ctx context.Context, hubID string) (*Hub, *HubData, error) {
	if hubID == "" {
		return nil, nil, ErrEmptyHubID
	}

	// Get device details - the hub info is embedded in the device response
	device, err := c.GetDevice(ctx, hubID)
	if err != nil {
		return nil, nil, fmt.Errorf("get hub device: %w", err)
	}

	// Verify it's actually a hub
	if device.Type != DeviceTypeHUB {
		return nil, nil, fmt.Errorf("device %s is not a hub (type: %s)", hubID, device.Type)
	}

	// Build Hub struct from device
	hub := &Hub{
		HubID: device.DeviceID,
		Name:  device.Label,
	}
	if hub.Name == "" {
		hub.Name = device.Name
	}

	// Extract hub data from device.Hub field
	var hubData *HubData
	if device.Hub != nil {
		hub.EUI = device.Hub.HubEUI
		hub.FirmwareVersion = device.Hub.FirmwareVersion

		// The local IP is in device.hub.hubData
		if device.Hub.HubData != nil {
			hubData = &HubData{
				LocalIP:                 device.Hub.HubData.LocalIP,
				MacAddress:              device.Hub.HubData.MacAddress,
				HardwareType:            device.Hub.HubData.HardwareType,
				HubLocalAPIAvailability: device.Hub.HubData.HubLocalAPIAvailability,
				ZigbeeEUI:               device.Hub.HubData.ZigbeeEUI,
				ZWaveRegion:             device.Hub.HubData.ZWaveRegion,
				ZWaveHomeID:             device.Hub.HubData.ZWaveHomeID,
			}
		}
	}

	return hub, hubData, nil
}

// GetHubLocalConnectionInfo retrieves the information needed to connect to
// a hub's local WebSocket API. Returns the hub ID, name, local IP, and port.
func (c *Client) GetHubLocalConnectionInfo(ctx context.Context, hubID string) (*HubLocalConnectionInfo, error) {
	hub, hubData, err := c.GetHubWithLocalIP(ctx, hubID)
	if err != nil {
		return nil, err
	}

	if hubData == nil || hubData.LocalIP == "" {
		return nil, fmt.Errorf("hub %s does not have a local IP address available", hubID)
	}

	return &HubLocalConnectionInfo{
		HubID:   hub.HubID,
		Name:    hub.Name,
		LocalIP: hubData.LocalIP,
		Port:    HubLocalDefaultPort,
	}, nil
}

// DiscoverHubForLocalConnection finds the first available hub and returns
// connection info for the local WebSocket API.
func (c *Client) DiscoverHubForLocalConnection(ctx context.Context) (*HubLocalConnectionInfo, error) {
	hubs, err := c.ListHubs(ctx)
	if err != nil {
		return nil, fmt.Errorf("list hubs: %w", err)
	}

	if len(hubs) == 0 {
		return nil, fmt.Errorf("no hubs found")
	}

	// Try each hub until we find one with a local IP
	for _, hub := range hubs {
		info, err := c.GetHubLocalConnectionInfo(ctx, hub.HubID)
		if err == nil && info.LocalIP != "" {
			return info, nil
		}
	}

	return nil, fmt.Errorf("no hub with local API available found")
}
