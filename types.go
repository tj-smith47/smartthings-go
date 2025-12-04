package smartthings

// Status represents the raw device status response as a flexible map.
// The SmartThings API returns deeply nested JSON structures that vary by device type.
type Status map[string]any

// DeviceType represents the type of device integration.
type DeviceType string

// Device type constants.
const (
	DeviceTypeDTH         DeviceType = "DTH"
	DeviceTypeEndpointApp DeviceType = "ENDPOINT_APP"
	DeviceTypeViper       DeviceType = "VIPER"
	DeviceTypeHub         DeviceType = "HUB"
	DeviceTypeBLE         DeviceType = "BLE"
	DeviceTypeBLED2D      DeviceType = "BLE_D2D"
	DeviceTypeMobile      DeviceType = "MOBILE"
	DeviceTypeOCF         DeviceType = "OCF"
	DeviceTypeLAN         DeviceType = "LAN"
	DeviceTypeVideo       DeviceType = "VIDEO"
)

// Device represents a SmartThings device with all available API fields.
type Device struct {
	DeviceID               string           `json:"deviceId"`
	Name                   string           `json:"name"`
	Label                  string           `json:"label"`
	ManufacturerName       string           `json:"manufacturerName,omitempty"`
	PresentationID         string           `json:"presentationId,omitempty"`
	DeviceTypeID           string           `json:"deviceTypeId,omitempty"`
	Type                   DeviceType       `json:"type,omitempty"`
	RoomID                 string           `json:"roomId,omitempty"`
	LocationID             string           `json:"locationId,omitempty"`
	ParentDeviceID         string           `json:"parentDeviceId,omitempty"`
	ChildDevices           []ChildDevice    `json:"childDevices,omitempty"`
	Components             []Component      `json:"components,omitempty"`
	Profile                *DeviceProfile   `json:"profile,omitempty"`
	App                    *DeviceApp       `json:"app,omitempty"`
	OCF                    *OCFDeviceInfo   `json:"ocf,omitempty"`
	Viper                  *ViperDeviceInfo `json:"viper,omitempty"`
	DeviceManufacturerCode string           `json:"deviceManufacturerCode,omitempty"`
	CreateTime             string           `json:"createTime,omitempty"`
	RestrictionTier        int              `json:"restrictionTier,omitempty"`
	Allowed                []string         `json:"allowed,omitempty"`
}

// ChildDevice represents a reference to a child device.
type ChildDevice struct {
	DeviceID string `json:"deviceId"`
}

// DeviceProfile references a device profile.
type DeviceProfile struct {
	ID string `json:"id"`
}

// DeviceApp references the installed app for the device.
type DeviceApp struct {
	InstalledAppID string         `json:"installedAppId"`
	ExternalID     string         `json:"externalId,omitempty"`
	Profile        *DeviceProfile `json:"profile,omitempty"`
}

// OCFDeviceInfo contains OCF (Open Connectivity Foundation) device information.
type OCFDeviceInfo struct {
	DeviceID                  string `json:"ocfDeviceId,omitempty"`
	Name                      string `json:"name,omitempty"`
	SpecVersion               string `json:"specVersion,omitempty"`
	VerticalDomainSpecVersion string `json:"verticalDomainSpecVersion,omitempty"`
	ManufacturerName          string `json:"manufacturerName,omitempty"`
	ModelNumber               string `json:"modelNumber,omitempty"`
	PlatformVersion           string `json:"platformVersion,omitempty"`
	PlatformOS                string `json:"platformOS,omitempty"`
	HwVersion                 string `json:"hwVersion,omitempty"`
	FirmwareVersion           string `json:"firmwareVersion,omitempty"`
	VendorID                  string `json:"vendorId,omitempty"`
}

// ViperDeviceInfo contains Viper (Samsung Connect) device information.
type ViperDeviceInfo struct {
	UniqueID        string `json:"uniqueIdentifier,omitempty"`
	MACAddress      string `json:"macAddress,omitempty"`
	HubID           string `json:"hubId,omitempty"`
	ProvisionedTime string `json:"provisionedTime,omitempty"`
}

// Component represents a device component (e.g., "main", "cooler", "freezer").
type Component struct {
	ID           string           `json:"id"`
	Label        string           `json:"label,omitempty"`
	Capabilities []CapabilityRef  `json:"capabilities,omitempty"`
	Categories   []DeviceCategory `json:"categories,omitempty"`
	Icon         string           `json:"icon,omitempty"`
}

// CapabilityRef references a capability with its version.
type CapabilityRef struct {
	ID      string `json:"id"`
	Version int    `json:"version,omitempty"`
}

// DeviceCategory describes a device's category.
type DeviceCategory struct {
	Name         string `json:"name"`
	CategoryType string `json:"categoryType,omitempty"`
}

// DeviceHealth represents the health status of a device.
type DeviceHealth struct {
	DeviceID        string `json:"deviceId"`
	State           string `json:"state"` // ONLINE, OFFLINE, UNKNOWN
	LastUpdatedDate string `json:"lastUpdatedDate,omitempty"`
}

// DeviceUpdate is the request body for updating a device.
type DeviceUpdate struct {
	Label string `json:"label,omitempty"`
}

