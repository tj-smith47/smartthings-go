package smartthings

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_GetLocationServiceInfo(t *testing.T) {
	tests := []struct {
		name       string
		locationID string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			locationID: "loc1",
			response:   `{"locationId": "loc1", "city": "Test City", "latitude": 37.5, "longitude": -122.0}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty location ID returns all",
			locationID: "",
			response:   `{"locationId": "", "city": "Default"}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.GetLocationServiceInfo(context.Background(), tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetServiceCapabilitiesList(t *testing.T) {
	tests := []struct {
		name       string
		locationID string
		response   string
		statusCode int
		wantErr    bool
		wantCount  int
	}{
		{
			name:       "successful response",
			locationID: "loc1",
			response:   `{"items": ["weather", "airQuality", "forecast"]}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  3,
		},
		{
			name:       "empty list",
			locationID: "loc1",
			response:   `{"items": []}`,
			statusCode: http.StatusOK,
			wantErr:    false,
			wantCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			caps, err := client.GetServiceCapabilitiesList(context.Background(), tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if !tt.wantErr && len(caps) != tt.wantCount {
				t.Errorf("expected %d capabilities, got %d", tt.wantCount, len(caps))
			}
		})
	}
}

func TestClient_CreateServiceSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscription   *ServiceSubscriptionRequest
		installedAppID string
		locationID     string
		response       string
		statusCode     int
		wantErr        bool
	}{
		{
			name: "successful creation",
			subscription: &ServiceSubscriptionRequest{
				Type:                   ServiceSubscriptionDirect,
				SubscribedCapabilities: []ServiceCapability{ServiceCapabilityWeather},
			},
			installedAppID: "app1",
			locationID:     "loc1",
			response:       `{"subscriptionId": "sub1", "capabilities": ["weather"]}`,
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "nil subscription",
			subscription:   nil,
			installedAppID: "app1",
			locationID:     "loc1",
			response:       ``,
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.CreateServiceSubscription(context.Background(), tt.subscription, tt.installedAppID, tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_DeleteServiceSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		installedAppID string
		locationID     string
		statusCode     int
		wantErr        bool
	}{
		{
			name:           "successful deletion",
			subscriptionID: "sub1",
			installedAppID: "app1",
			locationID:     "loc1",
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty subscription ID",
			subscriptionID: "",
			installedAppID: "app1",
			locationID:     "loc1",
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			err := client.DeleteServiceSubscription(context.Background(), tt.subscriptionID, tt.installedAppID, tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_DeleteAllServiceSubscriptions(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient("test-token", WithBaseURL(server.URL))
	err := client.DeleteAllServiceSubscriptions(context.Background(), "app1", "loc1")

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestClient_GetServiceCapability(t *testing.T) {
	tests := []struct {
		name       string
		capability ServiceCapability
		locationID string
		response   string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful response",
			capability: ServiceCapabilityWeather,
			locationID: "loc1",
			response:   `{"locationId": "loc1", "weather": {"relativeHumidity": 75}}`,
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "empty capability",
			capability: "",
			locationID: "loc1",
			response:   ``,
			statusCode: http.StatusOK,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.GetServiceCapability(context.Background(), tt.capability, tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_GetServiceCapabilitiesData(t *testing.T) {
	tests := []struct {
		name         string
		capabilities []ServiceCapability
		locationID   string
		response     string
		statusCode   int
		wantErr      bool
	}{
		{
			name:         "successful response",
			capabilities: []ServiceCapability{ServiceCapabilityWeather, ServiceCapabilityAirQuality},
			locationID:   "loc1",
			response:     `{"locationId": "loc1", "weather": {}, "airQuality": {}}`,
			statusCode:   http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "empty capabilities",
			capabilities: []ServiceCapability{},
			locationID:   "loc1",
			response:     ``,
			statusCode:   http.StatusOK,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(tt.response))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.GetServiceCapabilitiesData(context.Background(), tt.capabilities, tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestBuildServicePath(t *testing.T) {
	tests := []struct {
		installedAppID string
		locationID     string
		suffix         string
		want           string
	}{
		{"app1", "loc1", "/subscriptions", "/services/coordinate/installedapps/app1/locations/loc1/subscriptions"},
		{"", "loc1", "/subscriptions", "/services/coordinate/locations/loc1/subscriptions"},
		{"app1", "", "/subscriptions", "/services/coordinate/installedapps/app1/subscriptions"},
		{"", "", "/capabilities", "/services/coordinate/capabilities"},
	}

	for _, tt := range tests {
		got := buildServicePath(tt.installedAppID, tt.locationID, tt.suffix)
		if got != tt.want {
			t.Errorf("buildServicePath(%q, %q, %q) = %q, want %q",
				tt.installedAppID, tt.locationID, tt.suffix, got, tt.want)
		}
	}
}

func TestClient_UpdateServiceSubscription(t *testing.T) {
	tests := []struct {
		name           string
		subscriptionID string
		subscription   *ServiceSubscriptionRequest
		installedAppID string
		locationID     string
		statusCode     int
		wantErr        bool
	}{
		{
			name:           "successful update",
			subscriptionID: "sub1",
			subscription: &ServiceSubscriptionRequest{
				Type:                   ServiceSubscriptionDirect,
				SubscribedCapabilities: []ServiceCapability{ServiceCapabilityWeather},
			},
			installedAppID: "app1",
			locationID:     "loc1",
			statusCode:     http.StatusOK,
			wantErr:        false,
		},
		{
			name:           "empty subscription ID",
			subscriptionID: "",
			subscription: &ServiceSubscriptionRequest{
				Type: ServiceSubscriptionDirect,
			},
			installedAppID: "app1",
			locationID:     "loc1",
			statusCode:     http.StatusOK,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Write([]byte(`{"subscriptionId": "sub1"}`))
			}))
			defer server.Close()

			client, _ := NewClient("test-token", WithBaseURL(server.URL))
			_, err := client.UpdateServiceSubscription(context.Background(), tt.subscriptionID, tt.subscription, tt.installedAppID, tt.locationID)

			if tt.wantErr && err == nil {
				t.Error("expected error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
