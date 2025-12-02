package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Subscription represents a webhook subscription.
type Subscription struct {
	ID              string                       `json:"id"`
	InstalledAppID  string                       `json:"installedAppId"`
	SourceType      string                       `json:"sourceType"` // DEVICE, CAPABILITY, MODE, DEVICE_LIFECYCLE, etc.
	Device          *DeviceSubscription          `json:"device,omitempty"`
	Capability      *CapabilitySubscription      `json:"capability,omitempty"`
	Mode            *ModeSubscription            `json:"mode,omitempty"`
	DeviceLifecycle *DeviceLifecycleSubscription `json:"deviceLifecycle,omitempty"`
	DeviceHealth    *DeviceHealthSubscription    `json:"deviceHealth,omitempty"`
	SecuritySystem  *SecuritySystemSubscription  `json:"securityArmState,omitempty"`
	HubHealth       *HubHealthSubscription       `json:"hubHealth,omitempty"`
	SceneLifecycle  *SceneLifecycleSubscription  `json:"sceneLifecycle,omitempty"`
}

// DeviceSubscription specifies a device-level subscription.
type DeviceSubscription struct {
	DeviceID        string      `json:"deviceId"`
	ComponentID     string      `json:"componentId,omitempty"`
	Capability      string      `json:"capability,omitempty"`
	Attribute       string      `json:"attribute,omitempty"`
	StateChangeOnly bool        `json:"stateChangeOnly,omitempty"`
	Value           interface{} `json:"value,omitempty"`
	Modes           []string    `json:"modes,omitempty"`
}

// CapabilitySubscription specifies a capability-level subscription.
type CapabilitySubscription struct {
	LocationID      string      `json:"locationId"`
	Capability      string      `json:"capability"`
	Attribute       string      `json:"attribute,omitempty"`
	StateChangeOnly bool        `json:"stateChangeOnly,omitempty"`
	Value           interface{} `json:"value,omitempty"`
	Modes           []string    `json:"modes,omitempty"`
}

// ModeSubscription specifies a mode change subscription.
type ModeSubscription struct {
	LocationID string `json:"locationId"`
}

// DeviceLifecycleSubscription specifies device lifecycle subscription.
type DeviceLifecycleSubscription struct {
	DeviceIDs  []string `json:"deviceIds,omitempty"`
	LocationID string   `json:"locationId,omitempty"`
	Lifecycle  string   `json:"lifecycle,omitempty"` // CREATE, DELETE, UPDATE, etc.
}

// DeviceHealthSubscription specifies device health subscription.
type DeviceHealthSubscription struct {
	DeviceIDs  []string `json:"deviceIds,omitempty"`
	LocationID string   `json:"locationId,omitempty"`
}

// SecuritySystemSubscription specifies security arm state subscription.
type SecuritySystemSubscription struct {
	LocationID string `json:"locationId"`
}

// HubHealthSubscription specifies hub health subscription.
type HubHealthSubscription struct {
	LocationID string `json:"locationId"`
}

// SceneLifecycleSubscription specifies scene lifecycle subscription.
type SceneLifecycleSubscription struct {
	LocationID string `json:"locationId"`
}

// SubscriptionCreate is the request body for creating a subscription.
type SubscriptionCreate struct {
	SourceType      string                       `json:"sourceType"`
	Device          *DeviceSubscription          `json:"device,omitempty"`
	Capability      *CapabilitySubscription      `json:"capability,omitempty"`
	Mode            *ModeSubscription            `json:"mode,omitempty"`
	DeviceLifecycle *DeviceLifecycleSubscription `json:"deviceLifecycle,omitempty"`
	DeviceHealth    *DeviceHealthSubscription    `json:"deviceHealth,omitempty"`
	SecuritySystem  *SecuritySystemSubscription  `json:"securityArmState,omitempty"`
	HubHealth       *HubHealthSubscription       `json:"hubHealth,omitempty"`
	SceneLifecycle  *SceneLifecycleSubscription  `json:"sceneLifecycle,omitempty"`
}

// subscriptionListResponse is the API response for listing subscriptions.
type subscriptionListResponse struct {
	Items []Subscription `json:"items"`
}

// ListSubscriptions returns all subscriptions for an installed app.
func (c *Client) ListSubscriptions(ctx context.Context, installedAppID string) ([]Subscription, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}

	data, err := c.get(ctx, "/installedapps/"+installedAppID+"/subscriptions")
	if err != nil {
		return nil, err
	}

	var resp subscriptionListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse subscription list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// CreateSubscription creates a new subscription.
func (c *Client) CreateSubscription(ctx context.Context, installedAppID string, sub *SubscriptionCreate) (*Subscription, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}
	if sub == nil || sub.SourceType == "" {
		return nil, ErrInvalidSubscription
	}

	data, err := c.post(ctx, "/installedapps/"+installedAppID+"/subscriptions", sub)
	if err != nil {
		return nil, err
	}

	var created Subscription
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created subscription: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// DeleteSubscription deletes a specific subscription.
func (c *Client) DeleteSubscription(ctx context.Context, installedAppID, subscriptionID string) error {
	if installedAppID == "" {
		return ErrEmptyInstalledAppID
	}
	if subscriptionID == "" {
		return ErrEmptySubscriptionID
	}

	_, err := c.delete(ctx, "/installedapps/"+installedAppID+"/subscriptions/"+subscriptionID)
	return err
}

// DeleteAllSubscriptions deletes all subscriptions for an installed app.
func (c *Client) DeleteAllSubscriptions(ctx context.Context, installedAppID string) error {
	if installedAppID == "" {
		return ErrEmptyInstalledAppID
	}

	_, err := c.delete(ctx, "/installedapps/"+installedAppID+"/subscriptions")
	return err
}
