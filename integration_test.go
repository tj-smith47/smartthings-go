//go:build integration

package smartthings

import (
	"context"
	"os"
	"testing"
	"time"
)

// Integration tests require a valid SmartThings API token.
// Run with: go test -tags=integration -v
//
// Environment variables:
//   SMARTTHINGS_TOKEN - Personal Access Token (required)
//   SMARTTHINGS_DEVICE_ID - Device ID for command tests (optional)
//   SMARTTHINGS_LOCATION_ID - Location ID for location tests (optional)

func getTestToken(t *testing.T) string {
	token := os.Getenv("SMARTTHINGS_TOKEN")
	if token == "" {
		t.Skip("SMARTTHINGS_TOKEN not set, skipping integration test")
	}
	return token
}

func TestIntegration_ListDevices(t *testing.T) {
	token := getTestToken(t)
	client, err := NewClient(token)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	devices, err := client.ListDevices(ctx)
	if err != nil {
		t.Fatalf("ListDevices: %v", err)
	}

	t.Logf("Found %d devices", len(devices))
	for _, d := range devices {
		t.Logf("  - %s (%s): %s", d.Label, d.DeviceID, d.DeviceTypeName)
	}
}

func TestIntegration_ListLocations(t *testing.T) {
	token := getTestToken(t)
	client, err := NewClient(token)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	locations, err := client.ListLocations(ctx)
	if err != nil {
		t.Fatalf("ListLocations: %v", err)
	}

	t.Logf("Found %d locations", len(locations))
	for _, loc := range locations {
		t.Logf("  - %s (%s)", loc.Name, loc.LocationID)
	}
}

func TestIntegration_DevicesIterator(t *testing.T) {
	token := getTestToken(t)
	client, err := NewClient(token)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	count := 0
	for device, err := range client.Devices(ctx) {
		if err != nil {
			t.Fatalf("Devices iterator error: %v", err)
		}
		count++
		t.Logf("Device %d: %s", count, device.Label)
	}
	t.Logf("Iterated over %d devices", count)
}

func TestIntegration_GetDeviceStatus(t *testing.T) {
	token := getTestToken(t)
	deviceID := os.Getenv("SMARTTHINGS_DEVICE_ID")
	if deviceID == "" {
		t.Skip("SMARTTHINGS_DEVICE_ID not set")
	}

	client, err := NewClient(token)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	status, err := client.GetDeviceFullStatus(ctx, deviceID)
	if err != nil {
		t.Fatalf("GetDeviceFullStatus: %v", err)
	}

	t.Logf("Device status has %d components", len(status.Components))
}

func TestIntegration_Capabilities(t *testing.T) {
	token := getTestToken(t)
	client, err := NewClient(token)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get a common capability
	cap, err := client.GetCapability(ctx, "switch", 1)
	if err != nil {
		t.Fatalf("GetCapability: %v", err)
	}

	t.Logf("Capability: %s v%d", cap.ID, cap.Version)
	t.Logf("  Commands: %d", len(cap.Commands))
	t.Logf("  Attributes: %d", len(cap.Attributes))
}

func TestIntegration_Scenes(t *testing.T) {
	token := getTestToken(t)
	locationID := os.Getenv("SMARTTHINGS_LOCATION_ID")
	if locationID == "" {
		t.Skip("SMARTTHINGS_LOCATION_ID not set")
	}

	client, err := NewClient(token)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	scenes, err := client.ListScenes(ctx, locationID)
	if err != nil {
		t.Fatalf("ListScenes: %v", err)
	}

	t.Logf("Found %d scenes", len(scenes))
	for _, s := range scenes {
		t.Logf("  - %s (%s)", s.SceneName, s.SceneID)
	}
}

func TestIntegration_Retry(t *testing.T) {
	token := getTestToken(t)
	client, err := NewClient(token, WithRetry(&RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     1 * time.Second,
	}))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// This should work normally, retry only kicks in on transient errors
	_, err = client.ListDevices(ctx)
	if err != nil {
		t.Fatalf("ListDevices with retry: %v", err)
	}
}
