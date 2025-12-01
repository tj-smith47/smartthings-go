package smartthings

import (
	"math"
	"time"
)

// Appliance type constants for extraction.
const (
	ApplianceDryer        = "dryer"
	ApplianceWasher       = "washer"
	ApplianceDishwasher   = "dishwasher"
	ApplianceRange        = "range"
	ApplianceRefrigerator = "refrigerator"
)

// ExtractLaundryStatus extracts status from a laundry appliance (washer, dryer, dishwasher).
// The applianceType should be "dryer", "washer", or "dishwasher".
func ExtractLaundryStatus(status Status, applianceType string) *ApplianceStatus {
	// Determine the operating state key based on appliance type
	var opState map[string]any
	var ok bool

	// Try Samsung CE namespace first, then legacy
	switch applianceType {
	case ApplianceDryer:
		opState, ok = GetMap(status, "samsungce.dryerOperatingState")
		if !ok {
			opState, ok = GetMap(status, "dryerOperatingState")
		}
	case ApplianceWasher:
		opState, ok = GetMap(status, "samsungce.washerOperatingState")
		if !ok {
			opState, ok = GetMap(status, "washerOperatingState")
		}
	case ApplianceDishwasher:
		opState, ok = GetMap(status, "samsungce.dishwasherOperatingState")
		if !ok {
			opState, ok = GetMap(status, "dishwasherOperatingState")
		}
	default:
		return nil
	}

	if !ok || opState == nil {
		return nil
	}

	result := &ApplianceStatus{}

	// Check if appliance is running
	isRunning := checkMachineRunning(opState)
	if !isRunning {
		result.State = "idle"
		return result
	}

	result.State = "running"

	// Extract time-related fields
	extractLaundryTimeFields(opState, result)

	return result
}

// checkMachineRunning checks if the appliance is currently running.
func checkMachineRunning(opState map[string]any) bool {
	if opState == nil {
		return false
	}
	// Check machineState first (more reliable)
	if machineState, ok := GetMap(opState, "machineState"); ok {
		if value, ok := GetString(machineState, "value"); ok {
			return value == "running" || value == "run"
		}
	}
	// Fallback to operatingState
	if operatingState, ok := GetMap(opState, "operatingState"); ok {
		if value, ok := GetString(operatingState, "value"); ok {
			return value == "running" || value == "run"
		}
	}
	return false
}

// extractLaundryTimeFields extracts remaining time, completion time, and cycle progress.
func extractLaundryTimeFields(opState map[string]any, result *ApplianceStatus) {
	// Extract remaining time
	if remainingTime, ok := GetMap(opState, "remainingTime"); ok {
		if value, vok := GetFloat(remainingTime, "value"); vok && value > 0 {
			unit, _ := GetString(remainingTime, "unit")
			var mins int
			switch unit {
			case "min", "m":
				mins = int(math.Round(value))
			case "h", "hour", "hours":
				mins = int(math.Round(value * 60))
			case "s", "sec", "second", "seconds", "":
				// Default to seconds if no unit specified
				// Round up to ensure we don't show 0 when there's time remaining
				mins = int(math.Ceil(value / 60))
			default:
				// Unknown unit - assume seconds as safest default
				mins = int(math.Ceil(value / 60))
			}
			if mins < 0 {
				mins = 0
			}
			result.RemainingMins = &mins
		}
	}

	// Extract completion time
	var completionTimeStr string
	if ct, ok := GetMap(opState, "completionTime"); ok {
		if value, vok := GetString(ct, "value"); vok && value != "" && value != "1970-01-01T00:00:00Z" {
			completionTimeStr = value
			result.CompletionTime = &value

			// If remainingTime wasn't provided, calculate from completionTime
			if result.RemainingMins == nil {
				if t, err := time.Parse(time.RFC3339, value); err == nil {
					mins := int(time.Until(t).Minutes())
					if mins > 0 {
						result.RemainingMins = &mins
					}
				}
			}
		}
	}

	// Extract cycle progress
	if progress, ok := GetMap(opState, "progress"); ok {
		if value, vok := GetFloat(progress, "value"); vok {
			p := int(value)
			result.CycleProgress = &p
		}
	}

	// If progress is still nil but we have completionTime, calculate it
	if result.CycleProgress == nil && completionTimeStr != "" {
		if ct, ok := GetMap(opState, "completionTime"); ok {
			if timestamp, tok := GetString(ct, "timestamp"); tok && timestamp != "" {
				startTime, errStart := time.Parse(time.RFC3339, timestamp)
				endTime, errEnd := time.Parse(time.RFC3339, completionTimeStr)
				if errStart == nil && errEnd == nil {
					totalDuration := endTime.Sub(startTime).Minutes()
					elapsed := time.Since(startTime).Minutes()
					if totalDuration > 0 && elapsed > 0 {
						progress := int((elapsed / totalDuration) * 100)
						if progress > 100 {
							progress = 100
						}
						result.CycleProgress = &progress
					}
				}
			}
		}
	}
}

