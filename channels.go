package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// ChannelType represents the type of a driver channel.
type ChannelType string

const (
	// ChannelTypeDriver is a channel containing Edge drivers.
	ChannelTypeDriver ChannelType = "DRIVER"
)

// SubscriberType represents the type of channel subscriber.
type SubscriberType string

const (
	// SubscriberTypeHub indicates hub subscribers.
	SubscriberTypeHub SubscriberType = "HUB"
)

// Channel represents a SmartThings driver channel.
type Channel struct {
	ChannelID         string      `json:"channelId"`
	Name              string      `json:"name"`
	Description       string      `json:"description,omitempty"`
	TermsOfServiceURL string      `json:"termsOfServiceUrl,omitempty"`
	Type              ChannelType `json:"type,omitempty"`
	CreatedDate       string      `json:"createdDate,omitempty"`
	LastModifiedDate  string      `json:"lastModifiedDate,omitempty"`
}

// ChannelCreate is the request body for creating a channel.
type ChannelCreate struct {
	Name              string      `json:"name"`
	Description       string      `json:"description,omitempty"`
	TermsOfServiceURL string      `json:"termsOfServiceUrl,omitempty"`
	Type              ChannelType `json:"type,omitempty"`
}

// ChannelUpdate is the request body for updating a channel.
type ChannelUpdate struct {
	Name              string `json:"name,omitempty"`
	Description       string `json:"description,omitempty"`
	TermsOfServiceURL string `json:"termsOfServiceUrl,omitempty"`
}

// ChannelListOptions contains options for listing channels.
type ChannelListOptions struct {
	SubscriberType  SubscriberType // Filter by subscriber type
	SubscriberID    string         // Filter by subscriber ID (e.g., hub ID)
	IncludeReadOnly bool           // Include read-only channels
}

// DriverChannelDetails represents a driver assigned to a channel.
type DriverChannelDetails struct {
	ChannelID        string `json:"channelId"`
	DriverID         string `json:"driverId"`
	Version          string `json:"version"`
	CreatedDate      string `json:"createdDate,omitempty"`
	LastModifiedDate string `json:"lastModifiedDate,omitempty"`
}

// channelListResponse is the API response for listing channels.
type channelListResponse struct {
	Items []Channel `json:"items"`
}

// driverChannelListResponse is the API response for listing assigned drivers.
type driverChannelListResponse struct {
	Items []DriverChannelDetails `json:"items"`
}

