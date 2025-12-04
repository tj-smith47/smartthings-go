# smartthings-go

[![Go Reference](https://pkg.go.dev/badge/github.com/tj-smith47/smartthings-go.svg)](https://pkg.go.dev/github.com/tj-smith47/smartthings-go)
[![Coverage](https://img.shields.io/endpoint?url=https://raw.githubusercontent.com/tj-smith47/smartthings-go/main/coverage.json)](https://github.com/tj-smith47/smartthings-go/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/tj-smith47/smartthings-go)](https://goreportcard.com/report/github.com/tj-smith47/smartthings-go)

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
- **Response caching** for capabilities and device profiles
- **SSDP discovery** for local SmartThings hubs and Samsung TVs
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

### Token Storage Options

The library provides multiple token storage backends:

```go
// File-based storage (persists across restarts)
store := st.NewFileTokenStore("/path/to/tokens.json")

// In-memory storage (for testing or short-lived processes)
store := st.NewMemoryTokenStore()

// Custom storage (implement TokenStore interface)
type TokenStore interface {
    Load() (*TokenData, error)
    Save(data *TokenData) error
}

// Example: Redis-based storage
type RedisTokenStore struct {
    client *redis.Client
    key    string
}

func (r *RedisTokenStore) Load() (*st.TokenData, error) {
    data, err := r.client.Get(ctx, r.key).Bytes()
    if err != nil {
        return nil, err
    }
    var tokens st.TokenData
    return &tokens, json.Unmarshal(data, &tokens)
}

func (r *RedisTokenStore) Save(data *st.TokenData) error {
    bytes, _ := json.Marshal(data)
    return r.client.Set(ctx, r.key, bytes, 0).Err()
}
```

**Token Refresh Behavior:**
- Tokens are automatically refreshed when expired
- Refresh happens transparently during API calls
- Access tokens expire after ~24 hours
- Refresh tokens are long-lived (~30 days)
- If refresh fails, `NeedsReauthentication()` returns true

## API Usage

### Client Options

```go
// Basic client
client, err := st.NewClient("your-token")

// With options
client, err := st.NewClient("your-token",
    st.WithTimeout(60 * time.Second),
    st.WithRetry(st.DefaultRetryConfig()),
    st.WithCache(st.DefaultCacheConfig()), // Enable response caching
    st.WithBaseURL("https://custom-api.example.com"),
)

// Custom retry configuration
client, err := st.NewClient("your-token",
    st.WithRetry(&st.RetryConfig{
        MaxRetries:     3,
        InitialBackoff: 100 * time.Millisecond,
        MaxBackoff:     5 * time.Second,
        Multiplier:     2.0, // Exponential backoff
    }),
)

// Custom cache configuration
client, err := st.NewClient("your-token",
    st.WithCache(&st.CacheConfig{
        TTL:     15 * time.Minute, // Cache TTL
        MaxSize: 1000,             // Max cached items
    }),
)
```

**Cached Endpoints:**
- Capability definitions (rarely change)
- Device profiles (rarely change)
- Capability presentations (rarely change)

**Not Cached:**
- Device status (changes frequently)
- Device lists (membership changes)
- Commands/actions (side effects)

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

### Webhook Handling

When building SmartApps, you'll receive webhook callbacks from SmartThings. The library provides utilities for validating and processing these callbacks:

```go
// Validate webhook signature (HMAC-SHA256)
valid := st.ValidateWebhookSignature(requestBody, signatureHeader, appSecret)
if !valid {
    return errors.New("invalid webhook signature")
}

// Parse webhook event
var event st.WebhookEvent
if err := json.Unmarshal(requestBody, &event); err != nil {
    return err
}

// Handle based on lifecycle phase
switch event.Lifecycle {
case "PING":
    // Respond with challenge for initial registration
    response := st.PingResponse{PingData: event.PingData}

case "CONFIGURATION":
    // Return app configuration UI

case "INSTALL":
    // App installed - create subscriptions
    for _, device := range event.InstalledApp.Config.Devices {
        client.CreateSubscription(ctx, event.InstalledApp.InstalledAppID, ...)
    }

case "UPDATE":
    // App configuration updated

case "EVENT":
    // Device event received
    for _, evt := range event.Events {
        if evt.EventType == "DEVICE_EVENT" {
            fmt.Printf("Device %s: %s = %v\n",
                evt.DeviceEvent.DeviceID,
                evt.DeviceEvent.Attribute,
                evt.DeviceEvent.Value)
        }
    }

case "UNINSTALL":
    // App uninstalled - clean up
}
```

**Webhook Security:**
- Always validate the `X-ST-SIGNATURE` header using HMAC-SHA256
- Use HTTPS for your webhook endpoint
- Respond within 20 seconds to avoid timeout
- Return 200 OK for successful processing

### Capabilities

```go
// List all capabilities
caps, err := client.ListCapabilities(ctx)

// List capabilities by namespace (standard vs custom)
caps, err := client.ListCapabilitiesWithOptions(ctx, &st.ListCapabilitiesOptions{
    Namespace: st.CapabilityNamespaceSmartThings, // "st" for standard, "custom" for custom
})

// Get capability definition
cap, err := client.GetCapability(ctx, "switch", 1)
```

### Apps & Installed Apps

For SmartApp development:

```go
// List your registered apps
apps, err := client.ListApps(ctx)

// Get app details
app, err := client.GetApp(ctx, appID)

// List installations of your app
installs, err := client.ListInstalledApps(ctx, locationID)

// Get specific installation
install, err := client.GetInstalledApp(ctx, installedAppID)

// Get current (authorized) configuration
config, err := client.GetCurrentInstalledAppConfig(ctx, installedAppID)
if config != nil {
    // Access config values
    for name, entries := range config.Config {
        for _, entry := range entries {
            if entry.DeviceConfig != nil {
                fmt.Printf("Device: %s\n", entry.DeviceConfig.DeviceID)
            }
        }
    }
}

// List all configurations
configs, err := client.ListInstalledAppConfigs(ctx, installedAppID)

// Get specific configuration by ID
config, err := client.GetInstalledAppConfig(ctx, installedAppID, configID)

// Delete installation
err := client.DeleteInstalledApp(ctx, installedAppID)
```

### Schedules

Create scheduled automations:

```go
// List schedules for an installed app
schedules, err := client.ListSchedules(ctx, installedAppID)

// Create a scheduled trigger (cron expression)
schedule, err := client.CreateSchedule(ctx, installedAppID, &st.ScheduleRequest{
    Name: "daily-morning",
    Cron: &st.CronSchedule{
        Expression: "0 7 * * *", // Daily at 7 AM
        Timezone:   "America/New_York",
    },
})

// Delete a schedule
err := client.DeleteSchedule(ctx, installedAppID, scheduleName)
```

### Notifications

Send push notifications to SmartThings app users:

```go
// Send a push notification
resp, err := client.CreateNotification(ctx, &st.NotificationRequest{
    LocationID: locationID,
    Type:       st.NotificationTypeAlert,
    Messages: map[string]st.NotificationMessage{
        "en": {
            Title: "Washer Complete",
            Body:  "Your laundry is done!",
        },
    },
})
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

### Local Network Discovery

```go
// Discover SmartThings hubs on the local network via SSDP
discovery := st.NewDiscovery(5 * time.Second)
hubs, err := discovery.FindHubs(ctx)
for _, hub := range hubs {
    fmt.Printf("Found hub at %s:%d\n", hub.IP, hub.Port)
}

// Discover Samsung TVs
tvs, err := discovery.FindTVs(ctx)
for _, tv := range tvs {
    fmt.Printf("Found TV: %s at %s:%d\n", tv.Name, tv.IP, tv.Port)
}

// Discover both hubs and TVs at once
allDevices, err := discovery.DiscoverAll(ctx)
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

## Concurrency

The library is designed to be safe for concurrent use:

```go
// Client is safe to share across goroutines
client, _ := st.NewClient("your-token")

// Concurrent device polling
var wg sync.WaitGroup
for _, deviceID := range deviceIDs {
    wg.Add(1)
    go func(id string) {
        defer wg.Done()
        status, err := client.GetDeviceStatus(ctx, id)
        // Process status...
    }(deviceID)
}
wg.Wait()
```

**Thread Safety Notes:**
- All client methods are goroutine-safe
- Token refresh is synchronized (one refresh at a time)
- Cache operations are protected by mutex
- Token storage implementations should be thread-safe

## Rate Limiting

SmartThings API has rate limits. The library handles this automatically with configurable retry:

```go
client, _ := st.NewClient("your-token",
    st.WithRetry(&st.RetryConfig{
        MaxRetries:     5,
        InitialBackoff: 1 * time.Second,
    }),
)
```

When rate limited:
- Library retries with exponential backoff
- `IsRateLimited(err)` returns true for rate limit errors
- Consider spreading requests over time for bulk operations

### Enhanced Rate Limit Handling

Get detailed rate limit information:

```go
// RateLimitError includes Retry-After duration
var rle *st.RateLimitError
if errors.As(err, &rle) {
    fmt.Printf("Retry after: %v\n", rle.RetryAfter)
}

// Wait for rate limit to reset
if err := client.WaitForRateLimit(ctx); err != nil {
    return err // Context canceled
}

// Preemptive throttling
if client.ShouldThrottle(10) { // Threshold of 10 remaining
    time.Sleep(time.Second)
}

// Automatic throttling for bulk operations
throttler := st.NewRateLimitThrottler(client, 10, 100*time.Millisecond)
for _, deviceID := range devices {
    throttler.Wait(ctx)
    client.ExecuteCommand(ctx, deviceID, cmd)
}
```

## Batch Operations

Execute commands on multiple devices concurrently:

```go
// Same command to multiple devices
cmd := st.Command{Capability: "switch", Command: "on"}
results := client.ExecuteCommandBatch(ctx, []string{"device1", "device2", "device3"}, cmd, nil)

for _, r := range results {
    if r.Error != nil {
        log.Printf("Device %s failed: %v", r.DeviceID, r.Error)
    }
}

// Different commands to different devices
batch := []st.BatchCommand{
    {DeviceID: "light1", Commands: []st.Command{{Capability: "switch", Command: "on"}}},
    {DeviceID: "light2", Commands: []st.Command{{Capability: "switchLevel", Command: "setLevel", Arguments: []any{50}}}},
}
results := client.ExecuteCommandsBatch(ctx, batch, &st.BatchConfig{
    MaxConcurrent: 5,   // Limit concurrent API calls
    StopOnError:   false, // Continue on errors
})

// Fetch status for multiple devices
statusResults := client.GetDeviceStatusBatch(ctx, deviceIDs, nil)
```

## Structured Logging

Enable structured logging with Go's `log/slog`:

```go
import "log/slog"

// Option 1: Add logger to existing client
logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
client, _ := st.NewClient("token", st.WithLogger(logger))

// Option 2: Create client with request/response logging
logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
    Level: slog.LevelDebug,
}))
client, _ := st.NewLoggingClient("token", logger)

// Manual logging helpers
client.LogDeviceCommand(ctx, deviceID, "switch", "on", nil)
client.LogRateLimit(ctx, st.RateLimitInfo{Remaining: 50})

// Webhook event logging
st.LogWebhookEvent(logger, ctx, event, nil)
```

## Performance

### HTTP/2 Support

The library automatically uses HTTP/2 for connections to the SmartThings API, providing:
- Multiplexed requests over a single connection
- Header compression
- Reduced latency for concurrent requests

### Connection Pooling

Default transport settings (optimized for typical usage):

```go
// These are the default settings - no configuration needed
Transport: &http.Transport{
    MaxIdleConns:        100,  // Total idle connections across all hosts
    MaxIdleConnsPerHost: 10,   // Idle connections per host
    IdleConnTimeout:     90s,  // Close idle connections after 90s
    ForceAttemptHTTP2:   true, // Use HTTP/2 when available
}
```

For high-throughput applications, customize with `WithHTTPClient`:

```go
transport := &http.Transport{
    MaxIdleConns:        200,
    MaxIdleConnsPerHost: 50,  // More connections for parallel requests
    IdleConnTimeout:     120 * time.Second,
    ForceAttemptHTTP2:   true,
}

client, _ := st.NewClient("token",
    st.WithHTTPClient(&http.Client{
        Transport: transport,
        Timeout:   60 * time.Second,
    }),
)
```

### Best Practices

1. **Reuse the client** - Create one client and share it across goroutines
2. **Use iterators for large lists** - Automatic pagination without loading all items
3. **Enable caching** - Reduces API calls for capabilities and device profiles
4. **Batch commands** - Use `ExecuteCommands` for multiple commands to one device

## SmartThings API Reference

- Base URL: `https://api.smartthings.com/v1`
- [API Documentation](https://developer.smartthings.com/docs/api/public/)
- [Getting an API Token](https://account.smartthings.com/tokens)
- [Developer Workspace](https://developer.smartthings.com/) - Create SmartApps and device integrations
- [OAuth App Registration](https://developer.smartthings.com/workspace) - Register OAuth apps

## Disclaimer

Generated entirely by Claude Opus 4.5 over many iterations 🤖

## License

MIT License - see [LICENSE](LICENSE) for details.
