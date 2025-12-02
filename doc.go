// Package smartthings provides a Go client library for the Samsung SmartThings API.
//
// This library provides comprehensive access to the SmartThings public API, including
// device management, location/room organization, scene execution, automation rules,
// and webhook subscriptions.
//
// # Authentication
//
// The library supports two authentication methods:
//
// Personal Access Token (PAT) - Simple, good for personal projects:
//
//	client, err := smartthings.NewClient("your-personal-access-token")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
// OAuth 2.0 - Required for published apps, supports user authorization:
//
//	config := &smartthings.OAuthConfig{
//	    ClientID:     "your-client-id",
//	    ClientSecret: "your-client-secret",
//	    RedirectURL:  "https://your-app.com/callback",
//	    Scopes:       smartthings.DefaultScopes(),
//	}
//	store := smartthings.NewFileTokenStore("/path/to/tokens.json")
//	client, err := smartthings.NewOAuthClient(config, store)
//
// # Basic Usage
//
// List all devices:
//
//	devices, err := client.ListDevices(ctx)
//	for _, device := range devices {
//	    fmt.Printf("Device: %s (%s)\n", device.Label, device.DeviceID)
//	}
//
// Get device status:
//
//	status, err := client.GetDeviceStatus(ctx, deviceID)
//	power, ok := smartthings.GetString(status, "switch", "switch", "value")
//
// Execute a command:
//
//	err := client.ExecuteCommand(ctx, deviceID, smartthings.NewCommand("switch", "on"))
//
// # Pagination
//
// For large device lists, use pagination:
//
//	paged, err := client.ListDevicesWithOptions(ctx, &smartthings.ListDevicesOptions{
//	    Max:  100,
//	    Page: 0,
//	    Capability: []string{"switch"},
//	})
//	fmt.Printf("Got %d of %d devices\n", len(paged.Items), paged.PageInfo.TotalResults)
//
// Or automatically fetch all pages:
//
//	allDevices, err := client.ListAllDevices(ctx)
//
// # Retry Configuration
//
// Enable automatic retry for transient failures:
//
//	client, err := smartthings.NewClient("token",
//	    smartthings.WithRetry(smartthings.DefaultRetryConfig()),
//	)
//
// # Error Handling
//
// Check for specific error types:
//
//	status, err := client.GetDeviceStatus(ctx, deviceID)
//	if err != nil {
//	    if smartthings.IsUnauthorized(err) {
//	        // Token is invalid or expired
//	    } else if smartthings.IsNotFound(err) {
//	        // Device doesn't exist
//	    } else if smartthings.IsRateLimited(err) {
//	        // Too many requests
//	    }
//	}
//
// # API Coverage
//
// The library supports the following SmartThings API endpoints:
//
//   - Devices: List, get, update, delete, execute commands, health status
//   - Locations: CRUD operations for locations
//   - Rooms: CRUD operations for rooms within locations
//   - Scenes: List and execute scenes
//   - Rules: CRUD operations for automation rules
//   - Schedules: CRUD operations for cron-based schedules
//   - Subscriptions: Webhook subscription management
//   - Installed Apps: List, get, delete installed apps
//   - Capabilities: List and get capability definitions
//
// # Samsung-Specific Features
//
// The library includes specialized support for Samsung devices:
//
//   - TV control (power, volume, input, apps, picture/sound modes)
//   - Appliance status extraction (washer, dryer, dishwasher, range, refrigerator)
//   - Brilliant switch/dimmer support
//
// For more information, see https://developer.smartthings.com/docs/api/public/
package smartthings
