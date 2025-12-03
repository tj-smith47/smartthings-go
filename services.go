package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ServiceCapability represents an available service capability.
type ServiceCapability string

const (
	// ServiceCapabilityWeather is weather data.
	ServiceCapabilityWeather ServiceCapability = "weather"
	// ServiceCapabilityAirQuality is air quality data.
	ServiceCapabilityAirQuality ServiceCapability = "airQuality"
	// ServiceCapabilityForecast is weather forecast data.
	ServiceCapabilityForecast ServiceCapability = "forecast"
	// ServiceCapabilityAirQualityForecast is air quality forecast data.
	ServiceCapabilityAirQualityForecast ServiceCapability = "airQualityForecast"
)

// ServiceSubscriptionType represents the type of service subscription.
type ServiceSubscriptionType string

const (
	// ServiceSubscriptionDirect sends events directly.
	ServiceSubscriptionDirect ServiceSubscriptionType = "DIRECT"
	// ServiceSubscriptionExecution triggers executions.
	ServiceSubscriptionExecution ServiceSubscriptionType = "EXECUTION"
)

// ServiceLocationInfo contains location service information.
type ServiceLocationInfo struct {
	LocationID    string                    `json:"locationId,omitempty"`
	City          string                    `json:"city,omitempty"`
	Latitude      float64                   `json:"latitude,omitempty"`
	Longitude     float64                   `json:"longitude,omitempty"`
	Subscriptions []ServiceSubscriptionInfo `json:"subscriptions,omitempty"`
}

// ServiceSubscriptionInfo contains subscription details.
type ServiceSubscriptionInfo struct {
	SubscriptionID string                  `json:"subscriptionId,omitempty"`
	Type           ServiceSubscriptionType `json:"type,omitempty"`
	Capabilities   []ServiceCapability     `json:"capabilities,omitempty"`
}

// ServiceSubscriptionRequest is the request for creating/updating a subscription.
type ServiceSubscriptionRequest struct {
	Type                   ServiceSubscriptionType `json:"type"`
	Predicate              string                  `json:"predicate,omitempty"`
	SubscribedCapabilities []ServiceCapability     `json:"subscribedCapabilities"`
}

// ServiceNewSubscription is the response from creating a subscription.
type ServiceNewSubscription struct {
	SubscriptionID string              `json:"subscriptionId,omitempty"`
	Capabilities   []ServiceCapability `json:"capabilities,omitempty"`
}

// ServiceCapabilityData contains service capability data.
type ServiceCapabilityData struct {
	LocationID         string               `json:"locationId,omitempty"`
	Weather            *WeatherData         `json:"weather,omitempty"`
	AirQuality         *AirQualityData      `json:"airQuality,omitempty"`
	Forecast           *WeatherForecastData `json:"forecast,omitempty"`
	AirQualityForecast *AirQualityForecast  `json:"airQualityForecast,omitempty"`
}

// WeatherData contains current weather information.
type WeatherData struct {
	CloudCeiling          *float64 `json:"cloudCeiling,omitempty"`
	CloudCoverPhrase      string   `json:"cloudCoverPhrase,omitempty"`
	IconCode              int      `json:"iconCode,omitempty"`
	ConditionState        string   `json:"conditionState,omitempty"`
	RelativeHumidity      int      `json:"relativeHumidity,omitempty"`
	SunriseTimeLocal      string   `json:"sunriseTimeLocal,omitempty"`
	SunsetTimeLocal       string   `json:"sunsetTimeLocal,omitempty"`
	Temperature           float64  `json:"temperature,omitempty"`
	TemperatureFeelsLike  float64  `json:"temperatureFeelsLike,omitempty"`
	UVDescription         string   `json:"uvDescription,omitempty"`
	UVIndex               int      `json:"uvIndex,omitempty"`
	Visibility            float64  `json:"visibility,omitempty"`
	WindDirection         int      `json:"windDirection,omitempty"`
	WindDirectionCardinal string   `json:"windDirectionCardinal,omitempty"`
	WindGust              *float64 `json:"windGust,omitempty"`
	WindSpeed             float64  `json:"windSpeed,omitempty"`
	WxPhraseLong          string   `json:"wxPhraseLong,omitempty"`
}

