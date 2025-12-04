package smartthings

import (
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"strings"
)

// unmarshalResponse unmarshals JSON data with consistent error formatting.
// This helper reduces boilerplate across all API response parsing.
func unmarshalResponse[T any](data []byte, resourceName string) (*T, error) {
	var resp T
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w (body: %s)", resourceName, err, truncatePreview(data))
	}
	return &resp, nil
}

// truncatePreview returns a truncated string for error messages.
func truncatePreview(data []byte) string {
	s := string(data)
	if len(s) > 200 {
		return s[:200] + "..."
	}
	return s
}

// findCapability searches status for a capability, checking multiple namespaces.
// It tries exact match first, then samsungce.*, custom.*, and samsung.* namespaces.
func findCapability(status Status, names ...string) (map[string]any, string) {
	namespaces := []string{"", "samsungce.", "custom.", "samsung."}
	for _, name := range names {
		for _, ns := range namespaces {
			fullName := ns + name
			if cap, ok := GetMap(status, fullName); ok {
				return cap, fullName
			}
		}
	}
	return nil, ""
}

// DiscoverCapabilities returns all capability names found in a device status.
// Useful for debugging and discovering what capabilities a device supports.
func DiscoverCapabilities(status Status) []string {
	caps := make([]string, 0, len(status))
	for key := range status {
		caps = append(caps, key)
	}
	sort.Strings(caps)
	return caps
}

// FindOperatingStateCapabilities discovers all *OperatingState capabilities in a status.
// Returns a map of capability name to capability data.
func FindOperatingStateCapabilities(status Status) map[string]map[string]any {
	result := make(map[string]map[string]any)
	for key, value := range status {
		if strings.HasSuffix(key, "OperatingState") || strings.Contains(key, "operatingState") {
			if capData, ok := value.(map[string]any); ok {
				result[key] = capData
			}
		}
	}
	return result
}

// GetString navigates a nested map and returns a string value.
// Returns the value and true if found, or empty string and false if not.
//
// Example:
//
//	// Extract: status["switch"]["switch"]["value"]
//	power, ok := GetString(status, "switch", "switch", "value")
func GetString(data map[string]any, keys ...string) (string, bool) {
	val, ok := navigate(data, keys)
	if !ok {
		return "", false
	}
	s, ok := val.(string)
	return s, ok
}

// GetInt navigates a nested map and returns an int value.
// Handles JSON's float64 representation of numbers.
// Returns false if the value is outside the valid int range.
//
// Example:
//
//	// Extract: status["audioVolume"]["volume"]["value"]
//	volume, ok := GetInt(status, "audioVolume", "volume", "value")
func GetInt(data map[string]any, keys ...string) (int, bool) {
	val, ok := navigate(data, keys)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		// Check for overflow before conversion
		if v > float64(math.MaxInt) || v < float64(math.MinInt) || math.IsNaN(v) || math.IsInf(v, 0) {
			return 0, false
		}
		return int(v), true
	case int:
		return v, true
	case int64:
		// Check for overflow on 32-bit systems
		if v > int64(math.MaxInt) || v < int64(math.MinInt) {
			return 0, false
		}
		return int(v), true
	default:
		return 0, false
	}
}

