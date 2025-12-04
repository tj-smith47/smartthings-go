package smartthings

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// BenchmarkJSONUnmarshalDevice benchmarks JSON unmarshaling of device responses.
func BenchmarkJSONUnmarshalDevice(b *testing.B) {
	deviceJSON := []byte(`{
		"deviceId": "device-123",
		"name": "Test Device",
		"label": "Living Room Light",
		"manufacturerName": "Samsung",
		"presentationId": "pres-123",
		"deviceManufacturerCode": "samsung",
		"locationId": "loc-123",
		"ownerId": "owner-123",
		"roomId": "room-123",
		"deviceTypeName": "Light",
		"components": [
			{
				"id": "main",
				"label": "Main",
				"capabilities": [
					{"id": "switch", "version": 1},
					{"id": "switchLevel", "version": 1}
				]
			}
		],
		"createTime": "2024-01-01T00:00:00Z",
		"profile": {"id": "profile-123"}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var device Device
		if err := json.Unmarshal(deviceJSON, &device); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkJSONUnmarshalDeviceList benchmarks JSON unmarshaling of paginated device lists.
func BenchmarkJSONUnmarshalDeviceList(b *testing.B) {
	listJSON := []byte(`{
		"items": [
			{"deviceId": "device-1", "label": "Device 1"},
			{"deviceId": "device-2", "label": "Device 2"},
			{"deviceId": "device-3", "label": "Device 3"},
			{"deviceId": "device-4", "label": "Device 4"},
			{"deviceId": "device-5", "label": "Device 5"}
		],
		"_links": {"next": "/devices?page=1"}
	}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var resp PagedDevices
		if err := json.Unmarshal(listJSON, &resp); err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkCacheHit benchmarks cache lookup performance.
func BenchmarkCacheHit(b *testing.B) {
	cache := NewMemoryCache()

	// Pre-populate cache
	cap := &Capability{ID: "switch", Version: 1}
	cache.Set("switch:1", cap, 5*time.Minute)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("switch:1")
	}
}

// BenchmarkCacheMiss benchmarks cache miss performance.
func BenchmarkCacheMiss(b *testing.B) {
	cache := NewMemoryCache()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = cache.Get("nonexistent")
	}
}

// BenchmarkWebhookValidation benchmarks HMAC validation.
func BenchmarkWebhookValidation(b *testing.B) {
	secret := "test-secret-key"
	body := []byte(`{"lifecycle":"EVENT","eventData":{"events":[{"eventType":"DEVICE_EVENT"}]}}`)

	// Pre-compute a valid signature (base64 of HMAC-SHA256)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	signature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ValidateWebhookSignature(secret, body, signature)
	}
}

// BenchmarkClientRequest benchmarks a simple API request.
func BenchmarkClientRequest(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]Device{
			{DeviceID: "device-1", Label: "Test"},
		})
	}))
	defer server.Close()

	client, _ := NewClient("test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := client.ListDevices(ctx)
		if err != nil {
			b.Fatal(err)
		}
	}
}

// BenchmarkDevicesIterator benchmarks the iterator pattern.
func BenchmarkDevicesIterator(b *testing.B) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := PagedDevices{
			Items: []Device{
				{DeviceID: "device-1"},
				{DeviceID: "device-2"},
				{DeviceID: "device-3"},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client, _ := NewClient("test-token", WithBaseURL(server.URL))
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, err := range client.Devices(ctx) {
			if err != nil {
				b.Fatal(err)
			}
		}
	}
}

// BenchmarkStatusParsing benchmarks device status parsing.
func BenchmarkStatusParsing(b *testing.B) {
	status := Status{
		"switch": map[string]any{
			"switch": map[string]any{
				"value":     "on",
				"timestamp": "2024-01-01T00:00:00Z",
			},
		},
		"switchLevel": map[string]any{
			"level": map[string]any{
				"value":     float64(75),
				"unit":      "%",
				"timestamp": "2024-01-01T00:00:00Z",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate status access patterns
		if m, ok := GetMap(status, "switch"); ok {
			if sw, ok := GetMap(m, "switch"); ok {
				_, _ = GetString(sw, "value")
			}
		}
		if m, ok := GetMap(status, "switchLevel"); ok {
			if lvl, ok := GetMap(m, "level"); ok {
				_, _ = GetFloat(lvl, "value")
			}
		}
	}
}

// BenchmarkApplianceExtraction benchmarks appliance status extraction.
func BenchmarkApplianceExtraction(b *testing.B) {
	status := Status{
		"samsungce.washerOperatingState": map[string]any{
			"machineState": map[string]any{"value": "running"},
			"remainingTime": map[string]any{
				"value": float64(45),
				"unit":  "min",
			},
			"progress": map[string]any{"value": float64(60)},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = ExtractLaundryStatus(status, ApplianceWasher)
	}
}
