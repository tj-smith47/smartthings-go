package smartthings

import "context"

// SmartThingsClient defines the interface for SmartThings API operations.
// Both Client and OAuthClient implement this interface, enabling mocking for tests.
type SmartThingsClient interface {
	// Device Operations
	ListDevices(ctx context.Context) ([]Device, error)
	ListDevicesWithOptions(ctx context.Context, opts *ListDevicesOptions) (*PagedDevices, error)
	ListAllDevices(ctx context.Context) ([]Device, error)
	GetDevice(ctx context.Context, deviceID string) (*Device, error)
	GetDeviceStatus(ctx context.Context, deviceID string) (Status, error)
	GetDeviceFullStatus(ctx context.Context, deviceID string) (map[string]Status, error)
	GetDeviceStatusAllComponents(ctx context.Context, deviceID string) (Status, error)
	GetComponentStatus(ctx context.Context, deviceID, componentID string) (Status, error)
	ExecuteCommand(ctx context.Context, deviceID string, cmd Command) error
	ExecuteCommands(ctx context.Context, deviceID string, cmds []Command) error
	DeleteDevice(ctx context.Context, deviceID string) error
	UpdateDevice(ctx context.Context, deviceID string, update *DeviceUpdate) (*Device, error)
	GetDeviceHealth(ctx context.Context, deviceID string) (*DeviceHealth, error)

	// Location Operations
	ListLocations(ctx context.Context) ([]Location, error)
	GetLocation(ctx context.Context, locationID string) (*Location, error)
	CreateLocation(ctx context.Context, location *LocationCreate) (*Location, error)
	UpdateLocation(ctx context.Context, locationID string, update *LocationUpdate) (*Location, error)
	DeleteLocation(ctx context.Context, locationID string) error

	// Room Operations
	ListRooms(ctx context.Context, locationID string) ([]Room, error)
	GetRoom(ctx context.Context, locationID, roomID string) (*Room, error)
	CreateRoom(ctx context.Context, locationID string, room *RoomCreate) (*Room, error)
	UpdateRoom(ctx context.Context, locationID, roomID string, update *RoomUpdate) (*Room, error)
	DeleteRoom(ctx context.Context, locationID, roomID string) error

	// Scene Operations
	ListScenes(ctx context.Context, locationID string) ([]Scene, error)
	GetScene(ctx context.Context, sceneID string) (*Scene, error)
	ExecuteScene(ctx context.Context, sceneID string) error

	// Capability Operations
	ListCapabilities(ctx context.Context) ([]CapabilityReference, error)
	GetCapability(ctx context.Context, capabilityID string, version int) (*Capability, error)

	// Subscription Operations (webhooks)
	ListSubscriptions(ctx context.Context, installedAppID string) ([]Subscription, error)
	CreateSubscription(ctx context.Context, installedAppID string, sub *SubscriptionCreate) (*Subscription, error)
	DeleteSubscription(ctx context.Context, installedAppID, subscriptionID string) error
	DeleteAllSubscriptions(ctx context.Context, installedAppID string) error

	// Rule Operations
	ListRules(ctx context.Context, locationID string) ([]Rule, error)
	GetRule(ctx context.Context, ruleID string) (*Rule, error)
	CreateRule(ctx context.Context, locationID string, rule *RuleCreate) (*Rule, error)
	UpdateRule(ctx context.Context, ruleID string, rule *RuleUpdate) (*Rule, error)
	DeleteRule(ctx context.Context, ruleID string) error
	ExecuteRule(ctx context.Context, ruleID string) error

	// Schedule Operations
	ListSchedules(ctx context.Context, installedAppID string) ([]Schedule, error)
	GetSchedule(ctx context.Context, installedAppID, scheduleName string) (*Schedule, error)
	CreateSchedule(ctx context.Context, installedAppID string, schedule *ScheduleCreate) (*Schedule, error)
	DeleteSchedule(ctx context.Context, installedAppID, scheduleName string) error

	// InstalledApp Operations
	ListInstalledApps(ctx context.Context, locationID string) ([]InstalledApp, error)
	GetInstalledApp(ctx context.Context, installedAppID string) (*InstalledApp, error)
	DeleteInstalledApp(ctx context.Context, installedAppID string) error

	// Mode Operations
	ListModes(ctx context.Context, locationID string) ([]Mode, error)
	GetMode(ctx context.Context, locationID, modeID string) (*Mode, error)
	GetCurrentMode(ctx context.Context, locationID string) (*Mode, error)
	SetCurrentMode(ctx context.Context, locationID, modeID string) (*Mode, error)

	// History/Events Operations
	GetDeviceEvents(ctx context.Context, deviceID string, opts *HistoryOptions) (*PagedEvents, error)
	GetDeviceStates(ctx context.Context, deviceID string, opts *HistoryOptions) (*PagedStates, error)

	// App Operations
	ListApps(ctx context.Context) ([]App, error)
	GetApp(ctx context.Context, appID string) (*App, error)
	CreateApp(ctx context.Context, app *AppCreate) (*App, error)
	UpdateApp(ctx context.Context, appID string, update *AppUpdate) (*App, error)
	DeleteApp(ctx context.Context, appID string) error
	GetAppOAuth(ctx context.Context, appID string) (*AppOAuth, error)
	UpdateAppOAuth(ctx context.Context, appID string, oauth *AppOAuth) (*AppOAuth, error)
	GenerateAppOAuth(ctx context.Context, appID string) (*AppOAuthGenerated, error)

	// Device Profile Operations
	ListDeviceProfiles(ctx context.Context) ([]DeviceProfileFull, error)
	GetDeviceProfile(ctx context.Context, profileID string) (*DeviceProfileFull, error)
	CreateDeviceProfile(ctx context.Context, profile *DeviceProfileCreate) (*DeviceProfileFull, error)
	UpdateDeviceProfile(ctx context.Context, profileID string, update *DeviceProfileUpdate) (*DeviceProfileFull, error)
	DeleteDeviceProfile(ctx context.Context, profileID string) error
}

// Ensure Client implements SmartThingsClient at compile time.
var _ SmartThingsClient = (*Client)(nil)

// Ensure OAuthClient implements SmartThingsClient at compile time.
var _ SmartThingsClient = (*OAuthClient)(nil)
