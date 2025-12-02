package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"time"
)

// DeviceEvent represents a single event from a device's history.
type DeviceEvent struct {
	DeviceID    string    `json:"deviceId"`
	ComponentID string    `json:"componentId"`
	Capability  string    `json:"capability"`
	Attribute   string    `json:"attribute"`
	Value       any       `json:"value"`
	Unit        string    `json:"unit,omitempty"`
	StateChange bool      `json:"stateChange,omitempty"`
	Timestamp   time.Time `json:"time"`
}

// DeviceState represents a historical state snapshot.
type DeviceState struct {
	ComponentID string    `json:"componentId"`
	Capability  string    `json:"capability"`
	Attribute   string    `json:"attribute"`
	Value       any       `json:"value"`
	Unit        string    `json:"unit,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// HistoryOptions configures event/state history queries.
type HistoryOptions struct {
	Before *time.Time // Events before this time (exclusive)
	After  *time.Time // Events after this time (exclusive)
	Max    int        // Max results per page (1-200, default 20)
	Page   int        // Page number (0-based)
}

// PagedEvents is the paginated response for device events.
type PagedEvents struct {
	Items    []DeviceEvent `json:"items"`
	Links    Links         `json:"_links,omitempty"`
	PageInfo PageInfo      `json:"_page,omitempty"`
}

// PagedStates is the paginated response for device states.
type PagedStates struct {
	Items    []DeviceState `json:"items"`
	Links    Links         `json:"_links,omitempty"`
	PageInfo PageInfo      `json:"_page,omitempty"`
}

// buildHistoryQueryParams converts HistoryOptions to URL query parameters.
func buildHistoryQueryParams(opts *HistoryOptions) string {
	if opts == nil {
		return ""
	}

	params := url.Values{}
	if opts.Before != nil {
		params.Set("before", opts.Before.Format(time.RFC3339))
	}
	if opts.After != nil {
		params.Set("after", opts.After.Format(time.RFC3339))
	}
	if opts.Max > 0 {
		params.Set("max", strconv.Itoa(opts.Max))
	}
	if opts.Page > 0 {
		params.Set("page", strconv.Itoa(opts.Page))
	}

	if encoded := params.Encode(); encoded != "" {
		return "?" + encoded
	}
	return ""
}

// GetDeviceEvents returns the event history for a device.
func (c *Client) GetDeviceEvents(ctx context.Context, deviceID string, opts *HistoryOptions) (*PagedEvents, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}

	path := "/devices/" + deviceID + "/events" + buildHistoryQueryParams(opts)
	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp PagedEvents
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device events: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}

// GetDeviceStates returns historical state snapshots for a device.
func (c *Client) GetDeviceStates(ctx context.Context, deviceID string, opts *HistoryOptions) (*PagedStates, error) {
	if deviceID == "" {
		return nil, ErrEmptyDeviceID
	}

	path := "/devices/" + deviceID + "/states" + buildHistoryQueryParams(opts)
	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp PagedStates
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device states: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}
