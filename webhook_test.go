package smartthings

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestValidateWebhookSignature(t *testing.T) {
	secret := "test-secret-key"
	body := []byte(`{"lifecycle":"PING","executionId":"abc123"}`)

	// Compute valid signature
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	validSignature := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	t.Run("valid signature", func(t *testing.T) {
		if !ValidateWebhookSignature(secret, body, validSignature) {
			t.Error("expected valid signature to pass")
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		if ValidateWebhookSignature(secret, body, "invalid-signature") {
			t.Error("expected invalid signature to fail")
		}
	})

	t.Run("wrong secret", func(t *testing.T) {
		if ValidateWebhookSignature("wrong-secret", body, validSignature) {
			t.Error("expected wrong secret to fail")
		}
	})

	t.Run("modified body", func(t *testing.T) {
		modifiedBody := []byte(`{"lifecycle":"PING","executionId":"xyz789"}`)
		if ValidateWebhookSignature(secret, modifiedBody, validSignature) {
			t.Error("expected modified body to fail")
		}
	})

	t.Run("empty secret", func(t *testing.T) {
		if ValidateWebhookSignature("", body, validSignature) {
			t.Error("expected empty secret to fail")
		}
	})

	t.Run("empty body", func(t *testing.T) {
		if ValidateWebhookSignature(secret, []byte{}, validSignature) {
			t.Error("expected empty body to fail")
		}
	})

	t.Run("empty signature", func(t *testing.T) {
		if ValidateWebhookSignature(secret, body, "") {
			t.Error("expected empty signature to fail")
		}
	})
}

func TestParseWebhookRequest(t *testing.T) {
	secret := "test-secret-key"

	computeSignature := func(body []byte) string {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		return base64.StdEncoding.EncodeToString(mac.Sum(nil))
	}

	t.Run("valid PING event", func(t *testing.T) {
		body := []byte(`{"lifecycle":"PING","executionId":"exec-123","appId":"app-456","pingData":{"challenge":"abc123"}}`)
		signature := computeSignature(body)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(WebhookSignatureHeader, signature)

		event, err := ParseWebhookRequest(req, secret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Lifecycle != LifecyclePing {
			t.Errorf("Lifecycle = %q, want %q", event.Lifecycle, LifecyclePing)
		}
		if event.ExecutionID != "exec-123" {
			t.Errorf("ExecutionID = %q, want %q", event.ExecutionID, "exec-123")
		}
		if event.PingData == nil || event.PingData.Challenge != "abc123" {
			t.Error("expected PingData.Challenge = abc123")
		}
	})

	t.Run("valid EVENT lifecycle", func(t *testing.T) {
		body := []byte(`{
			"lifecycle":"EVENT",
			"executionId":"exec-789",
			"appId":"app-456",
			"eventData":{
				"authToken":"token-abc",
				"installedApp":{"installedAppId":"ia-123","locationId":"loc-456"},
				"events":[
					{"eventType":"DEVICE_EVENT","deviceEvent":{"deviceId":"dev-1","capability":"switch","attribute":"switch","value":"on"}}
				]
			}
		}`)
		signature := computeSignature(body)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(WebhookSignatureHeader, signature)

		event, err := ParseWebhookRequest(req, secret)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Lifecycle != LifecycleEvent {
			t.Errorf("Lifecycle = %q, want %q", event.Lifecycle, LifecycleEvent)
		}
		if event.EventData == nil {
			t.Fatal("expected EventData")
		}
		if len(event.EventData.Events) != 1 {
			t.Errorf("got %d events, want 1", len(event.EventData.Events))
		}
	})

	t.Run("missing signature header", func(t *testing.T) {
		body := []byte(`{"lifecycle":"PING"}`)
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))

		_, err := ParseWebhookRequest(req, secret)
		if err != ErrMissingSignature {
			t.Errorf("expected ErrMissingSignature, got %v", err)
		}
	})

	t.Run("invalid signature", func(t *testing.T) {
		body := []byte(`{"lifecycle":"PING"}`)
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(WebhookSignatureHeader, "invalid-signature")

		_, err := ParseWebhookRequest(req, secret)
		if err != ErrInvalidSignature {
			t.Errorf("expected ErrInvalidSignature, got %v", err)
		}
	})

	t.Run("empty body", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader([]byte{}))
		req.Header.Set(WebhookSignatureHeader, "some-signature")

		_, err := ParseWebhookRequest(req, secret)
		if err != ErrEmptyBody {
			t.Errorf("expected ErrEmptyBody, got %v", err)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		body := []byte(`not valid json`)
		signature := computeSignature(body)

		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		req.Header.Set(WebhookSignatureHeader, signature)

		_, err := ParseWebhookRequest(req, secret)
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})

	t.Run("skip validation with empty secret", func(t *testing.T) {
		body := []byte(`{"lifecycle":"PING","executionId":"exec-123"}`)
		req := httptest.NewRequest(http.MethodPost, "/webhook", bytes.NewReader(body))
		// No signature header set, empty secret should skip validation

		event, err := ParseWebhookRequest(req, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if event.Lifecycle != LifecyclePing {
			t.Errorf("Lifecycle = %q, want %q", event.Lifecycle, LifecyclePing)
		}
	})

	t.Run("read error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/webhook", &errorReader{})
		req.Header.Set(WebhookSignatureHeader, "some-signature")

		_, err := ParseWebhookRequest(req, secret)
		if err == nil {
			t.Fatal("expected error for read failure")
		}
	})
}

// errorReader is a helper that always returns an error on Read.
type errorReader struct{}

func (e *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

func TestPingResponse(t *testing.T) {
	challenge := "test-challenge-123"
	resp := PingResponse(challenge)

	pingData, ok := resp["pingData"].(map[string]string)
	if !ok {
		t.Fatal("expected pingData to be map[string]string")
	}
	if pingData["challenge"] != challenge {
		t.Errorf("challenge = %q, want %q", pingData["challenge"], challenge)
	}
}

func TestWebhookLifecycleConstants(t *testing.T) {
	tests := []struct {
		lifecycle WebhookLifecycle
		expected  string
	}{
		{LifecyclePing, "PING"},
		{LifecycleConfirmation, "CONFIRMATION"},
		{LifecycleConfiguration, "CONFIGURATION"},
		{LifecycleInstall, "INSTALL"},
		{LifecycleUpdate, "UPDATE"},
		{LifecycleEvent, "EVENT"},
		{LifecycleUninstall, "UNINSTALL"},
		{LifecycleOAuthCallback, "OAUTH_CALLBACK"},
	}

	for _, tt := range tests {
		if string(tt.lifecycle) != tt.expected {
			t.Errorf("%v = %q, want %q", tt.lifecycle, string(tt.lifecycle), tt.expected)
		}
	}
}
