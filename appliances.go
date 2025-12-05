package smartthings

import (
	"math"
	"strings"
	"time"
)

// Appliance type constants for extraction.
const (
	ApplianceDryer        = "dryer"
	ApplianceWasher       = "washer"
	ApplianceDishwasher   = "dishwasher"
	ApplianceRange        = "range"
	ApplianceRefrigerator = "refrigerator"
	ApplianceMicrowave    = "microwave"
	ApplianceAirConditioner = "airconditioner"
	ApplianceRobotVacuum  = "robotvacuum"
	ApplianceAirPurifier  = "airpurifier"
	ApplianceOven         = "oven"
)

// Samsung capability path constants for common namespaces.
const (
	// Samsung CE (Consumer Electronics) namespace
	nsSamsungCE = "samsungce."
	nsCustom    = "custom."
	nsSamsung   = "samsung."

	// State values
	stateIdle    = "idle"
	stateRunning = "running"
	stateRun     = "run"
	stateReady   = "ready"
	statePaused  = "paused"

	// Epoch timestamp for invalid time check
	epochTimestamp = "1970-01-01T00:00:00Z"
)

// laundryStateNames maps appliance types to their running state display names.
var laundryStateNames = map[string]string{
	ApplianceDryer:      "drying",
	ApplianceWasher:     "washing",
	ApplianceDishwasher: "running",
}

// knownOperatingStatePatterns lists known Samsung CE operating state capability suffixes.
var knownOperatingStatePatterns = []string{
	"washerOperatingState",
	"dryerOperatingState",
	"dishwasherOperatingState",
	"ovenOperatingState",
	"microwaveOperatingState",
	"robotCleanerOperatingState",
	"airConditionerOperatingState",
	"airPurifierOperatingState",
}

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

// ExtractGenericApplianceStatus extracts status from ANY Samsung appliance by
// auto-discovering capabilities. This works with washers, dryers, dishwashers,
// microwaves, air conditioners, robot vacuums, air purifiers, ovens, and more.
//
// The function searches for common Samsung CE capability patterns and extracts
// relevant data into a unified structure. Use DiscoverCapabilities to see what
// capabilities are available on a specific device.
//
// Example:
//
//	status, _ := client.GetDeviceStatus(ctx, deviceID)
//	appStatus := st.ExtractGenericApplianceStatus(status)
//	if appStatus.State == "running" {
//	    fmt.Printf("Appliance running: %d mins remaining\n", *appStatus.RemainingMins)
//	}
func ExtractGenericApplianceStatus(status Status) *GenericApplianceStatus {
	result := &GenericApplianceStatus{
		State: stateIdle,
		Extra: make(map[string]any),
	}

	// Discover all capabilities
	result.DiscoveredCapabilities = DiscoverCapabilities(status)

	// Find and process operating state capabilities
	opStates := FindOperatingStateCapabilities(status)
	for capName, opState := range opStates {
		extractOperatingStateData(opState, result)
		result.Extra["operatingStateCapability"] = capName
		break // Use first found operating state
	}

	// Extract temperature from various capability patterns
	extractGenericTemperature(status, result)

	// Extract door/contact sensor status
	extractContactStatus(status, result)

	// Extract power consumption
	extractPowerConsumption(status, result)

	// Extract mode information
	extractModeInfo(status, result)

	return result
}

