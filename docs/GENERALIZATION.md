# SmartThings-Go Generalization Roadmap

This document outlines the path to generalizing the smartthings-go library from Samsung-specific implementation to supporting all SmartThings devices.

## Current State: Samsung-Specific

### Hardcoded Samsung Namespaces

The library currently assumes Samsung device capabilities:

**`appliances.go`:**
- `samsungce.dryerOperatingState` (line 27)
- `samsungce.washerOperatingState` (line 32)
- `samsungce.dishwasherOperatingState` (line 37)
- `custom.cooktopOperatingState` (line 168)
- `samsungvd.remoteControl` (TV remote, line 258)
- `custom.picturemode`, `custom.soundmode` (TV settings)

### Samsung-Specific State Values
- "running", "run", "idle", "ready" - Samsung appliance states
- Assumes all appliances report in Celsius
- Hard-coded time/progress reporting format

### Limited Device Type Support
Only supports:
- Samsung washers, dryers, dishwashers
- Samsung ranges (oven + cooktop)
- Samsung refrigerators
- Samsung TVs

**Missing:**
- Generic switches, dimmers, locks
- Thermostats (Ecobee, Nest)
- Sensors (motion, contact, temperature)
- Other TV brands (LG, Sony, Vizio)
- Smart plugs, bulbs (Philips Hue, LIFX)

---

## Target State: Generic SmartThings

### Standard Capability Support

SmartThings defines standard capabilities that work across all brands:

| Capability | Description | Examples |
|------------|-------------|----------|
| `switch` | On/off control | Lights, plugs, switches |
| `switchLevel` | Dimming (0-100) | Dimmable lights, fans |
| `temperatureMeasurement` | Temperature sensor | Thermostats, sensors |
| `thermostat` | Climate control | Ecobee, Nest, Honeywell |
| `lock` | Lock/unlock | Yale, Schlage, August |
| `colorControl` | RGB color | Philips Hue, LIFX |
| `energyMeter` | Energy monitoring | Smart plugs |
| `motionSensor` | Motion detection | Generic sensors |
| `contactSensor` | Open/close | Door/window sensors |

**Goal:** Support these generic capabilities first, with vendor extensions as optional.

---

## Phase 1: Capability-Based Architecture

**Objective:** Refactor library to use standard capabilities instead of appliance types.

**Time Estimate:** 4-6 hours

### Changes Required

**1. New file: `capabilities.go`**

```go
package smartthings

import "fmt"

// CapabilityValue represents a single capability attribute value
type CapabilityValue struct {
    Value     interface{}            `json:"value"`
    Unit      string                 `json:"unit,omitempty"`
    Timestamp string                 `json:"timestamp,omitempty"`
    Data      map[string]interface{} `json:"data,omitempty"`
}

// GetCapability extracts a capability from device status
func GetCapability(status Status, capabilityID string) map[string]CapabilityValue {
    capMap, ok := GetMap(status, capabilityID)
    if !ok {
        return nil
    }

    result := make(map[string]CapabilityValue)
    for attr, data := range capMap {
        if dataMap, ok := data.(map[string]interface{}); ok {
            result[attr] = CapabilityValue{
                Value:     dataMap["value"],
                Unit:      getStringOrEmpty(dataMap, "unit"),
                Timestamp: getStringOrEmpty(dataMap, "timestamp"),
                Data:      dataMap,
            }
        }
    }
    return result
}

// GetSwitch returns switch state (on/off)
func GetSwitch(status Status) (bool, error) {
    if value, ok := GetString(status, "switch", "switch", "value"); ok {
        return value == "on", nil
    }
    return false, fmt.Errorf("switch capability not found")
}

// GetLevel returns dimmer level (0-100)
func GetLevel(status Status) (int, error) {
    if value, ok := GetInt(status, "switchLevel", "level", "value"); ok {
        return value, nil
    }
    return 0, fmt.Errorf("switchLevel capability not found")
}

// GetTemperature returns temperature measurement
func GetTemperature(status Status) (float64, string, error) {
    if value, ok := GetFloat(status, "temperatureMeasurement", "temperature", "value"); ok {
        unit, _ := GetString(status, "temperatureMeasurement", "temperature", "unit")
        if unit == "" {
            unit = "C" // Default to Celsius
        }
        return value, unit, nil
    }
    return 0, "", fmt.Errorf("temperatureMeasurement capability not found")
}

// GetLockState returns lock status
func GetLockState(status Status) (string, error) {
    if value, ok := GetString(status, "lock", "lock", "value"); ok {
        return value, nil // "locked", "unlocked", "unknown"
    }
    return "", fmt.Errorf("lock capability not found")
}

// GetMotion returns motion sensor state
func GetMotion(status Status) (bool, error) {
    if value, ok := GetString(status, "motionSensor", "motion", "value"); ok {
        return value == "active", nil
    }
    return false, fmt.Errorf("motionSensor capability not found")
}
```

