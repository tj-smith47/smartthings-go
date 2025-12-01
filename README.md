# smartthings-go

A Go client library for the Samsung SmartThings API.

## Installation

```bash
go get github.com/tj-smith47/smartthings-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    st "github.com/tj-smith47/smartthings-go"
)

func main() {
    // Create a client with your SmartThings API token
    client, err := st.NewClient("your-api-token")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    // List all devices
    devices, err := client.ListDevices(ctx)
    if err != nil {
        log.Fatal(err)
    }

    for _, device := range devices {
        fmt.Printf("Device: %s (%s)\n", device.Label, device.DeviceID)
    }
}
```

## Features

- Full SmartThings API v1 support
- Device listing and status retrieval
- Command execution
- TV control (power, volume, input, apps, picture/sound modes)
- Appliance status extraction (washer, dryer, dishwasher, range, refrigerator)
- Helper functions for navigating nested JSON responses
- Proper error handling with typed errors

## API Usage

### Client Creation

```go
// Basic client
client, err := st.NewClient("your-token")
if err != nil {
    log.Fatal(err)
}

// With options
client, err := st.NewClient("your-token",
    st.WithTimeout(60 * time.Second),
    st.WithBaseURL("https://custom-api.example.com"),
)
if err != nil {
    log.Fatal(err)
}
```

### Device Operations

```go
ctx := context.Background()

// List all devices
devices, err := client.ListDevices(ctx)

// Get a specific device
device, err := client.GetDevice(ctx, "device-id")

// Get device status (main component)
status, err := client.GetDeviceStatus(ctx, "device-id")

// Get full status (all components)
allStatus, err := client.GetDeviceStatusAllComponents(ctx, "device-id")

// Execute a command
err := client.ExecuteCommand(ctx, "device-id", st.NewCommand("switch", "on"))

// Execute multiple commands
err := client.ExecuteCommands(ctx, "device-id", []st.Command{
    st.NewCommand("switch", "on"),
    st.NewCommand("audioVolume", "setVolume", 50),
})
```

### TV Control

```go
// Get TV status
status, err := client.FetchTVStatus(ctx, tvDeviceID)
fmt.Printf("Power: %s, Volume: %d, Input: %s\n",
    status.Power, status.Volume, status.InputSource)

// Control TV
client.SetTVPower(ctx, tvDeviceID, true)       // Turn on
client.SetTVVolume(ctx, tvDeviceID, 25)        // Set volume
client.SetTVMute(ctx, tvDeviceID, true)        // Mute
client.SetTVInput(ctx, tvDeviceID, "HDMI1")    // Change input

// Remote control
client.SendTVKey(ctx, tvDeviceID, "UP")
client.SendTVKey(ctx, tvDeviceID, "ENTER")

// Media control
client.TVPlay(ctx, tvDeviceID)
client.TVPause(ctx, tvDeviceID)

// Apps
client.LaunchTVApp(ctx, tvDeviceID, "Netflix")

// Picture/Sound modes
client.SetPictureMode(ctx, tvDeviceID, "Movie")
client.SetSoundMode(ctx, tvDeviceID, "Standard")
```

### Appliance Status

```go
// Get washer/dryer/dishwasher status
status, _ := client.GetDeviceStatus(ctx, washerDeviceID)
laundryStatus := st.ExtractLaundryStatus(status, st.ApplianceWasher)
if laundryStatus != nil && laundryStatus.State == "running" {
    fmt.Printf("Washing: %d mins remaining\n", *laundryStatus.RemainingMins)
}

// Get range status
status, _ := client.GetDeviceStatus(ctx, rangeDeviceID)
rangeStatus := st.ExtractRangeStatus(status)
if rangeStatus.OvenActive {
    fmt.Printf("Oven: %d°F (target: %d°F)\n",
        *rangeStatus.OvenTemp, *rangeStatus.OvenTargetTemp)
}

// Get refrigerator status (requires full component status)
allStatus, _ := client.GetDeviceStatusAllComponents(ctx, fridgeDeviceID)
fridgeStatus := st.ExtractRefrigeratorStatus(allStatus)
fmt.Printf("Fridge: %d°F, Freezer: %d°F\n",
    *fridgeStatus.FridgeTemp, *fridgeStatus.FreezerTemp)
```

### Helper Functions

The library provides helper functions for navigating deeply nested JSON responses:

```go
status, _ := client.GetDeviceStatus(ctx, deviceID)

// Extract values from nested paths
power, ok := st.GetString(status, "switch", "switch", "value")
volume, ok := st.GetInt(status, "audioVolume", "volume", "value")
temp, ok := st.GetFloat(status, "temperatureMeasurement", "temperature", "value")
muted, ok := st.GetBool(status, "audioMute", "mute", "value")

// Navigate to nested maps
main, ok := st.GetMap(status, "main")

// Extract arrays
inputs, ok := st.GetArray(status, "mediaInputSource", "supportedInputSources", "value")

// Check string equality
isOn := st.GetStringEquals(status, "on", "switch", "switch", "value")

// Temperature conversion
fahrenheit := st.CelsiusToFahrenheit(celsius)
celsius := st.FahrenheitToCelsius(fahrenheit)
```

## Error Handling

```go
status, err := client.GetDeviceStatus(ctx, deviceID)
if err != nil {
    if st.IsUnauthorized(err) {
        // Token is invalid or expired
    } else if st.IsNotFound(err) {
        // Device doesn't exist
    } else if st.IsRateLimited(err) {
        // Too many requests
    } else {
        // Other API error
        var apiErr *st.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API Error %d: %s\n", apiErr.StatusCode, apiErr.Message)
        }
    }
}
```

## SmartThings API Reference

- Base URL: `https://api.smartthings.com/v1`
- [API Documentation](https://developer-preview.smartthings.com/docs/api/public/)
- [Getting an API Token](https://account.smartthings.com/tokens)

## License

MIT License - see [LICENSE](LICENSE) for details.