// ExtractRangeStatus extracts status from a Samsung range/oven.
func ExtractRangeStatus(status Status) *RangeStatus {
	result := &RangeStatus{}

	// Extract cooktop active status
	// Path: custom.cooktopOperatingState.cooktopOperatingState.value
	if value, ok := GetString(status, "custom.cooktopOperatingState", "cooktopOperatingState", "value"); ok {
		result.CooktopActive = value == "run"
	}

	// Check if oven is actively running (not just residual heat)
	// Path: ovenOperatingState.machineState.value
	if value, ok := GetString(status, "ovenOperatingState", "machineState", "value"); ok {
		// "ready" means oven is off/idle, anything else means it's running
		result.OvenActive = value != "ready"
	}

	// Only extract temperatures when oven is actively running
	if result.OvenActive {
		// Extract oven target temperature
		// Path: ovenSetpoint.ovenSetpoint.value
		if value, ok := GetFloat(status, "ovenSetpoint", "ovenSetpoint", "value"); ok && value > 0 {
			t := int(value)
			result.OvenTargetTemp = &t
		}

		// Extract current oven temperature
		// Path: temperatureMeasurement.temperature.value
		if value, ok := GetFloat(status, "temperatureMeasurement", "temperature", "value"); ok && value > 0 {
			t := int(value)
			result.OvenTemp = &t
		}
	}

	return result
}

// ExtractRefrigeratorStatus extracts status from a Samsung refrigerator.
// Note: This requires the full component status (from GetDeviceStatusAllComponents),
// not just the main component status.
func ExtractRefrigeratorStatus(allComponents Status) *RefrigeratorStatus {
	result := &RefrigeratorStatus{}

	// Extract fridge temperature from cooler component
	// Path: cooler.temperatureMeasurement.temperature.value
	if cooler, ok := GetMap(allComponents, "cooler"); ok {
		if value, vok := GetFloat(cooler, "temperatureMeasurement", "temperature", "value"); vok {
			// Convert Celsius to Fahrenheit
			fahrenheit := CelsiusToFahrenheit(value)
			result.FridgeTemp = &fahrenheit
		}

		// Extract door open status from cooler component
		// Path: cooler.contactSensor.contact.value
		if value, vok := GetString(cooler, "contactSensor", "contact", "value"); vok {
			result.DoorOpen = value == "open"
		}
	}

	// Extract freezer temperature from freezer component
	// Path: freezer.temperatureMeasurement.temperature.value
	if freezer, ok := GetMap(allComponents, "freezer"); ok {
		if value, vok := GetFloat(freezer, "temperatureMeasurement", "temperature", "value"); vok {
			// Convert Celsius to Fahrenheit
			fahrenheit := CelsiusToFahrenheit(value)
			result.FreezerTemp = &fahrenheit
		}
	}

	return result
}

// ExtractTVFields extracts TV-specific fields from status.
// This is an alias for GetTVStatus for consistency with other Extract* functions.
func ExtractTVFields(status Status) *TVStatus {
	return GetTVStatus(status)
}

// GetApplianceState determines the display state for an appliance based on its status.
// Returns a human-readable state string.
func GetApplianceState(status Status, applianceType string) string {
	switch applianceType {
	case ApplianceDryer:
		if laundryStatus := ExtractLaundryStatus(status, ApplianceDryer); laundryStatus != nil {
			if laundryStatus.State == "running" {
				return "drying"
			}
		}
		return "idle"

	case ApplianceWasher:
		if laundryStatus := ExtractLaundryStatus(status, ApplianceWasher); laundryStatus != nil {
			if laundryStatus.State == "running" {
				return "washing"
			}
		}
		return "idle"

	case ApplianceDishwasher:
		if laundryStatus := ExtractLaundryStatus(status, ApplianceDishwasher); laundryStatus != nil {
			if laundryStatus.State == "running" {
				return "running"
			}
		}
		return "idle"

	case ApplianceRange:
		rangeStatus := ExtractRangeStatus(status)
		if rangeStatus.OvenActive && rangeStatus.CooktopActive {
			return "cooking"
		}
		if rangeStatus.OvenActive {
			return "oven on"
		}
		if rangeStatus.CooktopActive {
			return "cooktop on"
		}
		return "idle"

	case ApplianceRefrigerator:
		// Refrigerator is always "running" when powered on
		return "running"

	default:
		return "unknown"
	}
}

// IsApplianceRunning checks if an appliance is actively doing something.
func IsApplianceRunning(status Status, applianceType string) bool {
	switch applianceType {
	case ApplianceDryer, ApplianceWasher, ApplianceDishwasher:
		if laundryStatus := ExtractLaundryStatus(status, applianceType); laundryStatus != nil {
			return laundryStatus.State == "running"
		}
		return false

	case ApplianceRange:
		rangeStatus := ExtractRangeStatus(status)
		return rangeStatus.OvenActive || rangeStatus.CooktopActive

	case ApplianceRefrigerator:
		// Refrigerator is always "running"
		return true

	default:
		return false
	}
}

// Brilliant Device Helpers

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