// ListChannels returns all channels accessible to the user.
func (c *Client) ListChannels(ctx context.Context, opts *ChannelListOptions) ([]Channel, error) {
	path := "/channels"

	if opts != nil {
		params := url.Values{}
		if opts.SubscriberType != "" {
			params.Set("subscriberType", string(opts.SubscriberType))
		}
		if opts.SubscriberID != "" {
			params.Set("subscriberId", opts.SubscriberID)
		}
		if opts.IncludeReadOnly {
			params.Set("includeReadOnly", "true")
		}
		if len(params) > 0 {
			path += "?" + params.Encode()
		}
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp channelListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListChannels: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetChannel returns a single channel by ID.
func (c *Client) GetChannel(ctx context.Context, channelID string) (*Channel, error) {
	if channelID == "" {
		return nil, ErrEmptyChannelID
	}

	data, err := c.get(ctx, "/channels/"+channelID)
	if err != nil {
		return nil, err
	}

	var channel Channel
	if err := json.Unmarshal(data, &channel); err != nil {
		return nil, fmt.Errorf("GetChannel: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &channel, nil
}

// CreateChannel creates a new driver channel.
func (c *Client) CreateChannel(ctx context.Context, channel *ChannelCreate) (*Channel, error) {
	if channel == nil || channel.Name == "" {
		return nil, ErrEmptyChannelName
	}

	data, err := c.post(ctx, "/channels", channel)
	if err != nil {
		return nil, err
	}

	var created Channel
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("CreateChannel: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// UpdateChannel updates an existing channel.
func (c *Client) UpdateChannel(ctx context.Context, channelID string, update *ChannelUpdate) (*Channel, error) {
	if channelID == "" {
		return nil, ErrEmptyChannelID
	}

	data, err := c.put(ctx, "/channels/"+channelID, update)
	if err != nil {
		return nil, err
	}

	var updated Channel
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("UpdateChannel: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// DeleteChannel deletes a channel.
func (c *Client) DeleteChannel(ctx context.Context, channelID string) error {
	if channelID == "" {
		return ErrEmptyChannelID
	}

	_, err := c.delete(ctx, "/channels/"+channelID)
	return err
}

// ListAssignedDrivers returns all drivers assigned to a channel.
func (c *Client) ListAssignedDrivers(ctx context.Context, channelID string) ([]DriverChannelDetails, error) {
	if channelID == "" {
		return nil, ErrEmptyChannelID
	}

	data, err := c.get(ctx, "/channels/"+channelID+"/drivers")
	if err != nil {
		return nil, err
	}

	var resp driverChannelListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListAssignedDrivers: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// AssignDriver assigns a driver version to a channel.
func (c *Client) AssignDriver(ctx context.Context, channelID, driverID, version string) (*DriverChannelDetails, error) {
	if channelID == "" {
		return nil, ErrEmptyChannelID
	}
	if driverID == "" {
		return nil, ErrEmptyDriverID
	}
	if version == "" {
		return nil, ErrEmptyDriverVersion
	}

	body := map[string]string{
		"driverId": driverID,
		"version":  version,
	}

	data, err := c.post(ctx, "/channels/"+channelID+"/drivers", body)
	if err != nil {
		return nil, err
	}

	var details DriverChannelDetails
	if err := json.Unmarshal(data, &details); err != nil {
		return nil, fmt.Errorf("AssignDriver: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &details, nil
}

// UnassignDriver removes a driver from a channel.
func (c *Client) UnassignDriver(ctx context.Context, channelID, driverID string) error {
	if channelID == "" {
		return ErrEmptyChannelID
	}
	if driverID == "" {
		return ErrEmptyDriverID
	}

	_, err := c.delete(ctx, "/channels/"+channelID+"/drivers/"+driverID)
	return err
}

// GetDriverChannelMetaInfo returns metadata for a specific driver in a channel.
func (c *Client) GetDriverChannelMetaInfo(ctx context.Context, channelID, driverID string) (*EdgeDriver, error) {
	if channelID == "" {
		return nil, ErrEmptyChannelID
	}
	if driverID == "" {
		return nil, ErrEmptyDriverID
	}

	data, err := c.get(ctx, "/channels/"+channelID+"/drivers/"+driverID+"/meta")
	if err != nil {
		return nil, err
	}

	var driver EdgeDriver
	if err := json.Unmarshal(data, &driver); err != nil {
		return nil, fmt.Errorf("GetDriverChannelMetaInfo: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &driver, nil
}

// EnrollHub enrolls a hub in a channel, allowing it to receive drivers from the channel.
func (c *Client) EnrollHub(ctx context.Context, channelID, hubID string) error {
	if channelID == "" {
		return ErrEmptyChannelID
	}
	if hubID == "" {
		return ErrEmptyHubID
	}

	body := map[string]string{
		"hubId": hubID,
	}

	_, err := c.post(ctx, "/channels/"+channelID+"/hubs", body)
	return err
}

// UnenrollHub removes a hub from a channel.
func (c *Client) UnenrollHub(ctx context.Context, channelID, hubID string) error {
	if channelID == "" {
		return ErrEmptyChannelID
	}
	if hubID == "" {
		return ErrEmptyHubID
	}

	_, err := c.delete(ctx, "/channels/"+channelID+"/hubs/"+hubID)
	return err
}
