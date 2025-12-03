package smartthings

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// EdgeDriver represents a SmartThings Edge driver.
type EdgeDriver struct {
	DriverID                  string                        `json:"driverId"`
	Name                      string                        `json:"name"`
	Description               string                        `json:"description,omitempty"`
	Version                   string                        `json:"version"`
	PackageKey                string                        `json:"packageKey,omitempty"`
	Permissions               map[string]any                `json:"permissions,omitempty"`
	Fingerprints              []EdgeDriverFingerprint       `json:"fingerprints,omitempty"`
	DeviceIntegrationProfiles []DeviceIntegrationProfileKey `json:"deviceIntegrationProfiles,omitempty"`
}

// EdgeDriverSummary represents summary information for an Edge driver.
type EdgeDriverSummary struct {
	DriverID                  string                        `json:"driverId"`
	Name                      string                        `json:"name"`
	Description               string                        `json:"description,omitempty"`
	Version                   string                        `json:"version"`
	Permissions               map[string]any                `json:"permissions,omitempty"`
	DeviceIntegrationProfiles []DeviceIntegrationProfileKey `json:"deviceIntegrationProfiles,omitempty"`
}

// DeviceIntegrationProfileKey represents a device integration profile reference.
type DeviceIntegrationProfileKey struct {
	ID           string `json:"id"`
	MajorVersion int    `json:"majorVersion,omitempty"`
}

// EdgeDriverFingerprint defines device identification rules for driver matching.
type EdgeDriverFingerprint struct {
	DeviceLabel        string                             `json:"deviceLabel,omitempty"`
	ZigbeeManufacturer *EdgeZigbeeManufacturerFingerprint `json:"zigbeeManufacturer,omitempty"`
	ZigbeeGeneric      *EdgeZigbeeGenericFingerprint      `json:"zigbeeGeneric,omitempty"`
	ZWaveManufacturer  *EdgeZWaveManufacturerFingerprint  `json:"zwaveManufacturer,omitempty"`
	ZWaveGeneric       *EdgeZWaveGenericFingerprint       `json:"zwaveGeneric,omitempty"`
}

// EdgeZigbeeManufacturerFingerprint matches Zigbee devices by manufacturer info.
type EdgeZigbeeManufacturerFingerprint struct {
	Manufacturer string `json:"manufacturer,omitempty"`
	Model        string `json:"model,omitempty"`
}

// EdgeZigbeeGenericFingerprint matches Zigbee devices by generic device type.
type EdgeZigbeeGenericFingerprint struct {
	DeviceIdentifier int `json:"deviceIdentifier,omitempty"`
}

// EdgeZWaveManufacturerFingerprint matches Z-Wave devices by manufacturer info.
type EdgeZWaveManufacturerFingerprint struct {
	ManufacturerID int `json:"manufacturerId,omitempty"`
	ProductType    int `json:"productType,omitempty"`
	ProductID      int `json:"productId,omitempty"`
}

// EdgeZWaveGenericFingerprint matches Z-Wave devices by generic/specific type.
type EdgeZWaveGenericFingerprint struct {
	GenericType  int               `json:"genericType,omitempty"`
	SpecificType int               `json:"specificType,omitempty"`
	CommandClass *EdgeCommandClass `json:"commandClass,omitempty"`
}

// EdgeCommandClass represents Z-Wave command class requirements.
type EdgeCommandClass struct {
	Controlled []string `json:"controlled,omitempty"`
	Supported  []string `json:"supported,omitempty"`
	Either     []string `json:"either,omitempty"`
}

// edgeDriverListResponse is the API response for listing drivers.
type edgeDriverListResponse struct {
	Items []EdgeDriverSummary `json:"items"`
}

// edgeDriverFullListResponse is the API response for listing full driver details.
type edgeDriverFullListResponse struct {
	Items []EdgeDriver `json:"items"`
}

// ListDrivers returns all drivers owned by the user.
func (c *Client) ListDrivers(ctx context.Context) ([]EdgeDriverSummary, error) {
	data, err := c.get(ctx, "/drivers")
	if err != nil {
		return nil, err
	}

	var resp edgeDriverListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListDrivers: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// ListDefaultDrivers returns all SmartThings default drivers.
func (c *Client) ListDefaultDrivers(ctx context.Context) ([]EdgeDriver, error) {
	data, err := c.get(ctx, "/drivers/default")
	if err != nil {
		return nil, err
	}

	var resp edgeDriverFullListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListDefaultDrivers: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetDriver returns a single driver by ID.
func (c *Client) GetDriver(ctx context.Context, driverID string) (*EdgeDriver, error) {
	if driverID == "" {
		return nil, ErrEmptyDriverID
	}

	data, err := c.get(ctx, "/drivers/"+driverID)
	if err != nil {
		return nil, err
	}

	var driver EdgeDriver
	if err := json.Unmarshal(data, &driver); err != nil {
		return nil, fmt.Errorf("GetDriver: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &driver, nil
}

// GetDriverRevision returns a specific version of a driver.
func (c *Client) GetDriverRevision(ctx context.Context, driverID, version string) (*EdgeDriver, error) {
	if driverID == "" {
		return nil, ErrEmptyDriverID
	}
	if version == "" {
		return nil, ErrEmptyDriverVersion
	}

	data, err := c.get(ctx, "/drivers/"+driverID+"/versions/"+version)
	if err != nil {
		return nil, err
	}

	var driver EdgeDriver
	if err := json.Unmarshal(data, &driver); err != nil {
		return nil, fmt.Errorf("GetDriverRevision: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &driver, nil
}

// DeleteDriver deletes a driver.
func (c *Client) DeleteDriver(ctx context.Context, driverID string) error {
	if driverID == "" {
		return ErrEmptyDriverID
	}

	_, err := c.delete(ctx, "/drivers/"+driverID)
	return err
}

// UploadDriver uploads a new driver package.
// The archiveData should be a ZIP archive containing the driver source code.
func (c *Client) UploadDriver(ctx context.Context, archiveData []byte) (*EdgeDriver, error) {
	if len(archiveData) == 0 {
		return nil, ErrEmptyDriverArchive
	}

	// Use custom request for multipart upload
	url := c.baseURL + "/drivers/package"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(archiveData))
	if err != nil {
		return nil, fmt.Errorf("UploadDriver: create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Content-Type", "application/zip")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("UploadDriver: execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("UploadDriver: read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.handleError(resp.StatusCode, respBody)
	}

	var driver EdgeDriver
	if err := json.Unmarshal(respBody, &driver); err != nil {
		return nil, fmt.Errorf("UploadDriver: parse response: %w (body: %s)", err, truncatePreview(respBody))
	}

	return &driver, nil
}
