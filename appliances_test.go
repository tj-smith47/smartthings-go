package smartthings

import (
	"testing"
)

func TestExtractLaundryStatus(t *testing.T) {
	t.Run("dryer running with time remaining", func(t *testing.T) {
		status := Status{
			"samsungce.dryerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"remainingTime": map[string]any{
					"value": float64(45),
					"unit":  "min",
				},
				"progress": map[string]any{"value": float64(60)},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceDryer)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.State != "running" {
			t.Errorf("State = %q, want %q", result.State, "running")
		}
		if result.RemainingMins == nil {
			t.Fatal("RemainingMins is nil")
		}
		if *result.RemainingMins != 45 {
			t.Errorf("RemainingMins = %d, want 45", *result.RemainingMins)
		}
		if result.CycleProgress == nil {
			t.Fatal("CycleProgress is nil")
		}
		if *result.CycleProgress != 60 {
			t.Errorf("CycleProgress = %d, want 60", *result.CycleProgress)
		}
	})

	t.Run("washer idle", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "stop"},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.State != "idle" {
			t.Errorf("State = %q, want %q", result.State, "idle")
		}
		if result.RemainingMins != nil {
			t.Error("RemainingMins should be nil for idle")
		}
	})

	t.Run("dishwasher with seconds unit", func(t *testing.T) {
		status := Status{
			"samsungce.dishwasherOperatingState": map[string]any{
				"machineState": map[string]any{"value": "run"},
				"remainingTime": map[string]any{
					"value": float64(3600), // 60 minutes in seconds
					"unit":  "s",
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceDishwasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.RemainingMins == nil {
			t.Fatal("RemainingMins is nil")
		}
		if *result.RemainingMins != 60 {
			t.Errorf("RemainingMins = %d, want 60", *result.RemainingMins)
		}
	})

	t.Run("washer with hours unit", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"remainingTime": map[string]any{
					"value": float64(2), // 2 hours
					"unit":  "h",
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.RemainingMins == nil {
			t.Fatal("RemainingMins is nil")
		}
		if *result.RemainingMins != 120 {
			t.Errorf("RemainingMins = %d, want 120", *result.RemainingMins)
		}
	})

	t.Run("washer with hour unit (singular)", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"remainingTime": map[string]any{
					"value": float64(1.5),
					"unit":  "hour",
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.RemainingMins == nil {
			t.Fatal("RemainingMins is nil")
		}
		if *result.RemainingMins != 90 {
			t.Errorf("RemainingMins = %d, want 90", *result.RemainingMins)
		}
	})

	t.Run("washer with unknown unit defaults to seconds", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"remainingTime": map[string]any{
					"value": float64(120), // 120 unknown = 2 mins (assuming seconds)
					"unit":  "xyz",
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.RemainingMins == nil {
			t.Fatal("RemainingMins is nil")
		}
		if *result.RemainingMins != 2 {
			t.Errorf("RemainingMins = %d, want 2", *result.RemainingMins)
		}
	})

	t.Run("washer with no unit defaults to seconds", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"remainingTime": map[string]any{
					"value": float64(180), // 180 seconds = 3 mins
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.RemainingMins == nil {
			t.Fatal("RemainingMins is nil")
		}
		if *result.RemainingMins != 3 {
			t.Errorf("RemainingMins = %d, want 3", *result.RemainingMins)
		}
	})

	t.Run("completion time calculates remaining when not provided", func(t *testing.T) {
		// Use a future time to test the calculation
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"completionTime": map[string]any{
					"value": "2099-01-15T15:30:00Z", // Far future
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		// Should have calculated remaining time
		if result.RemainingMins == nil {
			t.Fatal("RemainingMins should be calculated from completion time")
		}
		if *result.RemainingMins <= 0 {
			t.Error("RemainingMins should be positive for future completion time")
		}
	})

	t.Run("legacy namespace fallback", func(t *testing.T) {
		status := Status{
			"dryerOperatingState": map[string]any{
				"operatingState": map[string]any{"value": "run"},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceDryer)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.State != "running" {
			t.Errorf("State = %q, want %q", result.State, "running")
		}
	})

	t.Run("unknown appliance type", func(t *testing.T) {
		status := Status{}
		result := ExtractLaundryStatus(status, "microwave")
		if result != nil {
			t.Error("expected nil for unknown appliance type")
		}
	})

	t.Run("missing operating state", func(t *testing.T) {
		status := Status{}
		result := ExtractLaundryStatus(status, ApplianceDryer)
		if result != nil {
			t.Error("expected nil for missing operating state")
		}
	})

	t.Run("with completion time", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"completionTime": map[string]any{
					"value": "2024-01-15T15:30:00Z",
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.CompletionTime == nil {
			t.Fatal("CompletionTime is nil")
		}
		if *result.CompletionTime != "2024-01-15T15:30:00Z" {
			t.Errorf("CompletionTime = %q, want %q", *result.CompletionTime, "2024-01-15T15:30:00Z")
		}
	})

	t.Run("ignores epoch completion time", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"completionTime": map[string]any{
					"value": "1970-01-01T00:00:00Z",
				},
			},
		}

		result := ExtractLaundryStatus(status, ApplianceWasher)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.CompletionTime != nil {
			t.Error("CompletionTime should be nil for epoch time")
		}
	})
}

func TestCheckMachineRunning(t *testing.T) {
	tests := []struct {
		name    string
		opState map[string]any
		want    bool
	}{
		{
			name: "machineState running",
			opState: map[string]any{
				"machineState": map[string]any{"value": "running"},
			},
			want: true,
		},
		{
			name: "machineState run",
			opState: map[string]any{
				"machineState": map[string]any{"value": "run"},
			},
			want: true,
		},
		{
			name: "machineState stop",
			opState: map[string]any{
				"machineState": map[string]any{"value": "stop"},
			},
			want: false,
		},
		{
			name: "operatingState running (fallback)",
			opState: map[string]any{
				"operatingState": map[string]any{"value": "running"},
			},
			want: true,
		},
		{
			name:    "empty state",
			opState: map[string]any{},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkMachineRunning(tt.opState)
			if got != tt.want {
				t.Errorf("checkMachineRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractRangeStatus(t *testing.T) {
	t.Run("cooktop and oven active", func(t *testing.T) {
		status := Status{
			"custom.cooktopOperatingState": map[string]any{
				"cooktopOperatingState": map[string]any{"value": "run"},
			},
			"ovenOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
			},
			"ovenSetpoint": map[string]any{
				"ovenSetpoint": map[string]any{"value": float64(375)},
			},
			"temperatureMeasurement": map[string]any{
				"temperature": map[string]any{"value": float64(350)},
			},
		}

		result := ExtractRangeStatus(status)
		if result == nil {
			t.Fatal("result is nil")
		}
		if !result.CooktopActive {
			t.Error("CooktopActive should be true")
		}
		if !result.OvenActive {
			t.Error("OvenActive should be true")
		}
		if result.OvenTargetTemp == nil {
			t.Fatal("OvenTargetTemp is nil")
		}
		if *result.OvenTargetTemp != 375 {
			t.Errorf("OvenTargetTemp = %d, want 375", *result.OvenTargetTemp)
		}
		if result.OvenTemp == nil {
			t.Fatal("OvenTemp is nil")
		}
		if *result.OvenTemp != 350 {
			t.Errorf("OvenTemp = %d, want 350", *result.OvenTemp)
		}
	})

	t.Run("only cooktop active", func(t *testing.T) {
		status := Status{
			"custom.cooktopOperatingState": map[string]any{
				"cooktopOperatingState": map[string]any{"value": "run"},
			},
			"ovenOperatingState": map[string]any{
				"machineState": map[string]any{"value": "ready"},
			},
		}

		result := ExtractRangeStatus(status)
		if !result.CooktopActive {
			t.Error("CooktopActive should be true")
		}
		if result.OvenActive {
			t.Error("OvenActive should be false")
		}
		// Temps should be nil when oven is not active
		if result.OvenTemp != nil {
			t.Error("OvenTemp should be nil when oven is not active")
		}
	})

	t.Run("everything off", func(t *testing.T) {
		status := Status{
			"custom.cooktopOperatingState": map[string]any{
				"cooktopOperatingState": map[string]any{"value": "ready"},
			},
			"ovenOperatingState": map[string]any{
				"machineState": map[string]any{"value": "ready"},
			},
		}

		result := ExtractRangeStatus(status)
		if result.CooktopActive {
			t.Error("CooktopActive should be false")
		}
		if result.OvenActive {
			t.Error("OvenActive should be false")
		}
	})

	t.Run("empty status", func(t *testing.T) {
		result := ExtractRangeStatus(Status{})
		if result == nil {
			t.Fatal("result should not be nil")
		}
		if result.CooktopActive {
			t.Error("CooktopActive should be false for empty status")
		}
		if result.OvenActive {
			t.Error("OvenActive should be false for empty status")
		}
	})
}

func TestExtractRefrigeratorStatus(t *testing.T) {
	t.Run("all components present", func(t *testing.T) {
		allComponents := Status{
			"cooler": map[string]any{
				"temperatureMeasurement": map[string]any{
					"temperature": map[string]any{"value": float64(4.0)}, // 4°C = ~39°F
				},
				"contactSensor": map[string]any{
					"contact": map[string]any{"value": "closed"},
				},
			},
			"freezer": map[string]any{
				"temperatureMeasurement": map[string]any{
					"temperature": map[string]any{"value": float64(-18.0)}, // -18°C = ~0°F
				},
			},
		}

		result := ExtractRefrigeratorStatus(allComponents)
		if result == nil {
			t.Fatal("result is nil")
		}
		if result.FridgeTemp == nil {
			t.Fatal("FridgeTemp is nil")
		}
		if *result.FridgeTemp != 39 { // 4°C to F
			t.Errorf("FridgeTemp = %d, want 39", *result.FridgeTemp)
		}
		if result.FreezerTemp == nil {
			t.Fatal("FreezerTemp is nil")
		}
		if *result.FreezerTemp != 0 { // -18°C to F (approximately)
			t.Errorf("FreezerTemp = %d, want 0", *result.FreezerTemp)
		}
		if result.DoorOpen {
			t.Error("DoorOpen should be false")
		}
	})

	t.Run("door open", func(t *testing.T) {
		allComponents := Status{
			"cooler": map[string]any{
				"contactSensor": map[string]any{
					"contact": map[string]any{"value": "open"},
				},
			},
		}

		result := ExtractRefrigeratorStatus(allComponents)
		if !result.DoorOpen {
			t.Error("DoorOpen should be true")
		}
	})

	t.Run("empty status", func(t *testing.T) {
		result := ExtractRefrigeratorStatus(Status{})
		if result == nil {
			t.Fatal("result should not be nil")
		}
		if result.FridgeTemp != nil {
			t.Error("FridgeTemp should be nil")
		}
		if result.FreezerTemp != nil {
			t.Error("FreezerTemp should be nil")
		}
		if result.DoorOpen {
			t.Error("DoorOpen should be false by default")
		}
	})
}

func TestGetApplianceState(t *testing.T) {
	tests := []struct {
		name          string
		status        Status
		applianceType string
		want          string
	}{
		{
			name: "dryer running",
			status: Status{
				"samsungce.dryerOperatingState": map[string]any{
					"machineState": map[string]any{"value": "running"},
				},
			},
			applianceType: ApplianceDryer,
			want:          "drying",
		},
		{
			name: "dryer idle",
			status: Status{
				"samsungce.dryerOperatingState": map[string]any{
					"machineState": map[string]any{"value": "stop"},
				},
			},
			applianceType: ApplianceDryer,
			want:          "idle",
		},
		{
			name: "washer running",
			status: Status{
				"samsungce.washerOperatingState": map[string]any{
					"machineState": map[string]any{"value": "running"},
				},
			},
			applianceType: ApplianceWasher,
			want:          "washing",
		},
		{
			name: "dishwasher running",
			status: Status{
				"samsungce.dishwasherOperatingState": map[string]any{
					"machineState": map[string]any{"value": "run"},
				},
			},
			applianceType: ApplianceDishwasher,
			want:          "running",
		},
		{
			name: "range both active",
			status: Status{
				"custom.cooktopOperatingState": map[string]any{
					"cooktopOperatingState": map[string]any{"value": "run"},
				},
				"ovenOperatingState": map[string]any{
					"machineState": map[string]any{"value": "running"},
				},
			},
			applianceType: ApplianceRange,
			want:          "cooking",
		},
		{
			name: "range oven only",
			status: Status{
				"ovenOperatingState": map[string]any{
					"machineState": map[string]any{"value": "running"},
				},
			},
			applianceType: ApplianceRange,
			want:          "oven on",
		},
		{
			name: "range cooktop only",
			status: Status{
				"custom.cooktopOperatingState": map[string]any{
					"cooktopOperatingState": map[string]any{"value": "run"},
				},
			},
			applianceType: ApplianceRange,
			want:          "cooktop on",
		},
		{
			name:          "refrigerator always running",
			status:        Status{},
			applianceType: ApplianceRefrigerator,
			want:          "running",
		},
		{
			name:          "unknown type",
			status:        Status{},
			applianceType: "toaster",
			want:          "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetApplianceState(tt.status, tt.applianceType)
			if got != tt.want {
				t.Errorf("GetApplianceState() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestIsApplianceRunning(t *testing.T) {
	tests := []struct {
		name          string
		status        Status
		applianceType string
		want          bool
	}{
		{
			name: "dryer running",
			status: Status{
				"samsungce.dryerOperatingState": map[string]any{
					"machineState": map[string]any{"value": "running"},
				},
			},
			applianceType: ApplianceDryer,
			want:          true,
		},
		{
			name:          "dryer not running (missing)",
			status:        Status{},
			applianceType: ApplianceDryer,
			want:          false,
		},
		{
			name: "range cooktop active",
			status: Status{
				"custom.cooktopOperatingState": map[string]any{
					"cooktopOperatingState": map[string]any{"value": "run"},
				},
			},
			applianceType: ApplianceRange,
			want:          true,
		},
		{
			name:          "refrigerator always running",
			status:        Status{},
			applianceType: ApplianceRefrigerator,
			want:          true,
		},
		{
			name:          "unknown type",
			status:        Status{},
			applianceType: "blender",
			want:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsApplianceRunning(tt.status, tt.applianceType)
			if got != tt.want {
				t.Errorf("IsApplianceRunning() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtractBrilliantStatus(t *testing.T) {
	t.Run("switch on", func(t *testing.T) {
		status := Status{
			"switch": map[string]any{
				"switch": map[string]any{"value": "on"},
			},
		}

		result := ExtractBrilliantStatus("device-1", "Living Room", status)
		if result.ID != "device-1" {
			t.Errorf("ID = %q, want %q", result.ID, "device-1")
		}
		if result.Name != "Living Room" {
			t.Errorf("Name = %q, want %q", result.Name, "Living Room")
		}
		if !result.IsOn {
			t.Error("IsOn should be true")
		}
		if result.Type != "switch" {
			t.Errorf("Type = %q, want %q", result.Type, "switch")
		}
		if result.Brightness != nil {
			t.Error("Brightness should be nil for switch")
		}
	})

	t.Run("dimmer with brightness", func(t *testing.T) {
		status := Status{
			"switch": map[string]any{
				"switch": map[string]any{"value": "on"},
			},
			"switchLevel": map[string]any{
				"level": map[string]any{"value": float64(75)},
			},
		}

		result := ExtractBrilliantStatus("device-2", "Dimmer", status)
		if result.Type != "dimmer" {
			t.Errorf("Type = %q, want %q", result.Type, "dimmer")
		}
		if result.Brightness == nil {
			t.Fatal("Brightness is nil")
		}
		if *result.Brightness != 75 {
			t.Errorf("Brightness = %d, want 75", *result.Brightness)
		}
	})

	t.Run("switch off", func(t *testing.T) {
		status := Status{
			"switch": map[string]any{
				"switch": map[string]any{"value": "off"},
			},
		}

		result := ExtractBrilliantStatus("device-3", "Light", status)
		if result.IsOn {
			t.Error("IsOn should be false")
		}
	})
}

func TestExtractGenericApplianceStatus(t *testing.T) {
	t.Run("washer running with all fields", func(t *testing.T) {
		status := Status{
			"samsungce.washerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
				"remainingTime": map[string]any{
					"value": float64(30),
					"unit":  "min",
				},
				"completionTime": map[string]any{
					"value": "2025-12-04T15:30:00Z",
				},
				"progress": map[string]any{"value": float64(45)},
			},
			"temperatureMeasurement": map[string]any{
				"temperature": map[string]any{"value": float64(40)}, // 40C
			},
		}

		result := ExtractGenericApplianceStatus(status)
		if result.State != "running" {
			t.Errorf("State = %q, want %q", result.State, "running")
		}
		if result.RemainingMins == nil || *result.RemainingMins != 30 {
			t.Errorf("RemainingMins = %v, want 30", result.RemainingMins)
		}
		if result.Progress == nil || *result.Progress != 45 {
			t.Errorf("Progress = %v, want 45", result.Progress)
		}
		if result.CompletionTime == nil || *result.CompletionTime != "2025-12-04T15:30:00Z" {
			t.Errorf("CompletionTime = %v, want 2025-12-04T15:30:00Z", result.CompletionTime)
		}
		if result.Temperature == nil {
			t.Error("Temperature should not be nil")
		} else if *result.Temperature != 104 { // 40C = 104F
			t.Errorf("Temperature = %d, want 104", *result.Temperature)
		}
	})

	t.Run("air conditioner with mode", func(t *testing.T) {
		status := Status{
			"samsungce.airConditionerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "running"},
			},
			"airConditionerMode": map[string]any{
				"airConditionerMode": map[string]any{"value": "cooling"},
			},
			"temperatureMeasurement": map[string]any{
				"temperature": map[string]any{"value": float64(25)}, // 25C
			},
			"thermostatCoolingSetpoint": map[string]any{
				"coolingSetpoint": map[string]any{"value": float64(72)}, // Already F
			},
		}

		result := ExtractGenericApplianceStatus(status)
		if result.State != "running" {
			t.Errorf("State = %q, want %q", result.State, "running")
		}
		if result.Mode != "cooling" {
			t.Errorf("Mode = %q, want %q", result.Mode, "cooling")
		}
		if result.Temperature == nil {
			t.Error("Temperature should not be nil")
		}
	})

	t.Run("idle appliance", func(t *testing.T) {
		status := Status{
			"samsungce.dryerOperatingState": map[string]any{
				"machineState": map[string]any{"value": "idle"},
			},
		}

		result := ExtractGenericApplianceStatus(status)
		if result.State != "idle" {
			t.Errorf("State = %q, want %q", result.State, "idle")
		}
	})

	t.Run("empty status returns idle", func(t *testing.T) {
		result := ExtractGenericApplianceStatus(Status{})
		if result.State != "idle" {
			t.Errorf("State = %q, want %q", result.State, "idle")
		}
	})

	t.Run("discovers capabilities", func(t *testing.T) {
		status := Status{
			"switch":             map[string]any{},
			"temperatureMeasurement": map[string]any{},
			"powerMeter":         map[string]any{},
		}

		result := ExtractGenericApplianceStatus(status)
		if len(result.DiscoveredCapabilities) != 3 {
			t.Errorf("DiscoveredCapabilities count = %d, want 3", len(result.DiscoveredCapabilities))
		}
	})

	t.Run("extracts door open status", func(t *testing.T) {
		status := Status{
			"contactSensor": map[string]any{
				"contact": map[string]any{"value": "open"},
			},
		}

		result := ExtractGenericApplianceStatus(status)
		if !result.DoorOpen {
			t.Error("DoorOpen should be true")
		}
	})

	t.Run("extracts power consumption", func(t *testing.T) {
		status := Status{
			"powerMeter": map[string]any{
				"power": map[string]any{"value": float64(150.5)},
			},
		}

		result := ExtractGenericApplianceStatus(status)
		if result.PowerConsumption == nil {
			t.Error("PowerConsumption should not be nil")
		} else if *result.PowerConsumption != 150.5 {
			t.Errorf("PowerConsumption = %f, want 150.5", *result.PowerConsumption)
		}
	})
}

func TestConvertToMinutes(t *testing.T) {
	tests := []struct {
		value float64
		unit  string
		want  int
	}{
		{45, "min", 45},
		{2, "h", 120},
		{90, "s", 2},   // Rounds up
		{120, "", 2},   // Default to seconds
		{-10, "min", 0}, // Negative returns 0
	}

	for _, tt := range tests {
		got := convertToMinutes(tt.value, tt.unit)
		if got != tt.want {
			t.Errorf("convertToMinutes(%f, %q) = %d, want %d", tt.value, tt.unit, got, tt.want)
		}
	}
}