**2. Add to `devices.go`:**

```go
// GetDeviceCapabilities returns list of capabilities for a device
func (c *Client) GetDeviceCapabilities(ctx context.Context, deviceID string) ([]string, error) {
    device, err := c.GetDevice(ctx, deviceID)
    if err != nil {
        return nil, err
    }

    var caps []string
    for _, comp := range device.Components {
        for _, cap := range comp.Capabilities {
            caps = append(caps, cap)
        }
    }
    return caps, nil
}
```

### Testing
```go
func TestGenericCapabilities(t *testing.T) {
    client := NewClient(os.Getenv("SMARTTHINGS_TOKEN"))

    // Test with any switch device (not just Samsung)
    status, _ := client.GetDeviceStatus(ctx, "switch-device-id")

    isOn, err := GetSwitch(status)
    assert.NoError(t, err)
    assert.True(t, isOn)

    level, err := GetLevel(status)
    assert.NoError(t, err)
    assert.Equal(t, 75, level)
}
```

---

## Phase 2: Vendor Extensions System

**Objective:** Support vendor-specific capabilities as opt-in extensions.

**Time Estimate:** 3-4 hours

### Directory Structure
```
smartthings-go/
├── capabilities.go          # Generic capabilities
├── devices.go
├── client.go
├── types.go
├── vendors/
│   ├── registry.go          # Vendor registration system
│   ├── samsung/
│   │   ├── appliances.go    # Samsung-specific (move existing code)
│   │   ├── types.go         # Samsung-specific types
│   │   └── tv.go            # Samsung TV extensions
│   ├── philips/
│   │   └── hue.go           # Philips Hue extensions (future)
│   └── generic/
│       └── helpers.go       # Generic helper functions
```

### Vendor Registry Pattern

**New file: `vendors/registry.go`**
```go
package vendors

import "github.com/yourusername/smartthings-go"

type VendorExtension interface {
    Name() string
    SupportsDevice(deviceID string, capabilities []string) bool
    ExtractStatus(status smartthings.Status) interface{}
}

var registry = make(map[string]VendorExtension)

func Register(ext VendorExtension) {
    registry[ext.Name()] = ext
}

func GetExtension(name string) (VendorExtension, bool) {
    ext, ok := registry[name]
    return ext, ok
}

func GetExtensionForDevice(capabilities []string) VendorExtension {
    for _, ext := range registry {
        if ext.SupportsDevice("", capabilities) {
            return ext
        }
    }
    return nil
}
```

**Samsung Extension: `vendors/samsung/samsung.go`**
```go
package samsung

import (
    "strings"
    st "github.com/yourusername/smartthings-go"
    "github.com/yourusername/smartthings-go/vendors"
)

type SamsungExtension struct{}

func (s *SamsungExtension) Name() string {
    return "samsung"
}

func (s *SamsungExtension) SupportsDevice(deviceID string, capabilities []string) bool {
    // Check for Samsung-specific capabilities
    for _, cap := range capabilities {
        if strings.HasPrefix(cap, "samsungce.") || strings.HasPrefix(cap, "custom.") {
            return true
        }
    }
    return false
}

func (s *SamsungExtension) ExtractStatus(status st.Status) interface{} {
    // Use existing ExtractLaundryStatus, ExtractRangeStatus, etc.
    // Determine appliance type and call appropriate extractor
    return nil
}

func init() {
    vendors.Register(&SamsungExtension{})
}
```

---

## Phase 3: Backward Compatibility Layer

**Objective:** Ensure existing API server code continues to work.

**Time Estimate:** 2 hours

### Keep Existing Public API

