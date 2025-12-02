# smartthings-go

A comprehensive Go client library for the Samsung SmartThings API.

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

- **Full SmartThings API v1 support**
- **OAuth 2.0 authentication** with automatic token refresh
- **Device management** - list, get, update, delete, execute commands
- **Locations & Rooms** - CRUD operations for organization
- **Scenes** - list and execute scenes
- **Automation Rules** - CRUD operations for rules
- **Schedules** - cron-based scheduling
- **Subscriptions** - webhook management
- **Capabilities** - capability introspection
- **Pagination support** for large datasets
- **Automatic retry** with configurable backoff
- **TV control** (power, volume, input, apps, picture/sound modes)
- **Appliance status** extraction (washer, dryer, dishwasher, range, refrigerator)

## Authentication Methods

### Personal Access Token

Get a token from [SmartThings Tokens](https://account.smartthings.com/tokens):

```go
client, err := st.NewClient("your-personal-access-token")
```

### OAuth 2.0

For apps that need to access other users' devices:

```go
// 1. Configure OAuth
config := &st.OAuthConfig{
    ClientID:     "your-client-id",
    ClientSecret: "your-client-secret",
    RedirectURL:  "https://your-app.com/callback",
    Scopes:       st.DefaultScopes(), // r:devices:*, x:devices:*, r:locations:*
}

// 2. Create token store (file-based or in-memory)
store := st.NewFileTokenStore("/path/to/tokens.json")

// 3. Create OAuth client
client, err := st.NewOAuthClient(config, store)
if err != nil {
    log.Fatal(err)
}

// 4. Check if user needs to authenticate
if client.NeedsReauthentication() {
    state := generateSecureRandomState()
    authURL := client.GetAuthorizationURL(state)
    // Redirect user to authURL...
}

// 5. In your callback handler, exchange the code
func handleCallback(code string) error {
    return client.ExchangeCode(ctx, code)
}

// 6. Use the client - tokens refresh automatically
devices, err := client.ListDevices(ctx)
```

## API Usage

### Client Options

```go
// Basic client
client, err := st.NewClient("your-token")

// With options
client, err := st.NewClient("your-token",
    st.WithTimeout(60 * time.Second),
    st.WithRetry(st.DefaultRetryConfig()),
    st.WithBaseURL("https://custom-api.example.com"),
)
```

### Device Operations

```go
// List all devices
devices, err := client.ListDevices(ctx)

// List with pagination and filtering
paged, err := client.ListDevicesWithOptions(ctx, &st.ListDevicesOptions{
    Capability: []string{"switch"},
    LocationID: []string{"location-id"},
    Max:        100,
    Page:       0,
})

// Get all devices with automatic pagination
allDevices, err := client.ListAllDevices(ctx)

// Get a specific device
device, err := client.GetDevice(ctx, "device-id")

// Get device status (main component)
status, err := client.GetDeviceStatus(ctx, "device-id")

// Get device health
health, err := client.GetDeviceHealth(ctx, "device-id")

// Update device label
updated, err := client.UpdateDevice(ctx, "device-id", &st.DeviceUpdate{Label: "New Name"})

// Delete device
err := client.DeleteDevice(ctx, "device-id")

// Execute a command
err := client.ExecuteCommand(ctx, "device-id", st.NewCommand("switch", "on"))

// Execute multiple commands
err := client.ExecuteCommands(ctx, "device-id", []st.Command{
    st.NewCommand("switch", "on"),
    st.NewCommand("audioVolume", "setVolume", 50),
})
```

### Locations & Rooms

```go
// List locations
locations, err := client.ListLocations(ctx)

// Create a location
loc, err := client.CreateLocation(ctx, &st.LocationCreate{
    Name:        "Home",
    CountryCode: "US",
    TimeZoneID:  "America/New_York",
})

// List rooms in a location
rooms, err := client.ListRooms(ctx, locationID)

// Create a room
room, err := client.CreateRoom(ctx, locationID, &st.RoomCreate{Name: "Living Room"})
```

### Scenes

```go
// List scenes
scenes, err := client.ListScenes(ctx, locationID)

// Execute a scene
err := client.ExecuteScene(ctx, sceneID)
```

### Automation Rules

```go
// List rules
rules, err := client.ListRules(ctx, locationID)

// Create a rule
rule, err := client.CreateRule(ctx, locationID, &st.RuleCreate{
    Name: "Turn on lights at sunset",
    Actions: []st.RuleAction{
        // Rule definition...
    },
})

// Execute a rule manually
err := client.ExecuteRule(ctx, ruleID)
```

### Subscriptions (Webhooks)

```go
// List subscriptions
subs, err := client.ListSubscriptions(ctx, installedAppID)

// Create a device subscription
sub, err := client.CreateSubscription(ctx, installedAppID, &st.SubscriptionCreate{
    SourceType: "DEVICE",
    Device: &st.DeviceSubscription{
        DeviceID:   "device-id",
        Capability: "switch",
        Attribute:  "switch",
    },
})

// Delete all subscriptions
err := client.DeleteAllSubscriptions(ctx, installedAppID)
```

### Capabilities

```go
// List all capabilities
caps, err := client.ListCapabilities(ctx)

// Get capability definition
cap, err := client.GetCapability(ctx, "switch", 1)
```

### TV Control

```go
// Get TV status
status, err := client.FetchTVStatus(ctx, tvDeviceID)
fmt.Printf("Power: %s, Volume: %d\n", status.Power, status.Volume)

// Control TV
client.SetTVPower(ctx, tvDeviceID, true)
client.SetTVVolume(ctx, tvDeviceID, 25)
client.SetTVMute(ctx, tvDeviceID, true)
client.SetTVInput(ctx, tvDeviceID, "HDMI1")

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
```

### Helper Functions

```go
status, _ := client.GetDeviceStatus(ctx, deviceID)

// Extract values from nested paths
power, ok := st.GetString(status, "switch", "switch", "value")
volume, ok := st.GetInt(status, "audioVolume", "volume", "value")
temp, ok := st.GetFloat(status, "temperatureMeasurement", "temperature", "value")

// Check string equality
isOn := st.GetStringEquals(status, "on", "switch", "switch", "value")

// Temperature conversion
fahrenheit := st.CelsiusToFahrenheit(celsius)
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
    } else if st.IsDeviceOffline(err) {
        // Device is offline
    } else if st.IsTimeout(err) {
        // Request timed out
    } else {
        var apiErr *st.APIError
        if errors.As(err, &apiErr) {
            fmt.Printf("API Error %d: %s\n", apiErr.StatusCode, apiErr.Message)
        }
    }
}
```

## Testing

The library provides a `SmartThingsClient` interface for mocking:

```go
type SmartThingsClient interface {
    ListDevices(ctx context.Context) ([]Device, error)
    GetDevice(ctx context.Context, deviceID string) (*Device, error)
    // ... all other methods
}

// Both Client and OAuthClient implement this interface
var _ SmartThingsClient = (*Client)(nil)
var _ SmartThingsClient = (*OAuthClient)(nil)
```

## SmartThings API Reference

- Base URL: `https://api.smartthings.com/v1`
- [API Documentation](https://developer.smartthings.com/docs/api/public/)
- [Getting an API Token](https://account.smartthings.com/tokens)

## License

MIT License - see [LICENSE](LICENSE) for details.