// PageInfo contains pagination information from API responses.
type PageInfo struct {
	TotalPages   int `json:"totalPages,omitempty"`
	TotalResults int `json:"totalResults,omitempty"`
	CurrentPage  int `json:"currentPage,omitempty"`
}

// Links contains pagination links from API responses.
type Links struct {
	Next     string `json:"next,omitempty"`
	Previous string `json:"previous,omitempty"`
}

// ListDevicesOptions contains options for listing devices with pagination and filtering.
type ListDevicesOptions struct {
	Capability        []string // Filter by capability
	LocationID        []string // Filter by location
	DeviceID          []string // Filter by device IDs
	Type              string   // Filter by device type
	Max               int      // Max results per page (1-200, default 200)
	Page              int      // Page number (0-based)
	IncludeRestricted bool     // Include restricted devices
}

// PagedDevices is the response from ListDevicesWithOptions.
type PagedDevices struct {
	Items    []Device `json:"items"`
	Links    Links    `json:"_links,omitempty"`
	PageInfo PageInfo `json:"_page,omitempty"`
}

// DeviceListResponse is the response from the list devices API.
// Deprecated: Use PagedDevices instead for pagination support.
type DeviceListResponse struct {
	Items []Device `json:"items"`
}

// Command represents a SmartThings command to execute on a device.
type Command struct {
	Component  string `json:"component,omitempty"`
	Capability string `json:"capability"`
	Command    string `json:"command"`
	Arguments  []any  `json:"arguments,omitempty"`
}

// CommandRequest is the request body for executing commands.
type CommandRequest struct {
	Commands []Command `json:"commands"`
}

// TVStatus represents the current status of a Samsung TV.
type TVStatus struct {
	Power       string `json:"power"`        // "on" or "off"
	Volume      int    `json:"volume"`       // 0-100
	Muted       bool   `json:"muted"`        // true if muted
	InputSource string `json:"input_source"` // e.g., "HDMI1", "Netflix"
}

// TVInput represents an available TV input source.
type TVInput struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// TVApp represents an installed TV application.
type TVApp struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ApplianceStatus represents the status of a Samsung appliance (washer, dryer, dishwasher).
type ApplianceStatus struct {
	State          string  `json:"state"`                     // "run", "stop", "pause", etc.
	RemainingMins  *int    `json:"remaining_mins,omitempty"`  // Minutes remaining
	CompletionTime *string `json:"completion_time,omitempty"` // ISO8601 completion time
	CycleProgress  *int    `json:"cycle_progress,omitempty"`  // Percentage 0-100
}

// RangeStatus represents the status of a Samsung range/oven.
type RangeStatus struct {
	CooktopActive  bool `json:"cooktop_active"`             // true if any burner is on
	OvenActive     bool `json:"oven_active"`                // true if oven is running
	OvenTemp       *int `json:"oven_temp,omitempty"`        // Current oven temperature (F)
	OvenTargetTemp *int `json:"oven_target_temp,omitempty"` // Target oven temperature (F)
}

// RefrigeratorStatus represents the status of a Samsung refrigerator.
type RefrigeratorStatus struct {
	FridgeTemp  *int `json:"fridge_temp,omitempty"`  // Fridge temperature (F)
	FreezerTemp *int `json:"freezer_temp,omitempty"` // Freezer temperature (F)
	DoorOpen    bool `json:"door_open"`              // true if any door is open
}

// BrilliantDeviceStatus represents the status of a Brilliant smart switch.
type BrilliantDeviceStatus struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	IsOn       bool   `json:"is_on"`
	Type       string `json:"type"`       // "switch", "dimmer"
	Brightness *int   `json:"brightness"` // 0-100 for dimmers
}

// GenericApplianceStatus provides a unified status structure for ANY Samsung appliance.
// This works with washers, dryers, dishwashers, microwaves, air conditioners,
// robot vacuums, air purifiers, ovens, and other Samsung CE devices.
// Use ExtractGenericApplianceStatus to auto-discover capabilities from any device.
type GenericApplianceStatus struct {
	// State is the operating state (e.g., "idle", "running", "paused", "finished").
	State string `json:"state"`

	// RemainingMins is the time remaining in minutes (if applicable).
	RemainingMins *int `json:"remaining_mins,omitempty"`

	// CompletionTime is the ISO8601 completion timestamp (if applicable).
	CompletionTime *string `json:"completion_time,omitempty"`

	// Progress is the cycle progress 0-100% (if applicable).
	Progress *int `json:"progress,omitempty"`

	// Temperature is the current temperature in Fahrenheit (if applicable).
	Temperature *int `json:"temperature,omitempty"`

	// TargetTemp is the target temperature in Fahrenheit (if applicable).
	TargetTemp *int `json:"target_temp,omitempty"`

	// Mode is the current operating mode (e.g., "cooling", "heating", "auto").
	Mode string `json:"mode,omitempty"`

	// PowerConsumption is the current power usage in watts (if available).
	PowerConsumption *float64 `json:"power_consumption,omitempty"`

	// DoorOpen indicates if any door/lid is open (if applicable).
	DoorOpen bool `json:"door_open,omitempty"`

	// Extra contains additional capability-specific data that doesn't fit standard fields.
	Extra map[string]any `json:"extra,omitempty"`

	// DiscoveredCapabilities lists all capability names found on the device.
	DiscoveredCapabilities []string `json:"discovered_capabilities,omitempty"`
}
