package smartthings

import (
	"errors"
	"fmt"
)

// Sentinel errors returned by the SmartThings client.
// All errors are defined here for easy discovery and consistent organization.
var (
	// Authentication errors
	ErrUnauthorized = errors.New("smartthings: unauthorized (invalid or expired token)")
	ErrEmptyToken   = errors.New("smartthings: API token cannot be empty")

	// Resource errors
	ErrNotFound      = errors.New("smartthings: resource not found")
	ErrDeviceOffline = errors.New("smartthings: device is offline")

	// Rate limiting
	ErrRateLimited = errors.New("smartthings: rate limited (too many requests)")

	// Device validation errors
	ErrEmptyDeviceID    = errors.New("smartthings: device ID cannot be empty")
	ErrEmptyComponentID = errors.New("smartthings: component ID cannot be empty")

	// TV/media validation errors
	ErrEmptyInputID   = errors.New("smartthings: input ID cannot be empty")
	ErrEmptyKey       = errors.New("smartthings: key cannot be empty")
	ErrEmptyAppID     = errors.New("smartthings: app ID cannot be empty")
	ErrEmptyMode      = errors.New("smartthings: mode cannot be empty")
	ErrInvalidChannel = errors.New("smartthings: channel must be non-negative")

	// Location validation errors
	ErrEmptyLocationID   = errors.New("smartthings: location ID cannot be empty")
	ErrEmptyLocationName = errors.New("smartthings: location name cannot be empty")

	// Room validation errors
	ErrEmptyRoomID   = errors.New("smartthings: room ID cannot be empty")
	ErrEmptyRoomName = errors.New("smartthings: room name cannot be empty")

	// Scene validation errors
	ErrEmptySceneID = errors.New("smartthings: scene ID cannot be empty")

	// Rule validation errors
	ErrEmptyRuleID   = errors.New("smartthings: rule ID cannot be empty")
	ErrEmptyRuleName = errors.New("smartthings: rule name cannot be empty")

	// Schedule validation errors
	ErrEmptyScheduleName = errors.New("smartthings: schedule name cannot be empty")

	// InstalledApp/Subscription validation errors
	ErrEmptyInstalledAppID = errors.New("smartthings: installed app ID cannot be empty")
	ErrEmptySubscriptionID = errors.New("smartthings: subscription ID cannot be empty")
	ErrInvalidSubscription = errors.New("smartthings: invalid subscription configuration")

	// Capability validation errors
	ErrEmptyCapabilityID = errors.New("smartthings: capability ID cannot be empty")
)

// APIError represents an error response from the SmartThings API.
type APIError struct {
	StatusCode int
	Message    string
	RequestID  string
}

// Error implements the error interface.
func (e *APIError) Error() string {
	if e.RequestID != "" {
		return fmt.Sprintf("smartthings: API error %d: %s (request_id: %s)", e.StatusCode, e.Message, e.RequestID)
	}
	return fmt.Sprintf("smartthings: API error %d: %s", e.StatusCode, e.Message)
}

// IsUnauthorized returns true if the error indicates an authentication failure.
func IsUnauthorized(err error) bool {
	if errors.Is(err, ErrUnauthorized) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 401
	}
	return false
}

// IsNotFound returns true if the error indicates the resource was not found.
func IsNotFound(err error) bool {
	if errors.Is(err, ErrNotFound) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 404
	}
	return false
}

// IsRateLimited returns true if the error indicates rate limiting.
func IsRateLimited(err error) bool {
	if errors.Is(err, ErrRateLimited) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 429
	}
	return false
}

// IsDeviceOffline returns true if the error indicates the device is offline.
func IsDeviceOffline(err error) bool {
	if errors.Is(err, ErrDeviceOffline) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == 503
	}
	return false
}

// IsTimeout returns true if the error indicates a timeout.
func IsTimeout(err error) bool {
	var netErr interface{ Timeout() bool }
	return errors.As(err, &netErr) && netErr.Timeout()
}
