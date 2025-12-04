package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// DeviceProfileStatus represents the status of a device profile.
type DeviceProfileStatus string

// Device profile status constants.
const (
	ProfileStatusDevelopment DeviceProfileStatus = "DEVELOPMENT"
	ProfileStatusPublished   DeviceProfileStatus = "PUBLISHED"
)

// DeviceProfileFull represents a complete device profile definition.
type DeviceProfileFull struct {
	ID          string              `json:"id"`
	Name        string              `json:"name"`
	Components  []ProfileComponent  `json:"components"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	Status      DeviceProfileStatus `json:"status,omitempty"`
	Preferences []ProfilePreference `json:"preferences,omitempty"`
	Owner       *ProfileOwner       `json:"owner,omitempty"`
}

// ProfileComponent defines a component within a device profile.
type ProfileComponent struct {
	ID           string          `json:"id"`
	Label        string          `json:"label,omitempty"`
	Capabilities []CapabilityRef `json:"capabilities"`
	Categories   []string        `json:"categories,omitempty"`
}

// ProfilePreference defines a user-configurable preference for a device.
type ProfilePreference struct {
	Name        string   `json:"name"`
	Title       string   `json:"title,omitempty"`
	Description string   `json:"description,omitempty"`
	Type        string   `json:"type"` // integer, boolean, string, enumeration
	Required    bool     `json:"required,omitempty"`
	Default     any      `json:"default,omitempty"`
	Minimum     *float64 `json:"minimum,omitempty"`
	Maximum     *float64 `json:"maximum,omitempty"`
	Options     []string `json:"options,omitempty"` // For enumeration type
}

// ProfileOwner represents the owner of a device profile.
type ProfileOwner struct {
	OwnerType string `json:"ownerType"` // USER, ORGANIZATION
	OwnerID   string `json:"ownerId"`
}

// DeviceProfileCreate is the request body for creating a device profile.
type DeviceProfileCreate struct {
	Name        string              `json:"name"`
	Components  []ProfileComponent  `json:"components"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	Preferences []ProfilePreference `json:"preferences,omitempty"`
}

// DeviceProfileUpdate is the request body for updating a device profile.
type DeviceProfileUpdate struct {
	Components  []ProfileComponent  `json:"components,omitempty"`
	Metadata    map[string]string   `json:"metadata,omitempty"`
	Preferences []ProfilePreference `json:"preferences,omitempty"`
}

// profileListResponse is the API response for listing device profiles.
type profileListResponse struct {
	Items []DeviceProfileFull `json:"items"`
}

// ListDeviceProfiles returns all device profiles for the authenticated account.
func (c *Client) ListDeviceProfiles(ctx context.Context) ([]DeviceProfileFull, error) {
	data, err := c.get(ctx, "/deviceprofiles")
	if err != nil {
		return nil, err
	}

	var resp profileListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse device profile list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetDeviceProfile returns a single device profile by ID.
// Results are cached if caching is enabled.
func (c *Client) GetDeviceProfile(ctx context.Context, profileID string) (*DeviceProfileFull, error) {
	if profileID == "" {
		return nil, ErrEmptyProfileID
	}

	// Use cache if enabled
	ttl := c.getDeviceProfileTTL()
	if ttl > 0 {
		key := cacheKey("deviceprofile", profileID)
		result, err := c.getCached(key, ttl, func() (any, error) {
			return c.fetchDeviceProfile(ctx, profileID)
		})
		if err != nil {
			return nil, err
		}
		return result.(*DeviceProfileFull), nil
	}

	return c.fetchDeviceProfile(ctx, profileID)
}

// fetchDeviceProfile performs the actual API call for GetDeviceProfile.
func (c *Client) fetchDeviceProfile(ctx context.Context, profileID string) (*DeviceProfileFull, error) {
	data, err := c.get(ctx, "/deviceprofiles/"+profileID)
	if err != nil {
		return nil, err
	}

	var profile DeviceProfileFull
	if err := json.Unmarshal(data, &profile); err != nil {
		return nil, fmt.Errorf("failed to parse device profile: %w (body: %s)", err, truncatePreview(data))
	}

	return &profile, nil
}

// getDeviceProfileTTL returns the TTL for device profile caching, or 0 if caching is disabled.
func (c *Client) getDeviceProfileTTL() time.Duration {
	if c.cacheConfig == nil {
		return 0
	}
	return c.cacheConfig.DeviceProfileTTL
}

// CreateDeviceProfile creates a new device profile.
func (c *Client) CreateDeviceProfile(ctx context.Context, profile *DeviceProfileCreate) (*DeviceProfileFull, error) {
	if profile == nil || profile.Name == "" {
		return nil, ErrEmptyProfileName
	}

	data, err := c.post(ctx, "/deviceprofiles", profile)
	if err != nil {
		return nil, err
	}

	var created DeviceProfileFull
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created profile: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// UpdateDeviceProfile updates an existing device profile.
func (c *Client) UpdateDeviceProfile(ctx context.Context, profileID string, update *DeviceProfileUpdate) (*DeviceProfileFull, error) {
	if profileID == "" {
		return nil, ErrEmptyProfileID
	}

	data, err := c.put(ctx, "/deviceprofiles/"+profileID, update)
	if err != nil {
		return nil, err
	}

	var updated DeviceProfileFull
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated profile: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// DeleteDeviceProfile removes a device profile.
func (c *Client) DeleteDeviceProfile(ctx context.Context, profileID string) error {
	if profileID == "" {
		return ErrEmptyProfileID
	}

	_, err := c.delete(ctx, "/deviceprofiles/"+profileID)
	return err
}