**Deprecate but don't remove:**
```go
// appliances.go (deprecated but still exported)

// Deprecated: Use vendors/samsung.ExtractLaundryStatus instead
func ExtractLaundryStatus(status Status, applianceType string) *ApplianceStatus {
    return samsung.ExtractLaundryStatus(status, applianceType)
}

// Deprecated: Use vendors/samsung.ExtractRangeStatus instead
func ExtractRangeStatus(status Status) *RangeStatus {
    return samsung.ExtractRangeStatus(status)
}
```

### Add Deprecation Warnings

```go
// types.go

// Deprecated: Samsung-specific type. Use generic capabilities or vendors/samsung instead.
type ApplianceStatus struct {
    State          string
    RemainingMins  *int
    CompletionTime *string
    CycleProgress  *int
}
```

---

## Phase 4: Generic Device Discovery

**Objective:** Auto-detect device types based on capabilities.

**Time Estimate:** 4-6 hours

### Device Classifier

**New file: `classifier.go`**
```go
package smartthings

type DeviceType string

const (
    DeviceTypeSwitch      DeviceType = "switch"
    DeviceTypeDimmer      DeviceType = "dimmer"
    DeviceTypeThermostat  DeviceType = "thermostat"
    DeviceTypeLock        DeviceType = "lock"
    DeviceTypeSensor      DeviceType = "sensor"
    DeviceTypeAppliance   DeviceType = "appliance"
    DeviceTypeUnknown     DeviceType = "unknown"
)

func ClassifyDevice(capabilities []string) DeviceType {
    capSet := make(map[string]bool)
    for _, cap := range capabilities {
        capSet[cap] = true
    }

    // Check for appliances (Samsung-specific)
    if capSet["samsungce.washerOperatingState"] || capSet["samsungce.dryerOperatingState"] {
        return DeviceTypeAppliance
    }

    // Check for standard types
    if capSet["lock"] {
        return DeviceTypeLock
    }
    if capSet["thermostat"] {
        return DeviceTypeThermostat
    }
    if capSet["switchLevel"] {
        return DeviceTypeDimmer
    }
    if capSet["switch"] {
        return DeviceTypeSwitch
    }
    if capSet["temperatureMeasurement"] || capSet["motionSensor"] || capSet["contactSensor"] {
        return DeviceTypeSensor
    }

    return DeviceTypeUnknown
}

func GetDeviceIcon(deviceType DeviceType) string {
    icons := map[DeviceType]string{
        DeviceTypeSwitch:     "toggle",
        DeviceTypeDimmer:     "brightness",
        DeviceTypeThermostat: "thermostat",
        DeviceTypeLock:       "lock",
        DeviceTypeSensor:     "sensor",
        DeviceTypeAppliance:  "appliance",
    }
    return icons[deviceType]
}
```

---

## Phase 5: Standard Command Execution

**Objective:** Support standard commands for any device.

**Time Estimate:** 3-4 hours

**New file: `commands.go`**
```go
package smartthings

import "context"

func (c *Client) TurnOn(ctx context.Context, deviceID string) error {
    return c.ExecuteCommand(ctx, deviceID, "main", "switch", "on", nil)
}

func (c *Client) TurnOff(ctx context.Context, deviceID string) error {
    return c.ExecuteCommand(ctx, deviceID, "main", "switch", "off", nil)
}

func (c *Client) SetLevel(ctx context.Context, deviceID string, level int) error {
    return c.ExecuteCommand(ctx, deviceID, "main", "switchLevel", "setLevel", []interface{}{level})
}

func (c *Client) SetTemperature(ctx context.Context, deviceID string, temp float64) error {
    return c.ExecuteCommand(ctx, deviceID, "main", "thermostatCoolingSetpoint", "setCoolingSetpoint", []interface{}{temp})
}

func (c *Client) Lock(ctx context.Context, deviceID string) error {
    return c.ExecuteCommand(ctx, deviceID, "main", "lock", "lock", nil)
}

func (c *Client) Unlock(ctx context.Context, deviceID string) error {
    return c.ExecuteCommand(ctx, deviceID, "main", "lock", "unlock", nil)
}
```

---

## Phase 6: Testing & Examples

**Objective:** Comprehensive test coverage and example code.

**Time Estimate:** 4-6 hours

