package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Location represents a SmartThings location.
type Location struct {
	LocationID           string                 `json:"locationId"`
	Name                 string                 `json:"name"`
	CountryCode          string                 `json:"countryCode,omitempty"`
	Latitude             float64                `json:"latitude,omitempty"`
	Longitude            float64                `json:"longitude,omitempty"`
	RegionRadius         int                    `json:"regionRadius,omitempty"`
	TemperatureScale     string                 `json:"temperatureScale,omitempty"` // "F" or "C"
	TimeZoneID           string                 `json:"timeZoneId,omitempty"`
	Locale               string                 `json:"locale,omitempty"`
	BackgroundImage      string                 `json:"backgroundImage,omitempty"`
	AdditionalProperties map[string]interface{} `json:"additionalProperties,omitempty"`
	Created              string                 `json:"created,omitempty"`
	LastModified         string                 `json:"lastModified,omitempty"`
}

// LocationCreate is the request body for creating a location.
type LocationCreate struct {
	Name             string  `json:"name"`
	CountryCode      string  `json:"countryCode,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
	RegionRadius     int     `json:"regionRadius,omitempty"`
	TemperatureScale string  `json:"temperatureScale,omitempty"`
	TimeZoneID       string  `json:"timeZoneId,omitempty"`
	Locale           string  `json:"locale,omitempty"`
}

// LocationUpdate is the request body for updating a location.
type LocationUpdate struct {
	Name             string  `json:"name,omitempty"`
	Latitude         float64 `json:"latitude,omitempty"`
	Longitude        float64 `json:"longitude,omitempty"`
	RegionRadius     int     `json:"regionRadius,omitempty"`
	TemperatureScale string  `json:"temperatureScale,omitempty"`
	TimeZoneID       string  `json:"timeZoneId,omitempty"`
	Locale           string  `json:"locale,omitempty"`
}

// locationListResponse is the API response for listing locations.
type locationListResponse struct {
	Items []Location `json:"items"`
}

// ListLocations returns all locations associated with the account.
func (c *Client) ListLocations(ctx context.Context) ([]Location, error) {
	data, err := c.get(ctx, "/locations")
	if err != nil {
		return nil, err
	}

	var resp locationListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse location list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetLocation returns a single location by ID.
func (c *Client) GetLocation(ctx context.Context, locationID string) (*Location, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}

	data, err := c.get(ctx, "/locations/"+locationID)
	if err != nil {
		return nil, err
	}

	var location Location
	if err := json.Unmarshal(data, &location); err != nil {
		return nil, fmt.Errorf("failed to parse location: %w (body: %s)", err, truncatePreview(data))
	}

	return &location, nil
}

// CreateLocation creates a new location.
func (c *Client) CreateLocation(ctx context.Context, location *LocationCreate) (*Location, error) {
	if location == nil || location.Name == "" {
		return nil, ErrEmptyLocationName
	}

	data, err := c.post(ctx, "/locations", location)
	if err != nil {
		return nil, err
	}

	var created Location
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created location: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// UpdateLocation updates an existing location.
func (c *Client) UpdateLocation(ctx context.Context, locationID string, update *LocationUpdate) (*Location, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}

	data, err := c.put(ctx, "/locations/"+locationID, update)
	if err != nil {
		return nil, err
	}

	var updated Location
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated location: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// DeleteLocation deletes a location.
func (c *Client) DeleteLocation(ctx context.Context, locationID string) error {
	if locationID == "" {
		return ErrEmptyLocationID
	}

	_, err := c.delete(ctx, "/locations/"+locationID)
	return err
}
