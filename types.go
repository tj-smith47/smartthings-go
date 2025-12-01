package smartthings

// Status represents the raw device status response as a flexible map.
// The SmartThings API returns deeply nested JSON structures that vary by device type.
type Status map[string]any

// Device represents a SmartThings device.
type Device struct {
	DeviceID         string      `json:"deviceId"`
	Name             string      `json:"name"`
	Label            string      `json:"label"`
	ManufacturerName string      `json:"manufacturerName"`
	PresentationID   string      `json:"presentationId"`
	DeviceTypeID     string      `json:"deviceTypeId"`
	Type             string      `json:"type"`
	RoomID           string      `json:"roomId,omitempty"`
	Components       []Component `json:"components"`
}

// Component represents a device component (e.g., "main", "cooler", "freezer").
type Component struct {
	ID           string   `json:"id"`
	Label        string   `json:"label,omitempty"`
	Capabilities []string `json:"capabilities,omitempty"`
}

// DeviceListResponse is the response from the list devices API.
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
