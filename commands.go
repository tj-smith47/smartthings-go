package smartthings

// Command builders for common appliance control patterns.
// These simplify creating SmartThings commands with proper structure.

// NewPowerCommand creates a switch on/off command.
//
// Example:
//
//	cmd := NewPowerCommand(true) // Turn on
//	client.ExecuteCommand(ctx, deviceID, cmd)
func NewPowerCommand(on bool) Command {
	cmd := "off"
	if on {
		cmd = "on"
	}
	return Command{
		Capability: "switch",
		Command:    cmd,
	}
}

// NewCycleCommand creates a cycle selection command for laundry appliances.
// The capability should be "samsungce.washerCycle", "samsungce.dryerCycle", etc.
//
// Example:
//
//	cmd := NewCycleCommand("samsungce.washerCycle", "normal")
func NewCycleCommand(capability, cycle string) Command {
	return Command{
		Capability: capability,
		Command:    "setCycle",
		Arguments:  []any{cycle},
	}
}

// NewTemperatureCommand creates a temperature setpoint command.
// Useful for ovens, thermostats, and temperature-controlled appliances.
//
// Example:
//
//	cmd := NewTemperatureCommand("ovenSetpoint", "setOvenSetpoint", 350)
func NewTemperatureCommand(capability, command string, temp int) Command {
	return Command{
		Capability: capability,
		Command:    command,
		Arguments:  []any{temp},
	}
}

// NewModeCommand creates a mode selection command.
// Works with oven modes, AC modes, wash modes, etc.
//
// Example:
//
//	cmd := NewModeCommand("ovenMode", "setOvenMode", "Bake")
func NewModeCommand(capability, command, mode string) Command {
	return Command{
		Capability: capability,
		Command:    command,
		Arguments:  []any{mode},
	}
}

// NewChildLockCommand creates a child lock on/off command.
// Works with Samsung CE appliances that support samsungce.kidsLock.
//
// Example:
//
//	cmd := NewChildLockCommand(true) // Enable child lock
func NewChildLockCommand(enabled bool) Command {
	state := "off"
	if enabled {
		state = "on"
	}
	return Command{
		Capability: "samsungce.kidsLock",
		Command:    "setKidsLock",
		Arguments:  []any{state},
	}
}

// NewOperationCommand creates a start/pause/stop operation command.
// The capability should be "samsungce.washerOperatingState",
// "samsungce.dryerOperatingState", "samsungce.dishwasherOperation", etc.
//
// Example:
//
//	cmd := NewOperationCommand("samsungce.washerOperatingState", "start")
//	cmd := NewOperationCommand("samsungce.washerOperatingState", "pause")
//	cmd := NewOperationCommand("samsungce.washerOperatingState", "stop")
func NewOperationCommand(capability, operation string) Command {
	return Command{
		Capability: capability,
		Command:    operation,
	}
}

// NewLevelCommand creates a level/value selection command.
// Works with spin level, soil level, water temperature, etc.
//
// Example:
//
//	cmd := NewLevelCommand("custom.washerSpinLevel", "setWasherSpinLevel", "high")
func NewLevelCommand(capability, command, value string) Command {
	return Command{
		Capability: capability,
		Command:    command,
		Arguments:  []any{value},
	}
}

// NewToggleCommand creates a generic toggle command with enabled/disabled state.
// Works with features like power cool, power freeze, vacation mode, etc.
//
// Example:
//
//	cmd := NewToggleCommand("samsungce.powerCool", "setPowerCool", true)
func NewToggleCommand(capability, command string, enabled bool) Command {
	state := "off"
	if enabled {
		state = "on"
	}
	return Command{
		Capability: capability,
		Command:    command,
		Arguments:  []any{state},
	}
}

// NewLampCommand creates a lamp/light on/off command for appliances.
// Works with oven lights, refrigerator lights, etc.
//
// Example:
//
//	cmd := NewLampCommand(true) // Turn on oven light
func NewLampCommand(on bool) Command {
	state := "off"
	if on {
		state = "on"
	}
	return Command{
		Capability: "samsungce.lamp",
		Command:    "setLampState",
		Arguments:  []any{state},
	}
}

// NewRefreshCommand creates a refresh command to force device status update.
//
// Example:
//
//	cmd := NewRefreshCommand()
func NewRefreshCommand() Command {
	return Command{
		Capability: "refresh",
		Command:    "refresh",
	}
}