### Integration Tests

```go
// integration_test.go
func TestGenericSwitch(t *testing.T) {
    client := NewClient(os.Getenv("SMARTTHINGS_TOKEN"))

    // Test with any switch (not just Samsung)
    devices, _ := client.ListDevices(context.Background())
    var switchDevice string
    for _, d := range devices {
        if ClassifyDevice(d.Capabilities) == DeviceTypeSwitch {
            switchDevice = d.DeviceID
            break
        }
    }

    status, err := client.GetDeviceStatus(context.Background(), switchDevice)
    require.NoError(t, err)

    isOn, err := GetSwitch(status)
    require.NoError(t, err)
    // ...
}
```

### Example Code

**`examples/list_all_devices.go`:**
```go
package main

import (
    "context"
    "fmt"
    st "github.com/yourusername/smartthings-go"
)

func main() {
    client := st.NewClient(os.Getenv("SMARTTHINGS_TOKEN"))
    devices, _ := client.ListDevices(context.Background())

    for _, device := range devices {
        deviceType := st.ClassifyDevice(device.Capabilities)
        fmt.Printf("%s (%s): %v\n", device.Label, deviceType, device.Capabilities)
    }
}
```

---

## Migration Guide for API Server

### Before (current):
```go
import st "github.com/yourusername/smartthings-go"

// Hardcoded appliance extraction
rangeStatus := st.ExtractRangeStatus(st.Status(status))
gasRange.CooktopActive = &rangeStatus.CooktopActive
```

### After (generic + Samsung vendor):
```go
import (
    st "github.com/yourusername/smartthings-go"
    "github.com/yourusername/smartthings-go/vendors/samsung"
)

// Option 1: Use backward-compatible deprecated API (no changes)
rangeStatus := st.ExtractRangeStatus(st.Status(status))  // Still works

// Option 2: Use new vendor extension system
if samsung.IsSamsungDevice(device.Capabilities) {
    rangeStatus := samsung.ExtractRangeStatus(st.Status(status))
    gasRange.CooktopActive = &rangeStatus.CooktopActive
}

// Option 3: Use generic capabilities for non-Samsung
if temp, unit, err := st.GetTemperature(status); err == nil {
    gasRange.OvenTemp = &temp
}
```

---

## Effort Summary

| Phase | Description | Hours |
|-------|-------------|-------|
| 1 | Capability-based architecture | 4-6 |
| 2 | Vendor extensions system | 3-4 |
| 3 | Backward compatibility | 2 |
| 4 | Generic device discovery | 4-6 |
| 5 | Standard command execution | 3-4 |
| 6 | Testing & examples | 4-6 |
| **Total** | | **20-28** |

---

## Breaking Changes (v2.0.0)

If backward compatibility is dropped:
- Move Samsung code to `vendors/samsung` package
- Remove `ExtractLaundryStatus`, `ExtractRangeStatus` from root
- Rename `ApplianceStatus` → `samsung.ApplianceStatus`

**Users must update imports:**
```go
// Old (v1.x)
import st "github.com/yourusername/smartthings-go"
status := st.ExtractRangeStatus(...)

// New (v2.0)
import "github.com/yourusername/smartthings-go/vendors/samsung"
status := samsung.ExtractRangeStatus(...)
```

**Recommended:** Keep v1.x with deprecation warnings for 6 months, then release v2.0.

---

## Benefits of Generalization

1. **Broader Device Support**: Works with Philips Hue, Yale locks, Ecobee thermostats
2. **Community Contributions**: Easier for others to add vendor extensions
3. **Future-Proof**: New Samsung devices work via standard capabilities
4. **Reduced Maintenance**: Less hardcoded logic
5. **Better Testing**: Mock devices using standard capabilities

---

## Non-Goals

- Supporting non-SmartThings platforms (Zigbee, Z-Wave direct)
- Replacing SmartThings cloud API with local control
- Supporting SmartThings Schema connectors

---

## Resources

- [SmartThings Capability Reference](https://smartthings.developer.samsung.com/docs/api-ref/capabilities.html)
- [SmartThings Device Profiles](https://smartthings.developer.samsung.com/docs/devices/device-profiles.html)
- [Go Best Practices for Library Design](https://go.dev/doc/effective_go)
