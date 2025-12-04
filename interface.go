package smartthings

import (
	"context"
	"iter"
	"time"
)

// SmartThingsClient defines the interface for SmartThings API operations.
// Both Client and OAuthClient implement this interface, enabling mocking for tests.
type SmartThingsClient interface {
	// ============================================================================
	// Device Operations
	// ============================================================================

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
	Devices(ctx context.Context) iter.Seq2[Device, error]
	DevicesWithOptions(ctx context.Context, opts *ListDevicesOptions) iter.Seq2[Device, error]

	// ============================================================================
	// Batch Operations
	// ============================================================================

	ExecuteCommandBatch(ctx context.Context, deviceIDs []string, cmd Command, cfg *BatchConfig) []BatchResult
	ExecuteCommandsBatch(ctx context.Context, batch []BatchCommand, cfg *BatchConfig) []BatchResult
	GetDeviceStatusBatch(ctx context.Context, deviceIDs []string, cfg *BatchConfig) []BatchStatusResult

	// ============================================================================
	// Location Operations
	// ============================================================================

	ListLocations(ctx context.Context) ([]Location, error)
	GetLocation(ctx context.Context, locationID string) (*Location, error)
	CreateLocation(ctx context.Context, location *LocationCreate) (*Location, error)
	UpdateLocation(ctx context.Context, locationID string, update *LocationUpdate) (*Location, error)
	DeleteLocation(ctx context.Context, locationID string) error
	Locations(ctx context.Context) iter.Seq2[Location, error]

	// ============================================================================
	// Room Operations
	// ============================================================================

	ListRooms(ctx context.Context, locationID string) ([]Room, error)
	GetRoom(ctx context.Context, locationID, roomID string) (*Room, error)
	CreateRoom(ctx context.Context, locationID string, room *RoomCreate) (*Room, error)
	UpdateRoom(ctx context.Context, locationID, roomID string, update *RoomUpdate) (*Room, error)
	DeleteRoom(ctx context.Context, locationID, roomID string) error
	Rooms(ctx context.Context, locationID string) iter.Seq2[Room, error]

	// ============================================================================
	// Scene Operations
	// ============================================================================

	ListScenes(ctx context.Context, locationID string) ([]Scene, error)
	GetScene(ctx context.Context, sceneID string) (*Scene, error)
	ExecuteScene(ctx context.Context, sceneID string) error
	Scenes(ctx context.Context, locationID string) iter.Seq2[Scene, error]

	// ============================================================================
	// Capability Operations
	// ============================================================================

	ListCapabilities(ctx context.Context) ([]CapabilityReference, error)
	ListCapabilitiesWithOptions(ctx context.Context, opts *ListCapabilitiesOptions) ([]CapabilityReference, error)
	GetCapability(ctx context.Context, capabilityID string, version int) (*Capability, error)
	Capabilities(ctx context.Context) iter.Seq2[CapabilityReference, error]

	// ============================================================================
	// Subscription Operations (webhooks)
	// ============================================================================

	ListSubscriptions(ctx context.Context, installedAppID string) ([]Subscription, error)
	CreateSubscription(ctx context.Context, installedAppID string, sub *SubscriptionCreate) (*Subscription, error)
	DeleteSubscription(ctx context.Context, installedAppID, subscriptionID string) error
	DeleteAllSubscriptions(ctx context.Context, installedAppID string) error
	Subscriptions(ctx context.Context, installedAppID string) iter.Seq2[Subscription, error]

	// ============================================================================
	// Rule Operations
	// ============================================================================

	ListRules(ctx context.Context, locationID string) ([]Rule, error)
	GetRule(ctx context.Context, ruleID string) (*Rule, error)
	CreateRule(ctx context.Context, locationID string, rule *RuleCreate) (*Rule, error)
	UpdateRule(ctx context.Context, ruleID string, rule *RuleUpdate) (*Rule, error)
	DeleteRule(ctx context.Context, ruleID string) error
	ExecuteRule(ctx context.Context, ruleID string) error
	Rules(ctx context.Context, locationID string) iter.Seq2[Rule, error]

	// ============================================================================
	// Schedule Operations
	// ============================================================================

	ListSchedules(ctx context.Context, installedAppID string) ([]Schedule, error)
	GetSchedule(ctx context.Context, installedAppID, scheduleName string) (*Schedule, error)
	CreateSchedule(ctx context.Context, installedAppID string, schedule *ScheduleCreate) (*Schedule, error)
	DeleteSchedule(ctx context.Context, installedAppID, scheduleName string) error
	Schedules(ctx context.Context, installedAppID string) iter.Seq2[Schedule, error]

	// ============================================================================
	// InstalledApp Operations
	// ============================================================================

	ListInstalledApps(ctx context.Context, locationID string) ([]InstalledApp, error)
	GetInstalledApp(ctx context.Context, installedAppID string) (*InstalledApp, error)
	DeleteInstalledApp(ctx context.Context, installedAppID string) error
	ListInstalledAppConfigs(ctx context.Context, installedAppID string) ([]InstalledAppConfigItem, error)
	GetInstalledAppConfig(ctx context.Context, installedAppID, configID string) (*InstalledAppConfiguration, error)
	GetCurrentInstalledAppConfig(ctx context.Context, installedAppID string) (*InstalledAppConfiguration, error)
	InstalledApps(ctx context.Context, locationID string) iter.Seq2[InstalledApp, error]

	// ============================================================================
	// Mode Operations
	// ============================================================================

	ListModes(ctx context.Context, locationID string) ([]Mode, error)
	GetMode(ctx context.Context, locationID, modeID string) (*Mode, error)
	GetCurrentMode(ctx context.Context, locationID string) (*Mode, error)
	SetCurrentMode(ctx context.Context, locationID, modeID string) (*Mode, error)
	Modes(ctx context.Context, locationID string) iter.Seq2[Mode, error]

	// ============================================================================
	// History/Events Operations
	// ============================================================================

	GetDeviceEvents(ctx context.Context, deviceID string, opts *HistoryOptions) (*PagedEvents, error)
	GetDeviceStates(ctx context.Context, deviceID string, opts *HistoryOptions) (*PagedStates, error)
	DeviceEvents(ctx context.Context, deviceID string, opts *HistoryOptions) iter.Seq2[DeviceEvent, error]

	// ============================================================================
	// App Operations
	// ============================================================================

	ListApps(ctx context.Context) ([]App, error)
	GetApp(ctx context.Context, appID string) (*App, error)
	CreateApp(ctx context.Context, app *AppCreate) (*App, error)
	UpdateApp(ctx context.Context, appID string, update *AppUpdate) (*App, error)
	DeleteApp(ctx context.Context, appID string) error
	GetAppOAuth(ctx context.Context, appID string) (*AppOAuth, error)
	UpdateAppOAuth(ctx context.Context, appID string, oauth *AppOAuth) (*AppOAuth, error)
	GenerateAppOAuth(ctx context.Context, appID string) (*AppOAuthGenerated, error)
	Apps(ctx context.Context) iter.Seq2[App, error]

	// ============================================================================
	// Device Profile Operations
	// ============================================================================

	ListDeviceProfiles(ctx context.Context) ([]DeviceProfileFull, error)
	GetDeviceProfile(ctx context.Context, profileID string) (*DeviceProfileFull, error)
	CreateDeviceProfile(ctx context.Context, profile *DeviceProfileCreate) (*DeviceProfileFull, error)
	UpdateDeviceProfile(ctx context.Context, profileID string, update *DeviceProfileUpdate) (*DeviceProfileFull, error)
	DeleteDeviceProfile(ctx context.Context, profileID string) error
	DeviceProfiles(ctx context.Context) iter.Seq2[DeviceProfileFull, error]

	// ============================================================================
	// Device Preference Operations
	// ============================================================================

	ListDevicePreferences(ctx context.Context, namespace string) ([]DevicePreference, error)
	GetDevicePreference(ctx context.Context, preferenceID string) (*DevicePreference, error)
	CreateDevicePreference(ctx context.Context, pref *DevicePreferenceCreate) (*DevicePreference, error)
	UpdateDevicePreference(ctx context.Context, preferenceID string, pref *DevicePreference) (*DevicePreference, error)
	CreatePreferenceTranslations(ctx context.Context, preferenceID string, localization *PreferenceLocalization) (*PreferenceLocalization, error)
	GetPreferenceTranslations(ctx context.Context, preferenceID, locale string) (*PreferenceLocalization, error)
	ListPreferenceTranslations(ctx context.Context, preferenceID string) ([]LocaleReference, error)
	UpdatePreferenceTranslations(ctx context.Context, preferenceID string, localization *PreferenceLocalization) (*PreferenceLocalization, error)
	DevicePreferences(ctx context.Context, namespace string) iter.Seq2[DevicePreference, error]

	// ============================================================================
	// Presentation Operations
	// ============================================================================

	GeneratePresentation(ctx context.Context, profileID string) (*PresentationDeviceConfig, error)
	CreatePresentationConfig(ctx context.Context, config *PresentationDeviceConfigCreate) (*PresentationDeviceConfig, error)
	GetPresentationConfig(ctx context.Context, presentationID, manufacturerName string) (*PresentationDeviceConfig, error)
	GetDevicePresentation(ctx context.Context, presentationID, manufacturerName string) (*PresentationDevicePresentation, error)

	// ============================================================================
	// Hub Operations
	// ============================================================================

	GetHub(ctx context.Context, hubID string) (*Hub, error)
	GetHubCharacteristics(ctx context.Context, hubID string) (HubCharacteristics, error)
	ListEnrolledChannels(ctx context.Context, hubID string) ([]EnrolledChannel, error)
	ListInstalledDrivers(ctx context.Context, hubID string, deviceID string) ([]InstalledDriver, error)
	GetInstalledDriver(ctx context.Context, hubID, driverID string) (*InstalledDriver, error)
	InstallDriver(ctx context.Context, driverID, hubID, channelID string) error
	UninstallDriver(ctx context.Context, driverID, hubID string) error
	SwitchDriver(ctx context.Context, driverID, hubID, deviceID string, forceUpdate bool) error
	EnrolledChannels(ctx context.Context, hubID string) iter.Seq2[EnrolledChannel, error]
	InstalledDrivers(ctx context.Context, hubID, deviceID string) iter.Seq2[InstalledDriver, error]

	// ============================================================================
	// Edge Driver Operations
	// ============================================================================

	ListDrivers(ctx context.Context) ([]EdgeDriverSummary, error)
	ListDefaultDrivers(ctx context.Context) ([]EdgeDriver, error)
	GetDriver(ctx context.Context, driverID string) (*EdgeDriver, error)
	GetDriverRevision(ctx context.Context, driverID, version string) (*EdgeDriver, error)
	DeleteDriver(ctx context.Context, driverID string) error
	UploadDriver(ctx context.Context, archiveData []byte) (*EdgeDriver, error)
	Drivers(ctx context.Context) iter.Seq2[EdgeDriverSummary, error]

	// ============================================================================
	// Channel Operations
	// ============================================================================

	ListChannels(ctx context.Context, opts *ChannelListOptions) ([]Channel, error)
	GetChannel(ctx context.Context, channelID string) (*Channel, error)
	CreateChannel(ctx context.Context, channel *ChannelCreate) (*Channel, error)
	UpdateChannel(ctx context.Context, channelID string, update *ChannelUpdate) (*Channel, error)
	DeleteChannel(ctx context.Context, channelID string) error
	ListAssignedDrivers(ctx context.Context, channelID string) ([]DriverChannelDetails, error)
	AssignDriver(ctx context.Context, channelID, driverID, version string) (*DriverChannelDetails, error)
	UnassignDriver(ctx context.Context, channelID, driverID string) error
	GetDriverChannelMetaInfo(ctx context.Context, channelID, driverID string) (*EdgeDriver, error)
	EnrollHub(ctx context.Context, channelID, hubID string) error
	UnenrollHub(ctx context.Context, channelID, hubID string) error
	Channels(ctx context.Context, opts *ChannelListOptions) iter.Seq2[Channel, error]
	AssignedDrivers(ctx context.Context, channelID string) iter.Seq2[DriverChannelDetails, error]

	// ============================================================================
	// Virtual Device Operations
	// ============================================================================

	CreateVirtualDevice(ctx context.Context, req *VirtualDeviceCreateRequest) (*Device, error)
	CreateStandardVirtualDevice(ctx context.Context, req *VirtualDeviceStandardCreateRequest) (*Device, error)
	ListVirtualDevices(ctx context.Context, opts *VirtualDeviceListOptions) ([]Device, error)
	CreateVirtualDeviceEvents(ctx context.Context, deviceID string, events []VirtualDeviceEvent) (*VirtualDeviceEventsResponse, error)

	// ============================================================================
	// Schema (C2C Connector) Operations
	// ============================================================================

	ListSchemaApps(ctx context.Context, includeAllOrganizations bool) ([]SchemaApp, error)
	GetSchemaApp(ctx context.Context, appID string) (*SchemaApp, error)
	CreateSchemaApp(ctx context.Context, req *SchemaAppRequest, organizationID string) (*SchemaCreateResponse, error)
	UpdateSchemaApp(ctx context.Context, appID string, req *SchemaAppRequest, organizationID string) error
	DeleteSchemaApp(ctx context.Context, appID string) error
	GetSchemaAppPage(ctx context.Context, appID, locationID string) (*SchemaPage, error)
	RegenerateSchemaAppOAuth(ctx context.Context, appID string) (*SchemaCreateResponse, error)
	ListInstalledSchemaApps(ctx context.Context, locationID string) ([]InstalledSchemaApp, error)
	GetInstalledSchemaApp(ctx context.Context, isaID string) (*InstalledSchemaApp, error)
	DeleteInstalledSchemaApp(ctx context.Context, isaID string) error
	SchemaApps(ctx context.Context, includeAllOrganizations bool) iter.Seq2[SchemaApp, error]
	InstalledSchemaApps(ctx context.Context, locationID string) iter.Seq2[InstalledSchemaApp, error]

	// ============================================================================
	// Schema App Invitation Operations
	// ============================================================================

	CreateSchemaAppInvitation(ctx context.Context, invitation *SchemaAppInvitationCreate) (*SchemaAppInvitationID, error)
	ListSchemaAppInvitations(ctx context.Context, schemaAppID string) ([]SchemaAppInvitation, error)
	RevokeSchemaAppInvitation(ctx context.Context, invitationID string) error
	SchemaAppInvitations(ctx context.Context, schemaAppID string) iter.Seq2[SchemaAppInvitation, error]

	// ============================================================================
	// Organization Operations
	// ============================================================================

	ListOrganizations(ctx context.Context) ([]Organization, error)
	GetOrganization(ctx context.Context, organizationID string) (*Organization, error)
	Organizations(ctx context.Context) iter.Seq2[Organization, error]

	// ============================================================================
	// Notification Operations
	// ============================================================================

	CreateNotification(ctx context.Context, req *NotificationRequest) (*NotificationResponse, error)

	// ============================================================================
	// Service (Weather/Air Quality) Operations
	// ============================================================================

	GetLocationServiceInfo(ctx context.Context, locationID string) (*ServiceLocationInfo, error)
	GetServiceCapabilitiesList(ctx context.Context, locationID string) ([]ServiceCapability, error)
	GetServiceCapability(ctx context.Context, capability ServiceCapability, locationID string) (*ServiceCapabilityData, error)
	GetServiceCapabilitiesData(ctx context.Context, capabilities []ServiceCapability, locationID string) (*ServiceCapabilityData, error)
	CreateServiceSubscription(ctx context.Context, req *ServiceSubscriptionRequest, installedAppID, locationID string) (*ServiceNewSubscription, error)
	UpdateServiceSubscription(ctx context.Context, subscriptionID string, req *ServiceSubscriptionRequest, installedAppID, locationID string) (*ServiceNewSubscription, error)
	DeleteServiceSubscription(ctx context.Context, subscriptionID, installedAppID, locationID string) error
	DeleteAllServiceSubscriptions(ctx context.Context, installedAppID, locationID string) error

	// ============================================================================
	// TV Control Operations
	// ============================================================================

	FetchTVStatus(ctx context.Context, deviceID string) (*TVStatus, error)
	FetchTVInputs(ctx context.Context, deviceID string) ([]TVInput, error)
	SetTVPower(ctx context.Context, deviceID string, on bool) error
	SetTVVolume(ctx context.Context, deviceID string, volume int) error
	SetTVMute(ctx context.Context, deviceID string, muted bool) error
	SetTVInput(ctx context.Context, deviceID, inputID string) error
	SetTVChannel(ctx context.Context, deviceID string, channel int) error
	SendTVKey(ctx context.Context, deviceID, key string) error
	LaunchTVApp(ctx context.Context, deviceID, appID string) error
	TVPlay(ctx context.Context, deviceID string) error
	TVPause(ctx context.Context, deviceID string) error
	TVStop(ctx context.Context, deviceID string) error
	TVChannelUp(ctx context.Context, deviceID string) error
	TVChannelDown(ctx context.Context, deviceID string) error
	TVVolumeUp(ctx context.Context, deviceID string) error
	TVVolumeDown(ctx context.Context, deviceID string) error
	SetPictureMode(ctx context.Context, deviceID, mode string) error
	SetSoundMode(ctx context.Context, deviceID, mode string) error

	// ============================================================================
	// Rate Limit Operations
	// ============================================================================

	RateLimitInfo() *RateLimitInfo
	RateLimitResetTime() time.Time
	RemainingRequests() int
	ShouldThrottle(threshold int) bool
	WaitForRateLimit(ctx context.Context) error
	WaitForRateLimitErr(ctx context.Context, err error) error

	// ============================================================================
	// Cache Operations
	// ============================================================================

	InvalidateCache(resourceType string, ids ...string)
	InvalidateCapabilityCache()

	// ============================================================================
	// Token Operations
	// ============================================================================

	Token() string
	SetToken(token string)

	// ============================================================================
	// Logging Operations
	// ============================================================================

	LogRequest(ctx context.Context, method, path string)
	LogResponse(ctx context.Context, method, path string, statusCode int, duration time.Duration, err error)
	LogDeviceCommand(ctx context.Context, deviceID string, capability, command string, err error)
	LogRateLimit(ctx context.Context, info RateLimitInfo)
}

// Ensure Client implements SmartThingsClient at compile time.
var _ SmartThingsClient = (*Client)(nil)

// Ensure OAuthClient implements SmartThingsClient at compile time.
var _ SmartThingsClient = (*OAuthClient)(nil)
