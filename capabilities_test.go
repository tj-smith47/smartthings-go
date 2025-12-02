package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListCapabilities(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/capabilities" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/capabilities")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := capabilityListResponse{
				Items: []CapabilityReference{
					{ID: "switch", Version: 1, Status: "live"},
					{ID: "temperatureMeasurement", Version: 1, Status: "live"},
					{ID: "motionSensor", Version: 1, Status: "live"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		caps, err := client.ListCapabilities(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(caps) != 3 {
			t.Errorf("got %d capabilities, want 3", len(caps))
		}
		if caps[0].ID != "switch" {
			t.Errorf("caps[0].ID = %q, want %q", caps[0].ID, "switch")
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(capabilityListResponse{Items: []CapabilityReference{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		caps, err := client.ListCapabilities(context.Background())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(caps) != 0 {
			t.Errorf("got %d capabilities, want 0", len(caps))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListCapabilities(context.Background())
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestClient_GetCapability(t *testing.T) {
	t.Run("successful response with version", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/capabilities/switch/1" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/capabilities/switch/1")
			}
			cap := Capability{
				ID:      "switch",
				Version: 1,
				Status:  "live",
				Name:    "Switch",
				Attributes: map[string]CapabilityAttribute{
					"switch": {
						Schema: AttributeSchema{
							Type: "string",
							Enum: []interface{}{"on", "off"},
						},
					},
				},
				Commands: map[string]CapabilityCommand{
					"on":  {Name: "on"},
					"off": {Name: "off"},
				},
			}
			json.NewEncoder(w).Encode(cap)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		cap, err := client.GetCapability(context.Background(), "switch", 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cap.ID != "switch" {
			t.Errorf("ID = %q, want %q", cap.ID, "switch")
		}
		if cap.Name != "Switch" {
			t.Errorf("Name = %q, want %q", cap.Name, "Switch")
		}
		if len(cap.Attributes) != 1 {
			t.Errorf("got %d attributes, want 1", len(cap.Attributes))
		}
		if len(cap.Commands) != 2 {
			t.Errorf("got %d commands, want 2", len(cap.Commands))
		}
	})

	t.Run("successful response without version", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/capabilities/temperatureMeasurement" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/capabilities/temperatureMeasurement")
			}
			cap := Capability{
				ID:      "temperatureMeasurement",
				Version: 1,
				Status:  "live",
				Name:    "Temperature Measurement",
				Attributes: map[string]CapabilityAttribute{
					"temperature": {
						Schema: AttributeSchema{
							Type: "number",
						},
					},
				},
			}
			json.NewEncoder(w).Encode(cap)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		cap, err := client.GetCapability(context.Background(), "temperatureMeasurement", 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cap.ID != "temperatureMeasurement" {
			t.Errorf("ID = %q, want %q", cap.ID, "temperatureMeasurement")
		}
	})

	t.Run("empty capability ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetCapability(context.Background(), "", 1)
		if err != ErrEmptyCapabilityID {
			t.Errorf("expected ErrEmptyCapabilityID, got %v", err)
		}
	})

	t.Run("capability not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetCapability(context.Background(), "nonexistent", 1)
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})

	t.Run("capability with command arguments", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cap := Capability{
				ID:      "audioVolume",
				Version: 1,
				Status:  "live",
				Name:    "Audio Volume",
				Attributes: map[string]CapabilityAttribute{
					"volume": {
						Schema: AttributeSchema{
							Type:    "integer",
							Minimum: ptrFloat64(0),
							Maximum: ptrFloat64(100),
						},
						Setter: "setVolume",
					},
				},
				Commands: map[string]CapabilityCommand{
					"setVolume": {
						Name: "setVolume",
						Arguments: []CapabilityCommandArgument{
							{
								Name: "volume",
								Schema: AttributeSchema{
									Type:    "integer",
									Minimum: ptrFloat64(0),
									Maximum: ptrFloat64(100),
								},
							},
						},
					},
					"volumeUp":   {Name: "volumeUp"},
					"volumeDown": {Name: "volumeDown"},
				},
			}
			json.NewEncoder(w).Encode(cap)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		cap, err := client.GetCapability(context.Background(), "audioVolume", 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(cap.Commands) != 3 {
			t.Errorf("got %d commands, want 3", len(cap.Commands))
		}
		setVolume, ok := cap.Commands["setVolume"]
		if !ok {
			t.Fatal("missing setVolume command")
		}
		if len(setVolume.Arguments) != 1 {
			t.Errorf("setVolume has %d arguments, want 1", len(setVolume.Arguments))
		}
	})
}

// Helper function for tests
func ptrFloat64(f float64) *float64 {
	return &f
}
