package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
)

// PreferenceType represents the type of device preference.
type PreferenceType string

const (
	// PreferenceTypeInteger is an integer preference.
	PreferenceTypeInteger PreferenceType = "integer"
	// PreferenceTypeNumber is a number (decimal) preference.
	PreferenceTypeNumber PreferenceType = "number"
	// PreferenceTypeBoolean is a boolean preference.
	PreferenceTypeBoolean PreferenceType = "boolean"
	// PreferenceTypeString is a string preference.
	PreferenceTypeString PreferenceType = "string"
	// PreferenceTypeEnumeration is an enumeration preference.
	PreferenceTypeEnumeration PreferenceType = "enumeration"
)

// StringType represents the type of string input.
type StringType string

const (
	// StringTypeText is plain text.
	StringTypeText StringType = "text"
	// StringTypePassword is a password field.
	StringTypePassword StringType = "password"
	// StringTypeEmail is an email address.
	StringTypeEmail StringType = "email"
)

// DevicePreference represents a device preference definition.
type DevicePreference struct {
	PreferenceID   string         `json:"preferenceId"`
	Name           string         `json:"name"`
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	Required       bool           `json:"required,omitempty"`
	PreferenceType PreferenceType `json:"preferenceType"`

	// Type-specific fields
	// Integer
	MinimumInt *int `json:"minimum,omitempty"`
	MaximumInt *int `json:"maximum,omitempty"`
	DefaultInt *int `json:"default,omitempty"`

	// Number (stored in separate fields to avoid JSON conflicts)
	MinimumNum *float64 `json:"-"`
	MaximumNum *float64 `json:"-"`
	DefaultNum *float64 `json:"-"`

	// Boolean
	DefaultBool *bool `json:"-"`

	// String
	MinLength  *int       `json:"minLength,omitempty"`
	MaxLength  *int       `json:"maxLength,omitempty"`
	StringType StringType `json:"stringType,omitempty"`
	DefaultStr *string    `json:"-"`

	// Enumeration
	Options     map[string]string `json:"options,omitempty"`
	DefaultEnum *string           `json:"-"`
}

// DevicePreferenceCreate is the request for creating a device preference.
type DevicePreferenceCreate struct {
	Name           string         `json:"name"`
	Title          string         `json:"title,omitempty"`
	Description    string         `json:"description,omitempty"`
	Required       bool           `json:"required,omitempty"`
	PreferenceType PreferenceType `json:"preferenceType"`

	// Integer/Number type fields (use int for integer, float64 for number)
	Minimum any `json:"minimum,omitempty"`
	Maximum any `json:"maximum,omitempty"`

	// String type fields
	MinLength  *int       `json:"minLength,omitempty"`
	MaxLength  *int       `json:"maxLength,omitempty"`
	StringType StringType `json:"stringType,omitempty"`

	// Enumeration type fields
	Options map[string]string `json:"options,omitempty"`

	// Default value (type depends on PreferenceType)
	Default any `json:"default,omitempty"`
}

// PreferenceLocalization contains localized strings for a preference.
type PreferenceLocalization struct {
	Tag         string            `json:"tag"` // Locale tag, e.g., "en", "ko"
	Label       string            `json:"label,omitempty"`
	Description string            `json:"description,omitempty"`
	Options     map[string]string `json:"options,omitempty"` // For enumeration types
}

// LocaleReference is a reference to a locale.
type LocaleReference struct {
	Tag string `json:"tag"`
}

// devicePreferenceListResponse is the API response for listing preferences.
type devicePreferenceListResponse struct {
	Items []DevicePreference `json:"items"`
}

// localeReferenceListResponse is the API response for listing locale references.
type localeReferenceListResponse struct {
	Items []LocaleReference `json:"items"`
}