// GetFloat navigates a nested map and returns a float64 value.
//
// Example:
//
//	// Extract: status["temperatureMeasurement"]["temperature"]["value"]
//	temp, ok := GetFloat(status, "temperatureMeasurement", "temperature", "value")
func GetFloat(data map[string]any, keys ...string) (float64, bool) {
	val, ok := navigate(data, keys)
	if !ok {
		return 0, false
	}
	switch v := val.(type) {
	case float64:
		return v, true
	case int:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// GetBool navigates a nested map and returns a bool value.
//
// Example:
//
//	// Extract: status["switch"]["switch"]["value"] == true
//	isOn, ok := GetBool(status, "switch", "switch", "value")
func GetBool(data map[string]any, keys ...string) (bool, bool) {
	val, ok := navigate(data, keys)
	if !ok {
		return false, false
	}
	b, ok := val.(bool)
	return b, ok
}

// GetMap navigates a nested map and returns a map[string]any value.
//
// Example:
//
//	// Extract: status["main"]
//	main, ok := GetMap(status, "main")
func GetMap(data map[string]any, keys ...string) (map[string]any, bool) {
	val, ok := navigate(data, keys)
	if !ok {
		return nil, false
	}
	m, ok := val.(map[string]any)
	return m, ok
}

// GetArray navigates a nested map and returns a []any value.
//
// Example:
//
//	// Extract: status["mediaInputSource"]["supportedInputSources"]["value"]
//	inputs, ok := GetArray(status, "mediaInputSource", "supportedInputSources", "value")
func GetArray(data map[string]any, keys ...string) ([]any, bool) {
	val, ok := navigate(data, keys)
	if !ok {
		return nil, false
	}
	arr, ok := val.([]any)
	return arr, ok
}

// GetStringEquals checks if a nested string value equals the expected value.
//
// Example:
//
//	// Check: status["switch"]["switch"]["value"] == "on"
//	isOn := GetStringEquals(status, "on", "switch", "switch", "value")
func GetStringEquals(data map[string]any, expected string, keys ...string) bool {
	val, ok := GetString(data, keys...)
	return ok && val == expected
}

// navigate walks through a nested map following the provided keys.
// Returns the final value and true if successful, or nil and false if any key is missing.
func navigate(data map[string]any, keys []string) (any, bool) {
	if len(keys) == 0 {
		return data, true
	}

	current := data
	for i, key := range keys {
		val, exists := current[key]
		if !exists {
			return nil, false
		}

		// If this is the last key, return the value
		if i == len(keys)-1 {
			return val, true
		}

		// Otherwise, the value must be a map to continue navigating
		next, ok := val.(map[string]any)
		if !ok {
			return nil, false
		}
		current = next
	}

	return nil, false
}

// CelsiusToFahrenheit converts Celsius to Fahrenheit.
// Returns 0 if the input is NaN, Inf, or would overflow int range.
// For typical home appliance temperatures (-50°C to 500°C), this function
// is safe and accurate.
func CelsiusToFahrenheit(celsius float64) int {
	if math.IsNaN(celsius) || math.IsInf(celsius, 0) {
		return 0
	}
	result := celsius*9/5 + 32
	if result > float64(math.MaxInt32) || result < float64(math.MinInt32) {
		return 0
	}
	return int(result)
}

// FahrenheitToCelsius converts Fahrenheit to Celsius.
func FahrenheitToCelsius(fahrenheit int) float64 {
	return float64(fahrenheit-32) * 5 / 9
}

// ToStringSlice converts a []any to []string, filtering out non-string values.
// Useful for extracting supported options lists from Samsung API responses.
//
// Example:
//
//	arr, _ := GetArray(status, "samsungce.washerCycle", "supportedWasherCycle", "value")
//	cycles := ToStringSlice(arr) // []string{"normal", "delicate", "heavy"}
func ToStringSlice(arr []any) []string {
	if arr == nil {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// ToIntSlice converts a []any to []int, filtering out non-numeric values.
// Handles both int and float64 (JSON number representation).
//
// Example:
//
//	arr, _ := GetArray(status, "supportedTemperatures", "value")
//	temps := ToIntSlice(arr) // []int{34, 36, 38, 40, 42, 44}
func ToIntSlice(arr []any) []int {
	if arr == nil {
		return nil
	}
	result := make([]int, 0, len(arr))
	for _, v := range arr {
		switch n := v.(type) {
		case int:
			result = append(result, n)
		case float64:
			result = append(result, int(n))
		case int64:
			result = append(result, int(n))
		}
	}
	return result
}