// AirQualityData contains air quality information.
type AirQualityData struct {
	AirQualityIndex int `json:"airQualityIndex,omitempty"`
	O3Amount        int `json:"o3Amount,omitempty"`
	O3Index         int `json:"o3Index,omitempty"`
	NO2Amount       int `json:"no2Amount,omitempty"`
	NO2Index        int `json:"no2Index,omitempty"`
	SO2Amount       int `json:"so2Amount,omitempty"`
	SO2Index        int `json:"so2Index,omitempty"`
	COAmount        int `json:"coAmount,omitempty"`
	COIndex         int `json:"coIndex,omitempty"`
	PM10Amount      int `json:"pm10Amount,omitempty"`
	PM10Index       int `json:"pm10Index,omitempty"`
	PM25Amount      int `json:"pm25Amount,omitempty"`
	PM25Index       int `json:"pm25Index,omitempty"`
}

// WeatherForecastData contains weather forecast information.
type WeatherForecastData struct {
	Daily  []DailyForecast  `json:"daily,omitempty"`
	Hourly []HourlyForecast `json:"hourly,omitempty"`
}

// DailyForecast contains daily forecast information.
type DailyForecast struct {
	ForecastDate        string  `json:"forecastDate,omitempty"`
	ConditionState      string  `json:"conditionState,omitempty"`
	RelativeHumidity    int     `json:"relativeHumidity,omitempty"`
	TemperatureMax      float64 `json:"temperatureMax,omitempty"`
	TemperatureMin      float64 `json:"temperatureMin,omitempty"`
	PrecipitationChance int     `json:"precipitationChance,omitempty"`
	WxPhraseLong        string  `json:"wxPhraseLong,omitempty"`
}

// HourlyForecast contains hourly forecast information.
type HourlyForecast struct {
	ForecastTime        string  `json:"forecastTime,omitempty"`
	ConditionState      string  `json:"conditionState,omitempty"`
	RelativeHumidity    int     `json:"relativeHumidity,omitempty"`
	Temperature         float64 `json:"temperature,omitempty"`
	PrecipitationChance int     `json:"precipitationChance,omitempty"`
	WxPhraseLong        string  `json:"wxPhraseLong,omitempty"`
}

// AirQualityForecast contains air quality forecast information.
type AirQualityForecast struct {
	Daily []DailyAirQualityForecast `json:"daily,omitempty"`
}

// DailyAirQualityForecast contains daily air quality forecast.
type DailyAirQualityForecast struct {
	ForecastDate    string `json:"forecastDate,omitempty"`
	AirQualityIndex int    `json:"airQualityIndex,omitempty"`
}

