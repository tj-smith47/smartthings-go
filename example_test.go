package smartthings_test

import (
	"context"
	"fmt"
	"log"
	"time"

	st "github.com/tj-smith47/smartthings-go"
)

func ExampleNewClient() {
	// Create a client with a Personal Access Token
	client, err := st.NewClient("your-api-token")
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	devices, err := client.ListDevices(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, device := range devices {
		fmt.Printf("Device: %s\n", device.Label)
	}
}

func ExampleNewClient_withOptions() {
	// Create a client with custom options
	client, err := st.NewClient("your-api-token",
		st.WithTimeout(30*time.Second),
		st.WithRetry(&st.RetryConfig{
			MaxRetries:     3,
			InitialBackoff: 100 * time.Millisecond,
			MaxBackoff:     5 * time.Second,
		}),
		st.WithCache(st.DefaultCacheConfig()),
	)
	if err != nil {
		log.Fatal(err)
	}

	_ = client
}

func ExampleClient_Devices() {
	client, _ := st.NewClient("your-api-token")
	ctx := context.Background()

	// Iterate over all devices using the iterator pattern
	for device, err := range client.Devices(ctx) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Device: %s (%s)\n", device.Label, device.DeviceID)
	}
}

func ExampleClient_DevicesWithOptions() {
	client, _ := st.NewClient("your-api-token")
	ctx := context.Background()

	// Filter devices by capability and location
	opts := &st.ListDevicesOptions{
		Capability: []string{"switch"},
		LocationID: []string{"your-location-id"},
	}

	for device, err := range client.DevicesWithOptions(ctx, opts) {
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Switch: %s\n", device.Label)
	}
}

func ExampleClient_ExecuteCommands() {
	client, _ := st.NewClient("your-api-token")
	ctx := context.Background()

	// Turn on a light
	err := client.ExecuteCommands(ctx, "device-id", []st.Command{
		{
			Component:  "main",
			Capability: "switch",
			Command:    "on",
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_ExecuteCommands_dimmer() {
	client, _ := st.NewClient("your-api-token")
	ctx := context.Background()

	// Set dimmer level to 50%
	err := client.ExecuteCommands(ctx, "device-id", []st.Command{
		{
			Component:  "main",
			Capability: "switchLevel",
			Command:    "setLevel",
			Arguments:  []any{50},
		},
	})
	if err != nil {
		log.Fatal(err)
	}
}

func ExampleClient_GetDeviceFullStatus() {
	client, _ := st.NewClient("your-api-token")
	ctx := context.Background()

	// Returns map[componentID]Status
	components, err := client.GetDeviceFullStatus(ctx, "device-id")
	if err != nil {
		log.Fatal(err)
	}

	// Access the "main" component status
	if main, ok := components["main"]; ok {
		if sw, ok := st.GetMap(main, "switch"); ok {
			if attr, ok := st.GetMap(sw, "switch"); ok {
				if value, ok := st.GetString(attr, "value"); ok {
					fmt.Printf("Switch is: %s\n", value)
				}
			}
		}
	}
}

func ExampleClient_ListScenes() {
	client, _ := st.NewClient("your-api-token")
	ctx := context.Background()

	scenes, err := client.ListScenes(ctx, "location-id")
	if err != nil {
		log.Fatal(err)
	}

	for _, scene := range scenes {
		fmt.Printf("Scene: %s\n", scene.SceneName)
	}
}

func ExampleClient_ExecuteScene() {
	client, _ := st.NewClient("your-api-token")
	ctx := context.Background()

	err := client.ExecuteScene(ctx, "scene-id")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Scene executed")
}

func ExampleValidateWebhookSignature() {
	secret := "your-webhook-secret"
	body := []byte(`{"lifecycle":"EVENT","eventData":{}}`)
	signature := "base64-hmac-signature"

	if st.ValidateWebhookSignature(secret, body, signature) {
		fmt.Println("Valid webhook signature")
	} else {
		fmt.Println("Invalid webhook signature")
	}
}

func ExampleExtractLaundryStatus() {
	// Given a device status from GetDeviceFullStatus
	status := st.Status{
		"samsungce.washerOperatingState": map[string]any{
			"machineState": map[string]any{"value": "running"},
			"remainingTime": map[string]any{
				"value": float64(45),
				"unit":  "min",
			},
			"progress": map[string]any{"value": float64(60)},
		},
	}

	laundry := st.ExtractLaundryStatus(status, st.ApplianceWasher)
	if laundry != nil {
		fmt.Printf("State: %s, Remaining: %d mins\n", laundry.State, *laundry.RemainingMins)
	}
	// Output: State: running, Remaining: 45 mins
}

func ExampleDiscovery_FindHubs() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discovery := st.NewDiscovery(3 * time.Second)
	hubs, err := discovery.FindHubs(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, hub := range hubs {
		fmt.Printf("Hub found at %s\n", hub.IP)
	}
}

func ExampleDiscovery_FindTVs() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	discovery := st.NewDiscovery(3 * time.Second)
	tvs, err := discovery.FindTVs(ctx)
	if err != nil {
		log.Fatal(err)
	}

	for _, tv := range tvs {
		fmt.Printf("TV found at %s: %s\n", tv.IP, tv.Name)
	}
}

func ExampleWithCache() {
	// Enable caching for capability and device profile lookups
	client, _ := st.NewClient("your-api-token",
		st.WithCache(&st.CacheConfig{
			CapabilityTTL:    1 * time.Hour,
			DeviceProfileTTL: 30 * time.Minute,
		}),
	)

	ctx := context.Background()

	// First call fetches from API
	cap1, _ := client.GetCapability(ctx, "switch", 1)

	// Second call returns cached result
	cap2, _ := client.GetCapability(ctx, "switch", 1)

	// Both return the same data
	fmt.Printf("Cached: %v\n", cap1.ID == cap2.ID)
}

func ExampleWithRateLimitCallback() {
	// Get notified when rate limits are encountered
	client, _ := st.NewClient("your-api-token",
		st.WithRateLimitCallback(func(info st.RateLimitInfo) {
			fmt.Printf("Rate limit: %d remaining, resets at %s\n", info.Remaining, info.Reset)
		}),
	)

	_ = client
}
