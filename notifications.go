package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// NotificationRequestType represents the type of notification.
type NotificationRequestType string

const (
	// NotificationTypeAlert is an alert notification.
	NotificationTypeAlert NotificationRequestType = "ALERT"
	// NotificationTypeSuggestedAction is a suggested action notification.
	NotificationTypeSuggestedAction NotificationRequestType = "SUGGESTED_ACTION"
	// NotificationTypeEventLogging is an event logging notification.
	NotificationTypeEventLogging NotificationRequestType = "EVENT_LOGGING"
	// NotificationTypeAutomationInfo is an automation info notification.
	NotificationTypeAutomationInfo NotificationRequestType = "AUTOMATION_INFO"
)

// DeepLinkType represents the type of deep link.
type DeepLinkType string

const (
	// DeepLinkDevice links to a device.
	DeepLinkDevice DeepLinkType = "device"
	// DeepLinkInstalledApp links to an installed app.
	DeepLinkInstalledApp DeepLinkType = "installedApp"
	// DeepLinkLocation links to a location.
	DeepLinkLocation DeepLinkType = "location"
)

// NotificationMessage contains the title and body for a locale.
type NotificationMessage struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
}

// NotificationReplacement contains a key-value replacement for message templates.
type NotificationReplacement struct {
	Key   string `json:"key,omitempty"`
	Value string `json:"value,omitempty"`
}

// NotificationDeepLink defines a deep link target.
type NotificationDeepLink struct {
	Type DeepLinkType `json:"type"`
	ID   string       `json:"id"`
}

// NotificationRequest is the request body for creating a notification.
type NotificationRequest struct {
	LocationID   string                         `json:"locationId,omitempty"`
	Type         NotificationRequestType        `json:"type"`
	Messages     map[string]NotificationMessage `json:"messages"` // key is locale code (e.g., "en", "ko")
	Replacements []NotificationReplacement      `json:"replacements,omitempty"`
	DeepLink     *NotificationDeepLink          `json:"deepLink,omitempty"`
	ImageURL     string                         `json:"imageUrl,omitempty"`
}

// NotificationResponse is the response from creating a notification.
type NotificationResponse struct {
	Code    int    `json:"code,omitempty"`
	Message string `json:"message,omitempty"`
}

// CreateNotification sends a push notification to mobile apps at a location.
func (c *Client) CreateNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error) {
	if req == nil {
		return nil, ErrEmptyNotificationRequest
	}
	if req.Type == "" {
		return nil, ErrEmptyNotificationType
	}
	if len(req.Messages) == 0 {
		return nil, ErrEmptyNotificationMessages
	}

	data, err := c.post(ctx, "/notifications", req)
	if err != nil {
		return nil, err
	}

	var resp NotificationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreateNotification: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}
