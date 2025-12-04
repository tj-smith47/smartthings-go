package smartthings

// Brilliant Device Helpers
//
// Brilliant Home Control is a third-party smart home system that integrates
// with SmartThings. These helpers extract status from Brilliant switches
// and dimmers.

// ExtractBrilliantStatus extracts status from a Brilliant smart switch/dimmer.
func ExtractBrilliantStatus(deviceID, name string, status Status) *BrilliantDeviceStatus {
	result := &BrilliantDeviceStatus{
		ID:   deviceID,
		Name: name,
		Type: "switch",
	}

	// Extract switch state
	// Path: switch.switch.value
	if value, ok := GetString(status, "switch", "switch", "value"); ok {
		result.IsOn = value == "on"
	}

	// Check if it's a dimmer by looking for switchLevel
	// Path: switchLevel.level.value
	if level, ok := GetInt(status, "switchLevel", "level", "value"); ok {
		result.Type = "dimmer"
		result.Brightness = &level
	}

	return result
}
