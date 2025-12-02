package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Mode represents a SmartThings location mode (e.g., "Home", "Away", "Night").
type Mode struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Label string `json:"label,omitempty"`
}

// modeListResponse is the API response for listing modes.
type modeListResponse struct {
	Items []Mode `json:"items"`
}

// currentModeResponse is the API response for the current mode.
type currentModeResponse struct {
	ModeID string `json:"modeId"`
}

// setCurrentModeRequest is the request body for setting the current mode.
type setCurrentModeRequest struct {
	ModeID string `json:"modeId"`
}

// ListModes returns all modes for a location.
func (c *Client) ListModes(ctx context.Context, locationID string) ([]Mode, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}

	data, err := c.get(ctx, "/locations/"+locationID+"/modes")
	if err != nil {
		return nil, err
	}

	var resp modeListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse mode list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetMode returns a single mode by ID.
func (c *Client) GetMode(ctx context.Context, locationID, modeID string) (*Mode, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}
	if modeID == "" {
		return nil, ErrEmptyModeID
	}

	data, err := c.get(ctx, "/locations/"+locationID+"/modes/"+modeID)
	if err != nil {
		return nil, err
	}

	var mode Mode
	if err := json.Unmarshal(data, &mode); err != nil {
		return nil, fmt.Errorf("failed to parse mode: %w (body: %s)", err, truncatePreview(data))
	}

	return &mode, nil
}

// GetCurrentMode returns the currently active mode for a location.
func (c *Client) GetCurrentMode(ctx context.Context, locationID string) (*Mode, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}

	data, err := c.get(ctx, "/locations/"+locationID+"/modes/current")
	if err != nil {
		return nil, err
	}

	var mode Mode
	if err := json.Unmarshal(data, &mode); err != nil {
		return nil, fmt.Errorf("failed to parse current mode: %w (body: %s)", err, truncatePreview(data))
	}

	return &mode, nil
}

// SetCurrentMode changes the active mode for a location.
func (c *Client) SetCurrentMode(ctx context.Context, locationID, modeID string) (*Mode, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}
	if modeID == "" {
		return nil, ErrEmptyModeID
	}

	req := setCurrentModeRequest{ModeID: modeID}
	data, err := c.put(ctx, "/locations/"+locationID+"/modes/current", req)
	if err != nil {
		return nil, err
	}

	var mode Mode
	if err := json.Unmarshal(data, &mode); err != nil {
		return nil, fmt.Errorf("failed to parse mode response: %w (body: %s)", err, truncatePreview(data))
	}

	return &mode, nil
}