// ListDevicePreferences returns all device preferences, optionally filtered by namespace.
func (c *Client) ListDevicePreferences(ctx context.Context, namespace string) ([]DevicePreference, error) {
	path := "/devicepreferences"
	if namespace != "" {
		path += "?namespace=" + url.QueryEscape(namespace)
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp devicePreferenceListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListDevicePreferences: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetDevicePreference returns a single device preference by ID.
func (c *Client) GetDevicePreference(ctx context.Context, preferenceID string) (*DevicePreference, error) {
	if preferenceID == "" {
		return nil, ErrEmptyPreferenceID
	}

	data, err := c.get(ctx, "/devicepreferences/"+preferenceID)
	if err != nil {
		return nil, err
	}

	var pref DevicePreference
	if err := json.Unmarshal(data, &pref); err != nil {
		return nil, fmt.Errorf("GetDevicePreference: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &pref, nil
}

// CreateDevicePreference creates a new device preference.
func (c *Client) CreateDevicePreference(ctx context.Context, pref *DevicePreferenceCreate) (*DevicePreference, error) {
	if pref == nil || pref.Name == "" {
		return nil, ErrEmptyPreferenceName
	}

	data, err := c.post(ctx, "/devicepreferences", pref)
	if err != nil {
		return nil, err
	}

	var created DevicePreference
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("CreateDevicePreference: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// UpdateDevicePreference updates an existing device preference.
func (c *Client) UpdateDevicePreference(ctx context.Context, preferenceID string, pref *DevicePreference) (*DevicePreference, error) {
	if preferenceID == "" {
		return nil, ErrEmptyPreferenceID
	}

	data, err := c.put(ctx, "/devicepreferences/"+preferenceID, pref)
	if err != nil {
		return nil, err
	}

	var updated DevicePreference
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("UpdateDevicePreference: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// CreatePreferenceTranslations creates translations for a device preference.
func (c *Client) CreatePreferenceTranslations(ctx context.Context, preferenceID string, localization *PreferenceLocalization) (*PreferenceLocalization, error) {
	if preferenceID == "" {
		return nil, ErrEmptyPreferenceID
	}
	if localization == nil || localization.Tag == "" {
		return nil, ErrEmptyLocaleTag
	}

	data, err := c.post(ctx, "/devicepreferences/"+preferenceID+"/i18n", localization)
	if err != nil {
		return nil, err
	}

	var created PreferenceLocalization
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("CreatePreferenceTranslations: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// GetPreferenceTranslations returns translations for a specific locale.
func (c *Client) GetPreferenceTranslations(ctx context.Context, preferenceID, locale string) (*PreferenceLocalization, error) {
	if preferenceID == "" {
		return nil, ErrEmptyPreferenceID
	}
	if locale == "" {
		return nil, ErrEmptyLocaleTag
	}

	data, err := c.get(ctx, "/devicepreferences/"+preferenceID+"/i18n/"+locale)
	if err != nil {
		return nil, err
	}

	var localization PreferenceLocalization
	if err := json.Unmarshal(data, &localization); err != nil {
		return nil, fmt.Errorf("GetPreferenceTranslations: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &localization, nil
}

// ListPreferenceTranslations returns all available locales for a preference.
func (c *Client) ListPreferenceTranslations(ctx context.Context, preferenceID string) ([]LocaleReference, error) {
	if preferenceID == "" {
		return nil, ErrEmptyPreferenceID
	}

	data, err := c.get(ctx, "/devicepreferences/"+preferenceID+"/i18n")
	if err != nil {
		return nil, err
	}

	var resp localeReferenceListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListPreferenceTranslations: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// UpdatePreferenceTranslations updates translations for a device preference.
func (c *Client) UpdatePreferenceTranslations(ctx context.Context, preferenceID string, localization *PreferenceLocalization) (*PreferenceLocalization, error) {
	if preferenceID == "" {
		return nil, ErrEmptyPreferenceID
	}
	if localization == nil || localization.Tag == "" {
		return nil, ErrEmptyLocaleTag
	}

	data, err := c.put(ctx, "/devicepreferences/"+preferenceID+"/i18n/"+localization.Tag, localization)
	if err != nil {
		return nil, err
	}

	var updated PreferenceLocalization
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("UpdatePreferenceTranslations: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}