// extractOperatingStateData extracts common operating state fields.
func extractOperatingStateData(opState map[string]any, result *GenericApplianceStatus) {
	// Check machine state
	if checkMachineRunning(opState) {
		result.State = stateRunning
	} else if machineState, ok := GetMap(opState, "machineState"); ok {
		if value, vok := GetString(machineState, "value"); vok {
			result.State = value
		}
	}

	// Extract remaining time
	if remainingTime, ok := GetMap(opState, "remainingTime"); ok {
		if value, vok := GetFloat(remainingTime, "value"); vok && value > 0 {
			unit, _ := GetString(remainingTime, "unit")
			mins := convertToMinutes(value, unit)
			result.RemainingMins = &mins
		}
	}

	// Extract completion time
	if ct, ok := GetMap(opState, "completionTime"); ok {
		if value, vok := GetString(ct, "value"); vok && value != "" && value != epochTimestamp {
			result.CompletionTime = &value

			// Calculate remaining if not already set
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

	// Extract progress
	if progress, ok := GetMap(opState, "progress"); ok {
		if value, vok := GetFloat(progress, "value"); vok {
			p := int(value)
			result.Progress = &p
		}
	}
}

// convertToMinutes converts a time value to minutes based on unit.
func convertToMinutes(value float64, unit string) int {
	var mins int
	switch unit {
	case "min", "m":
		mins = int(math.Round(value))
	case "h", "hour", "hours":
		mins = int(math.Round(value * 60))
	case "s", "sec", "second", "seconds", "":
		mins = int(math.Ceil(value / 60))
	default:
		mins = int(math.Ceil(value / 60))
	}
	if mins < 0 {
		mins = 0
	}
	return mins
}

// extractGenericTemperature extracts temperature from various capability patterns.
func extractGenericTemperature(status Status, result *GenericApplianceStatus) {
	// Try temperatureMeasurement capability
	if temp, _ := findCapability(status, "temperatureMeasurement"); temp != nil {
		if value, ok := GetFloat(temp, "temperature", "value"); ok {
			f := CelsiusToFahrenheit(value)
			result.Temperature = &f
		}
	}

	// Try setpoint capabilities for target temperature
	setpointPatterns := []string{"ovenSetpoint", "thermostatCoolingSetpoint", "thermostatHeatingSetpoint"}
	for _, pattern := range setpointPatterns {
		if setpoint, _ := findCapability(status, pattern); setpoint != nil {
			// Try common paths for setpoint value
			for _, path := range []string{pattern, "setpoint", "coolingSetpoint", "heatingSetpoint"} {
				if value, ok := GetFloat(setpoint, path, "value"); ok && value > 0 {
					t := int(value)
					result.TargetTemp = &t
					break
				}
			}
			break
		}
	}
}

// extractContactStatus extracts door/lid open status from contact sensors.
func extractContactStatus(status Status, result *GenericApplianceStatus) {
	if contact, _ := findCapability(status, "contactSensor"); contact != nil {
		if value, ok := GetString(contact, "contact", "value"); ok {
			result.DoorOpen = value == "open"
		}
	}
}

// extractPowerConsumption extracts power usage from various capabilities.
func extractPowerConsumption(status Status, result *GenericApplianceStatus) {
	// Try powerMeter capability
	if power, _ := findCapability(status, "powerMeter"); power != nil {
		if value, ok := GetFloat(power, "power", "value"); ok {
			result.PowerConsumption = &value
		}
	}

	// Try powerConsumptionReport capability (Samsung CE)
	if power, _ := findCapability(status, "powerConsumptionReport"); power != nil {
		if report, ok := GetMap(power, "powerConsumption"); ok {
			if value, vok := GetFloat(report, "value", "power"); vok {
				result.PowerConsumption = &value
			}
		}
	}
}

// extractModeInfo extracts operating mode from various capabilities.
func extractModeInfo(status Status, result *GenericApplianceStatus) {
	// Common mode capability patterns
	modePatterns := []string{
		"airConditionerMode",
		"airPurifierFanMode",
		"robotCleanerMovement",
		"washerMode",
		"dryerMode",
		"ovenMode",
	}

	for _, pattern := range modePatterns {
		if modeCap, capName := findCapability(status, pattern); modeCap != nil {
			// Extract the mode field that matches the capability name
			baseName := strings.TrimPrefix(capName, nsSamsungCE)
			baseName = strings.TrimPrefix(baseName, nsCustom)
			baseName = strings.TrimPrefix(baseName, nsSamsung)

			if modeVal, ok := GetMap(modeCap, baseName); ok {
				if value, vok := GetString(modeVal, "value"); vok {
					result.Mode = value
					return
				}
			}
			// Try generic "mode" field
			if mode, ok := GetMap(modeCap, "mode"); ok {
				if value, vok := GetString(mode, "value"); vok {
					result.Mode = value
					return
				}
			}
		}
	}
}

// GetApplianceState determines the display state for an appliance based on its status.
// Returns a human-readable state string.
func GetApplianceState(status Status, applianceType string) string {
	// Handle laundry appliances (dryer, washer, dishwasher) using lookup table
	if stateName, isLaundry := laundryStateNames[applianceType]; isLaundry {
		if ls := ExtractLaundryStatus(status, applianceType); ls != nil && ls.State == stateRunning {
			return stateName
		}
		return stateIdle
	}

	// Handle range/oven
	if applianceType == ApplianceRange {
		rs := ExtractRangeStatus(status)
		if rs.OvenActive && rs.CooktopActive {
			return "cooking"
		}
		if rs.OvenActive {
			return "oven on"
		}
		if rs.CooktopActive {
			return "cooktop on"
		}
		return stateIdle
	}

	// Refrigerator is always running
	if applianceType == ApplianceRefrigerator {
		return stateRunning
	}

	// TV state based on switch
	if applianceType == "tv" {
		if power, ok := GetString(status, "switch", "switch", "value"); ok {
			return power // "on" or "off"
		}
		return "off"
	}

	// Try generic extraction for unknown appliance types
	if generic := ExtractGenericApplianceStatus(status); generic != nil && generic.State != stateIdle {
		return generic.State
	}

	return "unknown"
}

// IsApplianceRunning checks if an appliance is actively doing something.
func IsApplianceRunning(status Status, applianceType string) bool {
	// Handle laundry appliances using lookup table
	if _, isLaundry := laundryStateNames[applianceType]; isLaundry {
		if ls := ExtractLaundryStatus(status, applianceType); ls != nil {
			return ls.State == stateRunning
		}
		return false
	}

	// Handle range/oven
	if applianceType == ApplianceRange {
		rs := ExtractRangeStatus(status)
		return rs.OvenActive || rs.CooktopActive
	}

	// Refrigerator is always running
	if applianceType == ApplianceRefrigerator {
		return true
	}

	// Fall back to generic extraction for unknown appliance types
	if generic := ExtractGenericApplianceStatus(status); generic != nil {
		return generic.State == stateRunning || generic.State == stateRun
	}

	return false
}

// ==================== Detailed Status Extractors ====================
// These extractors provide comprehensive status including remoteControlStatus
// which is critical for determining if the appliance can be controlled remotely.

// extractRemoteControlEnabled checks if remote control is enabled for an appliance.
// This is CRITICAL - Samsung appliances require this to be enabled on the physical
// device before they accept remote commands. The setting often resets when the
// appliance powers off.
func extractRemoteControlEnabled(status Status) bool {
	// Try samsungce.remoteControlStatus first (most common)
	if value, ok := GetString(status, "samsungce.remoteControlStatus", "remoteControlEnabled", "value"); ok {
		return value == "true"
	}
	// Try without namespace
	if value, ok := GetString(status, "remoteControlStatus", "remoteControlEnabled", "value"); ok {
		return value == "true"
	}
	// Try alternative path structure
	if remoteCtrl, ok := GetMap(status, "samsungce.remoteControlStatus"); ok {
		if enabled, ok := GetMap(remoteCtrl, "remoteControlEnabled"); ok {
			if value, ok := GetString(enabled, "value"); ok {
				return value == "true"
			}
		}
	}
	return false
}

// extractChildLockEnabled checks if the child lock is engaged.
func extractChildLockEnabled(status Status) bool {
	// Try samsungce.kidsLock
	if value, ok := GetString(status, "samsungce.kidsLock", "lockState", "value"); ok {
		return value == "locked"
	}
	// Try alternative path
	if kidsLock, ok := GetMap(status, "samsungce.kidsLock"); ok {
		if lockState, ok := GetMap(kidsLock, "lockState"); ok {
			if value, ok := GetString(lockState, "value"); ok {
				return value == "locked"
			}
		}
	}
	return false
}

// extractSupportedCyclesAndOptions extracts cycle codes and the union of all supported
// options from the nested supportedCycles structure used by Samsung CE appliances.
// The structure is: supportedCycles.value[].cycle and supportedCycles.value[].supportedOptions
func extractSupportedCyclesAndOptions(cycleCap map[string]any) (cycles []string, options map[string][]string) {
	options = make(map[string][]string)
	optionSets := make(map[string]map[string]bool) // For deduplication

	arr, ok := GetArray(cycleCap, "supportedCycles", "value")
	if !ok {
		return nil, options
	}

	for _, item := range arr {
		cycleObj, ok := item.(map[string]any)
		if !ok {
			continue
		}

		// Extract cycle code
		if cycleCode, ok := cycleObj["cycle"].(string); ok && cycleCode != "" {
			cycles = append(cycles, cycleCode)
		}

		// Extract supported options for this cycle
		suppOpts, ok := cycleObj["supportedOptions"].(map[string]any)
		if !ok {
			continue
		}

		// Process each option type (soilLevel, spinLevel, waterTemperature, dryingLevel, etc.)
		for optName, optData := range suppOpts {
			optMap, ok := optData.(map[string]any)
			if !ok {
				continue
			}

			// Get the options array
			optArr, ok := optMap["options"].([]any)
			if !ok {
				continue
			}

			// Initialize set if needed
			if optionSets[optName] == nil {
				optionSets[optName] = make(map[string]bool)
			}

			// Add all options to the set
			for _, opt := range optArr {
				if optStr, ok := opt.(string); ok && optStr != "" {
					optionSets[optName][optStr] = true
				}
			}
		}
	}

	// Convert sets to slices
	for optName, optSet := range optionSets {
		for opt := range optSet {
			options[optName] = append(options[optName], opt)
		}
	}

	return cycles, options
}

// ExtractWasherDetailedStatus extracts comprehensive washer status.
// This includes operating state, current settings, supported options, and
// the critical remoteControlEnabled flag for the UI warning banner.
//
// Example:
//
//	status, _ := client.GetDeviceStatus(ctx, washerID)
//	washer := st.ExtractWasherDetailedStatus(status)
//	if !washer.RemoteControlEnabled {
//	    // Show warning: "Enable Remote Control on the washer to control it"
//	}
func ExtractWasherDetailedStatus(status Status) *WasherDetailedStatus {
	result := &WasherDetailedStatus{
		State: stateIdle,
	}

	// Extract remote control status FIRST (critical for UI)
	result.RemoteControlEnabled = extractRemoteControlEnabled(status)

	// Extract child lock
	result.ChildLockOn = extractChildLockEnabled(status)

	// Extract operating state from laundry status
	if laundry := ExtractLaundryStatus(status, ApplianceWasher); laundry != nil {
		result.State = laundry.State
		result.RemainingMins = laundry.RemainingMins
		result.CompletionTime = laundry.CompletionTime
		result.CycleProgress = laundry.CycleProgress
	}

	// Extract current cycle
	// Try samsungce.washerCycle first, then custom.washerCycle
	if value, ok := GetString(status, "samsungce.washerCycle", "washerCycle", "value"); ok {
		result.CurrentCycle = value
	} else if value, ok := GetString(status, "custom.washerCycle", "washerCycle", "value"); ok {
		result.CurrentCycle = value
	}

	// Extract supported cycles and options from nested supportedCycles structure
	var cycleCap map[string]any
	var ok bool
	if cycleCap, ok = GetMap(status, "samsungce.washerCycle"); !ok {
		cycleCap, _ = GetMap(status, "custom.washerCycle")
	}
	if cycleCap != nil {
		cycles, options := extractSupportedCyclesAndOptions(cycleCap)
		if len(cycles) > 0 {
			result.SupportedCycles = cycles
		}
		if opts := options["waterTemperature"]; len(opts) > 0 {
			result.SupportedWaterTemps = opts
		}
		if opts := options["spinLevel"]; len(opts) > 0 {
			result.SupportedSpinLevels = opts
		}
		if opts := options["soilLevel"]; len(opts) > 0 {
			result.SupportedSoilLevels = opts
		}
	}

	// Fallback: try legacy simple array format if nested extraction found nothing
	if len(result.SupportedCycles) == 0 {
		if arr, ok := GetArray(status, "samsungce.washerCycle", "supportedWasherCycle", "value"); ok {
			result.SupportedCycles = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "custom.washerCycle", "supportedWasherCycle", "value"); ok {
			result.SupportedCycles = ToStringSlice(arr)
		}
	}

	// Extract water temperature (current value)
	if value, ok := GetString(status, "custom.washerWaterTemperature", "washerWaterTemperature", "value"); ok {
		result.WaterTemperature = value
	} else if value, ok := GetString(status, "samsungce.washerWaterTemperature", "washerWaterTemperature", "value"); ok {
		result.WaterTemperature = value
	}

	// Fallback: try legacy simple array format for water temps
	if len(result.SupportedWaterTemps) == 0 {
		if arr, ok := GetArray(status, "custom.washerWaterTemperature", "supportedWasherWaterTemperature", "value"); ok {
			result.SupportedWaterTemps = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "samsungce.washerWaterTemperature", "supportedWasherWaterTemperature", "value"); ok {
			result.SupportedWaterTemps = ToStringSlice(arr)
		}
	}

	// Extract spin level (current value)
	if value, ok := GetString(status, "custom.washerSpinLevel", "washerSpinLevel", "value"); ok {
		result.SpinLevel = value
	} else if value, ok := GetString(status, "samsungce.washerSpinLevel", "washerSpinLevel", "value"); ok {
		result.SpinLevel = value
	}

	// Fallback: try legacy simple array format for spin levels
	if len(result.SupportedSpinLevels) == 0 {
		if arr, ok := GetArray(status, "custom.washerSpinLevel", "supportedWasherSpinLevel", "value"); ok {
			result.SupportedSpinLevels = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "samsungce.washerSpinLevel", "supportedWasherSpinLevel", "value"); ok {
			result.SupportedSpinLevels = ToStringSlice(arr)
		}
	}

	// Extract soil level (current value)
	if value, ok := GetString(status, "custom.washerSoilLevel", "washerSoilLevel", "value"); ok {
		result.SoilLevel = value
	} else if value, ok := GetString(status, "samsungce.washerSoilLevel", "washerSoilLevel", "value"); ok {
		result.SoilLevel = value
	}

	// Fallback: try legacy simple array format for soil levels
	if len(result.SupportedSoilLevels) == 0 {
		if arr, ok := GetArray(status, "custom.washerSoilLevel", "supportedWasherSoilLevel", "value"); ok {
			result.SupportedSoilLevels = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "samsungce.washerSoilLevel", "supportedWasherSoilLevel", "value"); ok {
			result.SupportedSoilLevels = ToStringSlice(arr)
		}
	}

	return result
}

// ExtractDryerDetailedStatus extracts comprehensive dryer status.
// Similar to washer but with dryer-specific fields (temperature, drying level).
//
// Example:
//
//	status, _ := client.GetDeviceStatus(ctx, dryerID)
//	dryer := st.ExtractDryerDetailedStatus(status)
//	if dryer.State == "running" {
//	    fmt.Printf("Drying: %d mins remaining\n", *dryer.RemainingMins)
//	}
func ExtractDryerDetailedStatus(status Status) *DryerDetailedStatus {
	result := &DryerDetailedStatus{
		State: stateIdle,
	}

	// Extract remote control status FIRST (critical for UI)
	result.RemoteControlEnabled = extractRemoteControlEnabled(status)

	// Extract child lock
	result.ChildLockOn = extractChildLockEnabled(status)

	// Extract operating state from laundry status
	if laundry := ExtractLaundryStatus(status, ApplianceDryer); laundry != nil {
		result.State = laundry.State
		result.RemainingMins = laundry.RemainingMins
		result.CompletionTime = laundry.CompletionTime
		result.CycleProgress = laundry.CycleProgress
	}

	// Extract current cycle
	if value, ok := GetString(status, "samsungce.dryerCycle", "dryerCycle", "value"); ok {
		result.CurrentCycle = value
	} else if value, ok := GetString(status, "custom.dryerCycle", "dryerCycle", "value"); ok {
		result.CurrentCycle = value
	}

	// Extract supported cycles and options from nested supportedCycles structure
	var cycleCap map[string]any
	var ok bool
	if cycleCap, ok = GetMap(status, "samsungce.dryerCycle"); !ok {
		cycleCap, _ = GetMap(status, "custom.dryerCycle")
	}
	if cycleCap != nil {
		cycles, options := extractSupportedCyclesAndOptions(cycleCap)
		if len(cycles) > 0 {
			result.SupportedCycles = cycles
		}
		if opts := options["dryingTemperature"]; len(opts) > 0 {
			result.SupportedTemperatures = opts
		}
		if opts := options["dryingLevel"]; len(opts) > 0 {
			result.SupportedDryingLevels = opts
		}
	}

	// Fallback: try legacy simple array format if nested extraction found nothing
	if len(result.SupportedCycles) == 0 {
		if arr, ok := GetArray(status, "samsungce.dryerCycle", "supportedDryerCycle", "value"); ok {
			result.SupportedCycles = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "custom.dryerCycle", "supportedDryerCycle", "value"); ok {
			result.SupportedCycles = ToStringSlice(arr)
		}
	}

	// Extract drying temperature (current value)
	if value, ok := GetString(status, "samsungce.dryerDryingTemperature", "dryingTemperature", "value"); ok {
		result.DryingTemperature = value
	} else if value, ok := GetString(status, "custom.dryerDryingTemperature", "dryingTemperature", "value"); ok {
		result.DryingTemperature = value
	}

	// Fallback: try legacy simple array format for temperatures
	if len(result.SupportedTemperatures) == 0 {
		if arr, ok := GetArray(status, "samsungce.dryerDryingTemperature", "supportedDryingTemperature", "value"); ok {
			result.SupportedTemperatures = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "custom.dryerDryingTemperature", "supportedDryingTemperature", "value"); ok {
			result.SupportedTemperatures = ToStringSlice(arr)
		}
	}

	// Extract drying level (wrinkleFree, normal, more, less)
	if value, ok := GetString(status, "samsungce.dryerDryingLevel", "dryingLevel", "value"); ok {
		result.DryingLevel = value
	} else if value, ok := GetString(status, "custom.dryerDryingLevel", "dryingLevel", "value"); ok {
		result.DryingLevel = value
	}

	// Fallback: try legacy simple array format for drying levels
	if len(result.SupportedDryingLevels) == 0 {
		if arr, ok := GetArray(status, "samsungce.dryerDryingLevel", "supportedDryingLevel", "value"); ok {
			result.SupportedDryingLevels = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "custom.dryerDryingLevel", "supportedDryingLevel", "value"); ok {
			result.SupportedDryingLevels = ToStringSlice(arr)
		}
	}

	// Extract drying time if set
	if value, ok := GetString(status, "samsungce.dryerDryingTime", "dryingTime", "value"); ok {
		result.DryingTime = value
	}

	return result
}

// ExtractDishwasherDetailedStatus extracts comprehensive dishwasher status.
// Dishwashers have simpler controls than washers/dryers.
//
// Example:
//
//	status, _ := client.GetDeviceStatus(ctx, dishwasherID)
//	dishwasher := st.ExtractDishwasherDetailedStatus(status)
func ExtractDishwasherDetailedStatus(status Status) *DishwasherDetailedStatus {
	result := &DishwasherDetailedStatus{
		State: stateIdle,
	}

	// Extract remote control status FIRST (critical for UI)
	result.RemoteControlEnabled = extractRemoteControlEnabled(status)

	// Extract child lock
	result.ChildLockOn = extractChildLockEnabled(status)

	// Extract operating state from laundry status
	if laundry := ExtractLaundryStatus(status, ApplianceDishwasher); laundry != nil {
		result.State = laundry.State
		result.RemainingMins = laundry.RemainingMins
		result.CompletionTime = laundry.CompletionTime
		result.CycleProgress = laundry.CycleProgress
	}

	// Extract current wash course
	if value, ok := GetString(status, "samsungce.dishwasherWashingCourse", "washingCourse", "value"); ok {
		result.CurrentCourse = value
	} else if value, ok := GetString(status, "custom.dishwasherWashingCourse", "washingCourse", "value"); ok {
		result.CurrentCourse = value
	}

	// Extract supported courses (dishwasher uses simple array format: supportedCourses)
	if arr, ok := GetArray(status, "samsungce.dishwasherWashingCourse", "supportedCourses", "value"); ok {
		result.SupportedCourses = ToStringSlice(arr)
	} else if arr, ok := GetArray(status, "custom.dishwasherWashingCourse", "supportedCourses", "value"); ok {
		result.SupportedCourses = ToStringSlice(arr)
	}

	// Fallback: try legacy format (supportedWashingCourse)
	if len(result.SupportedCourses) == 0 {
		if arr, ok := GetArray(status, "samsungce.dishwasherWashingCourse", "supportedWashingCourse", "value"); ok {
			result.SupportedCourses = ToStringSlice(arr)
		} else if arr, ok := GetArray(status, "custom.dishwasherWashingCourse", "supportedWashingCourse", "value"); ok {
			result.SupportedCourses = ToStringSlice(arr)
		}
	}

	return result
}

// ExtractRangeDetailedStatus extracts comprehensive range/oven status.
// Note: Cooktop state is read-only - it cannot be controlled via API for safety.
//
// Example:
//
//	status, _ := client.GetDeviceStatus(ctx, rangeID)
//	rangeStatus := st.ExtractRangeDetailedStatus(status)
//	if rangeStatus.CooktopActive {
//	    fmt.Println("Warning: Cooktop is on (cannot be controlled remotely)")
//	}
func ExtractRangeDetailedStatus(status Status) *RangeDetailedStatus {
	result := &RangeDetailedStatus{}

	// Extract remote control status FIRST (critical for UI)
	result.RemoteControlEnabled = extractRemoteControlEnabled(status)

	// Extract child lock
	result.ChildLockOn = extractChildLockEnabled(status)

	// Use existing ExtractRangeStatus for basic state
	basic := ExtractRangeStatus(status)
	result.CooktopActive = basic.CooktopActive
	result.OvenActive = basic.OvenActive
	result.OvenTemp = basic.OvenTemp
	result.OvenTargetTemp = basic.OvenTargetTemp

	// Extract oven mode
	if value, ok := GetString(status, "ovenMode", "ovenMode", "value"); ok {
		result.OvenMode = value
	} else if value, ok := GetString(status, "samsungce.ovenMode", "ovenMode", "value"); ok {
		result.OvenMode = value
	}

	// Extract supported oven modes
	if arr, ok := GetArray(status, "ovenMode", "supportedOvenModes", "value"); ok {
		result.SupportedOvenModes = ToStringSlice(arr)
	} else if arr, ok := GetArray(status, "samsungce.ovenMode", "supportedOvenModes", "value"); ok {
		result.SupportedOvenModes = ToStringSlice(arr)
	}

	// Extract oven light state
	if value, ok := GetString(status, "samsungce.lamp", "lampState", "value"); ok {
		result.OvenLightOn = value == "on"
	}

	// Extract remaining time from oven operating state
	if opState, ok := GetMap(status, "ovenOperatingState"); ok {
		if remainingTime, ok := GetMap(opState, "remainingTime"); ok {
			if value, ok := GetFloat(remainingTime, "value"); ok && value > 0 {
				unit, _ := GetString(remainingTime, "unit")
				mins := convertToMinutes(value, unit)
				result.RemainingMins = &mins
			}
		}
	}

	// Extract temperature limits if available
	if setpoint, ok := GetMap(status, "ovenSetpoint"); ok {
		if minTemp, ok := GetFloat(setpoint, "ovenSetpoint", "range", "minimum"); ok {
			t := int(minTemp)
			result.OvenTempMin = &t
		}
		if maxTemp, ok := GetFloat(setpoint, "ovenSetpoint", "range", "maximum"); ok {
			t := int(maxTemp)
			result.OvenTempMax = &t
		}
	}

	return result
}
