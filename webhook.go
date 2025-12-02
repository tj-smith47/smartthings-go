package smartthings

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// Webhook signature header name used by SmartThings.
const WebhookSignatureHeader = "X-ST-Signature"

// WebhookLifecycle represents the type of webhook event.
type WebhookLifecycle string

// Webhook lifecycle constants.
const (
	LifecyclePing          WebhookLifecycle = "PING"
	LifecycleConfirmation  WebhookLifecycle = "CONFIRMATION"
	LifecycleConfiguration WebhookLifecycle = "CONFIGURATION"
	LifecycleInstall       WebhookLifecycle = "INSTALL"
	LifecycleUpdate        WebhookLifecycle = "UPDATE"
	LifecycleEvent         WebhookLifecycle = "EVENT"
	LifecycleUninstall     WebhookLifecycle = "UNINSTALL"
	LifecycleOAuthCallback WebhookLifecycle = "OAUTH_CALLBACK"
)

// WebhookEvent represents a webhook request from SmartThings.
type WebhookEvent struct {
	Lifecycle         WebhookLifecycle   `json:"lifecycle"`
	ExecutionID       string             `json:"executionId"`
	AppID             string             `json:"appId"`
	Locale            string             `json:"locale,omitempty"`
	Version           string             `json:"version,omitempty"`
	PingData          *PingData          `json:"pingData,omitempty"`
	ConfirmationData  *ConfirmationData  `json:"confirmationData,omitempty"`
	ConfigurationData *ConfigurationData `json:"configurationData,omitempty"`
	InstallData       *InstallData       `json:"installData,omitempty"`
	UpdateData        *UpdateData        `json:"updateData,omitempty"`
	EventData         *EventData         `json:"eventData,omitempty"`
	UninstallData     *UninstallData     `json:"uninstallData,omitempty"`
	OAuthCallbackData *OAuthCallbackData `json:"oAuthCallbackData,omitempty"`
}

// PingData contains data for PING lifecycle events.
type PingData struct {
	Challenge string `json:"challenge"`
}

// ConfirmationData contains data for CONFIRMATION lifecycle events.
type ConfirmationData struct {
	AppID           string `json:"appId"`
	ConfirmationURL string `json:"confirmationUrl"`
}

// ConfigurationData contains data for CONFIGURATION lifecycle events.
type ConfigurationData struct {
	InstalledAppID string    `json:"installedAppId"`
	Phase          string    `json:"phase"` // INITIALIZE, PAGE
	PageID         string    `json:"pageId,omitempty"`
	PreviousPageID string    `json:"previousPageId,omitempty"`
	Config         ConfigMap `json:"config,omitempty"`
}

// ConfigMap represents SmartApp configuration values.
type ConfigMap map[string][]ConfigEntry

// ConfigEntry represents a single configuration value.
type ConfigEntry struct {
	ValueType    string `json:"valueType"`
	StringConfig *struct {
		Value string `json:"value"`
	} `json:"stringConfig,omitempty"`
	DeviceConfig *struct {
		DeviceID    string `json:"deviceId"`
		ComponentID string `json:"componentId,omitempty"`
	} `json:"deviceConfig,omitempty"`
}

// InstallData contains data for INSTALL lifecycle events.
type InstallData struct {
	AuthToken    string          `json:"authToken"`
	RefreshToken string          `json:"refreshToken"`
	InstalledApp InstalledAppRef `json:"installedApp"`
}

// InstalledAppRef contains reference data for an installed app.
type InstalledAppRef struct {
	InstalledAppID string    `json:"installedAppId"`
	LocationID     string    `json:"locationId"`
	Config         ConfigMap `json:"config,omitempty"`
}

// UpdateData contains data for UPDATE lifecycle events.
type UpdateData struct {
	AuthToken      string          `json:"authToken"`
	RefreshToken   string          `json:"refreshToken"`
	InstalledApp   InstalledAppRef `json:"installedApp"`
	PreviousConfig ConfigMap       `json:"previousConfig,omitempty"`
}

// EventData contains data for EVENT lifecycle events.
type EventData struct {
	AuthToken    string            `json:"authToken"`
	InstalledApp InstalledAppRef   `json:"installedApp"`
	Events       []DeviceEventData `json:"events,omitempty"`
}

// DeviceEventData represents an event from a device.
type DeviceEventData struct {
	EventType   string             `json:"eventType"` // DEVICE_EVENT, TIMER_EVENT, etc.
	DeviceEvent *DeviceEventDetail `json:"deviceEvent,omitempty"`
	TimerEvent  *TimerEventDetail  `json:"timerEvent,omitempty"`
}

// DeviceEventDetail contains details of a device event.
type DeviceEventDetail struct {
	EventID          string `json:"eventId"`
	LocationID       string `json:"locationId"`
	DeviceID         string `json:"deviceId"`
	ComponentID      string `json:"componentId"`
	Capability       string `json:"capability"`
	Attribute        string `json:"attribute"`
	Value            any    `json:"value"`
	StateChange      bool   `json:"stateChange,omitempty"`
	SubscriptionName string `json:"subscriptionName,omitempty"`
}

// TimerEventDetail contains details of a timer event.
type TimerEventDetail struct {
	EventID string `json:"eventId"`
	Name    string `json:"name"`
	Type    string `json:"type"` // CRON, ONCE
	Time    string `json:"time"`
}

// UninstallData contains data for UNINSTALL lifecycle events.
type UninstallData struct {
	InstalledApp InstalledAppRef `json:"installedApp"`
}

// OAuthCallbackData contains data for OAUTH_CALLBACK lifecycle events.
type OAuthCallbackData struct {
	InstalledAppID string `json:"installedAppId"`
	URLPath        string `json:"urlPath"`
}

// Webhook validation errors.
var (
	ErrInvalidSignature = errors.New("smartthings: invalid webhook signature")
	ErrMissingSignature = errors.New("smartthings: missing webhook signature header")
	ErrEmptyBody        = errors.New("smartthings: empty webhook body")
)

// ValidateWebhookSignature verifies the HMAC-SHA256 signature of a webhook request.
// The signature is computed as: base64(hmac-sha256(secret, body))
// Uses constant-time comparison to prevent timing attacks.
func ValidateWebhookSignature(secret string, body []byte, signature string) bool {
	if secret == "" || len(body) == 0 || signature == "" {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expected := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	return hmac.Equal([]byte(expected), []byte(signature))
}

// ParseWebhookRequest parses and validates a webhook request from SmartThings.
// It reads the request body and verifies the signature using the provided secret.
// If secret is empty, signature validation is skipped (not recommended for production).
func ParseWebhookRequest(r *http.Request, secret string) (*WebhookEvent, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read webhook body: %w", err)
	}

	if len(body) == 0 {
		return nil, ErrEmptyBody
	}

	// Validate signature if secret is provided
	if secret != "" {
		signature := r.Header.Get(WebhookSignatureHeader)
		if signature == "" {
			return nil, ErrMissingSignature
		}
		if !ValidateWebhookSignature(secret, body, signature) {
			return nil, ErrInvalidSignature
		}
	}

	var event WebhookEvent
	if err := json.Unmarshal(body, &event); err != nil {
		return nil, fmt.Errorf("failed to parse webhook event: %w", err)
	}

	return &event, nil
}

// PingResponse creates the response for a PING lifecycle event.
func PingResponse(challenge string) map[string]any {
	return map[string]any{
		"pingData": map[string]string{
			"challenge": challenge,
		},
	}
}