// GetLocationServiceInfo returns service information for a location.
func (c *Client) GetLocationServiceInfo(ctx context.Context, locationID string) (*ServiceLocationInfo, error) {
	path := "/services/coordinate/locations"
	if locationID != "" {
		path += "/" + locationID
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var info ServiceLocationInfo
	if err := json.Unmarshal(data, &info); err != nil {
		return nil, fmt.Errorf("GetLocationServiceInfo: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &info, nil
}

// GetServiceCapabilitiesList returns available service capabilities for a location.
func (c *Client) GetServiceCapabilitiesList(ctx context.Context, locationID string) ([]ServiceCapability, error) {
	path := "/services/coordinate/locations"
	if locationID != "" {
		path += "/" + locationID
	}
	path += "/capabilities"

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp struct {
		Capabilities []ServiceCapability `json:"items"`
	}
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("GetServiceCapabilitiesList: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Capabilities, nil
}

// CreateServiceSubscription creates a new service subscription.
func (c *Client) CreateServiceSubscription(ctx context.Context, req *ServiceSubscriptionRequest, installedAppID, locationID string) (*ServiceNewSubscription, error) {
	if req == nil {
		return nil, ErrEmptyServiceSubscriptionRequest
	}

	path := buildServicePath(installedAppID, locationID, "/subscriptions")

	data, err := c.post(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var sub ServiceNewSubscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, fmt.Errorf("CreateServiceSubscription: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &sub, nil
}

// UpdateServiceSubscription updates an existing service subscription.
func (c *Client) UpdateServiceSubscription(ctx context.Context, subscriptionID string, req *ServiceSubscriptionRequest, installedAppID, locationID string) (*ServiceNewSubscription, error) {
	if subscriptionID == "" {
		return nil, ErrEmptySubscriptionID
	}
	if req == nil {
		return nil, ErrEmptyServiceSubscriptionRequest
	}

	path := buildServicePath(installedAppID, locationID, "/subscriptions/"+subscriptionID)

	data, err := c.put(ctx, path, req)
	if err != nil {
		return nil, err
	}

	var sub ServiceNewSubscription
	if err := json.Unmarshal(data, &sub); err != nil {
		return nil, fmt.Errorf("UpdateServiceSubscription: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &sub, nil
}

// DeleteServiceSubscription deletes a service subscription.
func (c *Client) DeleteServiceSubscription(ctx context.Context, subscriptionID, installedAppID, locationID string) error {
	if subscriptionID == "" {
		return ErrEmptySubscriptionID
	}

	path := buildServicePath(installedAppID, locationID, "/subscriptions/"+subscriptionID)

	_, err := c.delete(ctx, path)
	return err
}

// DeleteAllServiceSubscriptions deletes all service subscriptions.
func (c *Client) DeleteAllServiceSubscriptions(ctx context.Context, installedAppID, locationID string) error {
	path := buildServicePath(installedAppID, locationID, "/subscriptions")
	_, err := c.delete(ctx, path)
	return err
}

// GetServiceCapability returns data for a single service capability.
func (c *Client) GetServiceCapability(ctx context.Context, capability ServiceCapability, locationID string) (*ServiceCapabilityData, error) {
	if capability == "" {
		return nil, ErrEmptyServiceCapability
	}

	path := "/services/coordinate/locations"
	if locationID != "" {
		path += "/" + locationID
	}
	path += "/capabilities/" + string(capability)

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var capData ServiceCapabilityData
	if err := json.Unmarshal(data, &capData); err != nil {
		return nil, fmt.Errorf("GetServiceCapability: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &capData, nil
}

// GetServiceCapabilitiesData returns data for multiple service capabilities.
func (c *Client) GetServiceCapabilitiesData(ctx context.Context, capabilities []ServiceCapability, locationID string) (*ServiceCapabilityData, error) {
	if len(capabilities) == 0 {
		return nil, ErrEmptyServiceCapabilities
	}

	path := "/services/coordinate/locations"
	if locationID != "" {
		path += "/" + locationID
	}

	// Build capability names for query parameter
	capNames := make([]string, len(capabilities))
	for i, cap := range capabilities {
		capNames[i] = string(cap)
	}
	path += "/capabilities?name=" + url.QueryEscape(strings.Join(capNames, ","))

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var capData ServiceCapabilityData
	if err := json.Unmarshal(data, &capData); err != nil {
		return nil, fmt.Errorf("GetServiceCapabilitiesData: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &capData, nil
}

// buildServicePath constructs the service API path.
func buildServicePath(installedAppID, locationID, suffix string) string {
	path := "/services/coordinate"
	if installedAppID != "" {
		path += "/installedapps/" + installedAppID
	}
	if locationID != "" {
		path += "/locations/" + locationID
	}
	return path + suffix
}
