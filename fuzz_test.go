package smartthings

import (
	"encoding/json"
	"testing"
)

// FuzzWebhookSignature fuzzes the webhook signature validation.
// Run with: go test -fuzz=FuzzWebhookSignature
func FuzzWebhookSignature(f *testing.F) {
	// Add seed corpus
	f.Add("secret123", []byte(`{"lifecycle":"EVENT"}`), "abc123signature")
	f.Add("", []byte(`{}`), "")
	f.Add("key", []byte(`null`), "sig")
	f.Add("test-secret", []byte(`{"nested": {"deep": {"value": 1}}}`), "x")

	f.Fuzz(func(t *testing.T, secret string, body []byte, signature string) {
		// Should not panic
		_ = ValidateWebhookSignature(secret, body, signature)
	})
}

// FuzzDeviceJSONParsing fuzzes device JSON unmarshaling.
// Run with: go test -fuzz=FuzzDeviceJSONParsing
func FuzzDeviceJSONParsing(f *testing.F) {
	f.Add([]byte(`{"deviceId":"123","label":"Test"}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"components":[]}`))
	f.Add([]byte(`{"deviceId":"","components":[{"id":"main","capabilities":[]}]}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var device Device
		// Should not panic - errors are acceptable
		_ = json.Unmarshal(data, &device)
	})
}

// FuzzStatusParsing fuzzes status map parsing.
// Run with: go test -fuzz=FuzzStatusParsing
func FuzzStatusParsing(f *testing.F) {
	f.Add([]byte(`{"switch":{"switch":{"value":"on"}}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"cap":{"attr":{"value":123}}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var status Status
		if err := json.Unmarshal(data, &status); err != nil {
			return // Invalid JSON is acceptable
		}

		// Exercise the helper functions - should not panic
		for capName := range status {
			if m, ok := GetMap(status, capName); ok {
				for attrName := range m {
					if attr, ok := GetMap(m, attrName); ok {
						_, _ = GetString(attr, "value")
						_, _ = GetFloat(attr, "value")
						_, _ = GetBool(attr, "value")
					}
				}
			}
		}
	})
}

// FuzzWebhookEventParsing fuzzes webhook event JSON parsing.
// Run with: go test -fuzz=FuzzWebhookEventParsing
func FuzzWebhookEventParsing(f *testing.F) {
	f.Add([]byte(`{"lifecycle":"EVENT","eventData":{"events":[]}}`))
	f.Add([]byte(`{"lifecycle":"CONFIRMATION","confirmationData":{"confirmationUrl":"https://example.com"}}`))
	f.Add([]byte(`{}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var event WebhookEvent
		// Should not panic
		_ = json.Unmarshal(data, &event)
	})
}

// FuzzCapabilityParsing fuzzes capability JSON parsing.
// Run with: go test -fuzz=FuzzCapabilityParsing
func FuzzCapabilityParsing(f *testing.F) {
	f.Add([]byte(`{"id":"switch","version":1,"commands":{},"attributes":{}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"id":"","version":0}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var cap Capability
		// Should not panic
		_ = json.Unmarshal(data, &cap)
	})
}

// FuzzCommandRequestParsing fuzzes command request building.
// Run with: go test -fuzz=FuzzCommandRequestParsing
func FuzzCommandRequestParsing(f *testing.F) {
	f.Add("switch", "on", "main")
	f.Add("switchLevel", "setLevel", "main")
	f.Add("", "", "")

	f.Fuzz(func(t *testing.T, capability, command, component string) {
		// Should not panic
		req := CommandRequest{
			Commands: []Command{
				{
					Capability: capability,
					Command:    command,
					Component:  component,
				},
			},
		}
		_, _ = json.Marshal(req)
	})
}

// FuzzApplianceStatusExtraction fuzzes appliance status extraction.
// Run with: go test -fuzz=FuzzApplianceStatusExtraction
func FuzzApplianceStatusExtraction(f *testing.F) {
	f.Add([]byte(`{"samsungce.washerOperatingState":{"machineState":{"value":"running"}}}`))
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"samsungce.dryerOperatingState":{"remainingTime":{"value":45,"unit":"min"}}}`))

	f.Fuzz(func(t *testing.T, data []byte) {
		var status Status
		if err := json.Unmarshal(data, &status); err != nil {
			return
		}

		// Should not panic
		_ = ExtractLaundryStatus(status, ApplianceWasher)
		_ = ExtractLaundryStatus(status, ApplianceDryer)
		_ = ExtractLaundryStatus(status, ApplianceDishwasher)
		_ = ExtractRangeStatus(status)
		_ = ExtractRefrigeratorStatus(status)
	})
}

// FuzzOAuthConfig fuzzes OAuth URL generation.
// Run with: go test -fuzz=FuzzOAuthConfig
func FuzzOAuthConfig(f *testing.F) {
	f.Add("client123", "https://example.com/callback", "state456")
	f.Add("", "", "")
	f.Add("id with spaces", "not-a-url", "special!@#$%chars")

	f.Fuzz(func(t *testing.T, clientID, redirectURL, state string) {
		config := &OAuthConfig{
			ClientID:    clientID,
			RedirectURL: redirectURL,
		}
		// Should not panic
		_ = GetAuthorizationURL(config, state)
	})
}
