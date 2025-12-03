package smartthings

import (
	"context"
	"iter"
)

// Devices returns an iterator over all devices with automatic pagination.
// Stops iteration early if an error occurs or context is cancelled.
func (c *Client) Devices(ctx context.Context) iter.Seq2[Device, error] {
	return c.DevicesWithOptions(ctx, nil)
}

// DevicesWithOptions returns a device iterator with custom filtering options.
// The iterator automatically handles pagination, fetching additional pages as needed.
func (c *Client) DevicesWithOptions(ctx context.Context, opts *ListDevicesOptions) iter.Seq2[Device, error] {
	return func(yield func(Device, error) bool) {
		page := 0
		if opts != nil && opts.Page > 0 {
			page = opts.Page
		}

		for {
			select {
			case <-ctx.Done():
				yield(Device{}, ctx.Err())
				return
			default:
			}

			// Build options for this page
			reqOpts := &ListDevicesOptions{
				Max:  200, // default max
				Page: page,
			}
			if opts != nil {
				reqOpts.Capability = opts.Capability
				reqOpts.LocationID = opts.LocationID
				reqOpts.DeviceID = opts.DeviceID
				reqOpts.Type = opts.Type
				reqOpts.IncludeRestricted = opts.IncludeRestricted
				if opts.Max > 0 {
					reqOpts.Max = opts.Max
				}
			}

			resp, err := c.ListDevicesWithOptions(ctx, reqOpts)
			if err != nil {
				yield(Device{}, err)
				return
			}

			for _, device := range resp.Items {
				if !yield(device, nil) {
					return // caller stopped iteration
				}
			}

			if resp.Links.Next == "" || len(resp.Items) == 0 {
				return // no more pages
			}
			page++
		}
	}
}

// Locations returns an iterator over all locations.
func (c *Client) Locations(ctx context.Context) iter.Seq2[Location, error] {
	return func(yield func(Location, error) bool) {
		select {
		case <-ctx.Done():
			yield(Location{}, ctx.Err())
			return
		default:
		}

		locations, err := c.ListLocations(ctx)
		if err != nil {
			yield(Location{}, err)
			return
		}

		for _, loc := range locations {
			if !yield(loc, nil) {
				return
			}
		}
	}
}

// Rooms returns an iterator over all rooms in a location.
func (c *Client) Rooms(ctx context.Context, locationID string) iter.Seq2[Room, error] {
	return func(yield func(Room, error) bool) {
		if locationID == "" {
			yield(Room{}, ErrEmptyLocationID)
			return
		}

		select {
		case <-ctx.Done():
			yield(Room{}, ctx.Err())
			return
		default:
		}

		rooms, err := c.ListRooms(ctx, locationID)
		if err != nil {
			yield(Room{}, err)
			return
		}

		for _, room := range rooms {
			if !yield(room, nil) {
				return
			}
		}
	}
}

// Rules returns an iterator over all rules for a location.
func (c *Client) Rules(ctx context.Context, locationID string) iter.Seq2[Rule, error] {
	return func(yield func(Rule, error) bool) {
		select {
		case <-ctx.Done():
			yield(Rule{}, ctx.Err())
			return
		default:
		}

		rules, err := c.ListRules(ctx, locationID)
		if err != nil {
			yield(Rule{}, err)
			return
		}

		for _, rule := range rules {
			if !yield(rule, nil) {
				return
			}
		}
	}
}

// Scenes returns an iterator over all scenes for a location.
func (c *Client) Scenes(ctx context.Context, locationID string) iter.Seq2[Scene, error] {
	return func(yield func(Scene, error) bool) {
		select {
		case <-ctx.Done():
			yield(Scene{}, ctx.Err())
			return
		default:
		}

		scenes, err := c.ListScenes(ctx, locationID)
		if err != nil {
			yield(Scene{}, err)
			return
		}

		for _, scene := range scenes {
			if !yield(scene, nil) {
				return
			}
		}
	}
}

// DeviceEvents returns an iterator over device events with automatic pagination.
func (c *Client) DeviceEvents(ctx context.Context, deviceID string, opts *HistoryOptions) iter.Seq2[DeviceEvent, error] {
	return func(yield func(DeviceEvent, error) bool) {
		if deviceID == "" {
			yield(DeviceEvent{}, ErrEmptyDeviceID)
			return
		}

		page := 0
		if opts != nil && opts.Page > 0 {
			page = opts.Page
		}

		for {
			select {
			case <-ctx.Done():
				yield(DeviceEvent{}, ctx.Err())
				return
			default:
			}

			reqOpts := &HistoryOptions{
				Max:  200,
				Page: page,
			}
			if opts != nil {
				reqOpts.Before = opts.Before
				reqOpts.After = opts.After
				if opts.Max > 0 {
					reqOpts.Max = opts.Max
				}
			}

			resp, err := c.GetDeviceEvents(ctx, deviceID, reqOpts)
			if err != nil {
				yield(DeviceEvent{}, err)
				return
			}

			for _, event := range resp.Items {
				if !yield(event, nil) {
					return
				}
			}

			if resp.Links.Next == "" || len(resp.Items) == 0 {
				return
			}
			page++
		}
	}
}

// Apps returns an iterator over all apps.
func (c *Client) Apps(ctx context.Context) iter.Seq2[App, error] {
	return func(yield func(App, error) bool) {
		select {
		case <-ctx.Done():
			yield(App{}, ctx.Err())
			return
		default:
		}

		apps, err := c.ListApps(ctx)
		if err != nil {
			yield(App{}, err)
			return
		}

		for _, app := range apps {
			if !yield(app, nil) {
				return
			}
		}
	}
}

// DeviceProfiles returns an iterator over all device profiles.
func (c *Client) DeviceProfiles(ctx context.Context) iter.Seq2[DeviceProfileFull, error] {
	return func(yield func(DeviceProfileFull, error) bool) {
		select {
		case <-ctx.Done():
			yield(DeviceProfileFull{}, ctx.Err())
			return
		default:
		}

		profiles, err := c.ListDeviceProfiles(ctx)
		if err != nil {
			yield(DeviceProfileFull{}, err)
			return
		}

		for _, profile := range profiles {
			if !yield(profile, nil) {
				return
			}
		}
	}
}

// Capabilities returns an iterator over all capabilities.
func (c *Client) Capabilities(ctx context.Context) iter.Seq2[CapabilityReference, error] {
	return func(yield func(CapabilityReference, error) bool) {
		select {
		case <-ctx.Done():
			yield(CapabilityReference{}, ctx.Err())
			return
		default:
		}

		caps, err := c.ListCapabilities(ctx)
		if err != nil {
			yield(CapabilityReference{}, err)
			return
		}

		for _, cap := range caps {
			if !yield(cap, nil) {
				return
			}
		}
	}
}

// InstalledApps returns an iterator over all installed apps for a location.
func (c *Client) InstalledApps(ctx context.Context, locationID string) iter.Seq2[InstalledApp, error] {
	return func(yield func(InstalledApp, error) bool) {
		select {
		case <-ctx.Done():
			yield(InstalledApp{}, ctx.Err())
			return
		default:
		}

		apps, err := c.ListInstalledApps(ctx, locationID)
		if err != nil {
			yield(InstalledApp{}, err)
			return
		}

		for _, app := range apps {
			if !yield(app, nil) {
				return
			}
		}
	}
}

// Subscriptions returns an iterator over all subscriptions for an installed app.
func (c *Client) Subscriptions(ctx context.Context, installedAppID string) iter.Seq2[Subscription, error] {
	return func(yield func(Subscription, error) bool) {
		if installedAppID == "" {
			yield(Subscription{}, ErrEmptyInstalledAppID)
			return
		}

		select {
		case <-ctx.Done():
			yield(Subscription{}, ctx.Err())
			return
		default:
		}

		subs, err := c.ListSubscriptions(ctx, installedAppID)
		if err != nil {
			yield(Subscription{}, err)
			return
		}

		for _, sub := range subs {
			if !yield(sub, nil) {
				return
			}
		}
	}
}

// Schedules returns an iterator over all schedules for an installed app.
func (c *Client) Schedules(ctx context.Context, installedAppID string) iter.Seq2[Schedule, error] {
	return func(yield func(Schedule, error) bool) {
		if installedAppID == "" {
			yield(Schedule{}, ErrEmptyInstalledAppID)
			return
		}

		select {
		case <-ctx.Done():
			yield(Schedule{}, ctx.Err())
			return
		default:
		}

		schedules, err := c.ListSchedules(ctx, installedAppID)
		if err != nil {
			yield(Schedule{}, err)
			return
		}

		for _, schedule := range schedules {
			if !yield(schedule, nil) {
				return
			}
		}
	}
}

// Modes returns an iterator over all modes for a location.
func (c *Client) Modes(ctx context.Context, locationID string) iter.Seq2[Mode, error] {
	return func(yield func(Mode, error) bool) {
		if locationID == "" {
			yield(Mode{}, ErrEmptyLocationID)
			return
		}

		select {
		case <-ctx.Done():
			yield(Mode{}, ctx.Err())
			return
		default:
		}

		modes, err := c.ListModes(ctx, locationID)
		if err != nil {
			yield(Mode{}, err)
			return
		}

		for _, mode := range modes {
			if !yield(mode, nil) {
				return
			}
		}
	}
}

// Organizations returns an iterator over all organizations.
func (c *Client) Organizations(ctx context.Context) iter.Seq2[Organization, error] {
	return func(yield func(Organization, error) bool) {
		select {
		case <-ctx.Done():
			yield(Organization{}, ctx.Err())
			return
		default:
		}

		orgs, err := c.ListOrganizations(ctx)
		if err != nil {
			yield(Organization{}, err)
			return
		}

		for _, org := range orgs {
			if !yield(org, nil) {
				return
			}
		}
	}
}

// Channels returns an iterator over all channels.
func (c *Client) Channels(ctx context.Context, opts *ChannelListOptions) iter.Seq2[Channel, error] {
	return func(yield func(Channel, error) bool) {
		select {
		case <-ctx.Done():
			yield(Channel{}, ctx.Err())
			return
		default:
		}

		channels, err := c.ListChannels(ctx, opts)
		if err != nil {
			yield(Channel{}, err)
			return
		}

		for _, channel := range channels {
			if !yield(channel, nil) {
				return
			}
		}
	}
}

// Drivers returns an iterator over all edge drivers.
func (c *Client) Drivers(ctx context.Context) iter.Seq2[EdgeDriverSummary, error] {
	return func(yield func(EdgeDriverSummary, error) bool) {
		select {
		case <-ctx.Done():
			yield(EdgeDriverSummary{}, ctx.Err())
			return
		default:
		}

		drivers, err := c.ListDrivers(ctx)
		if err != nil {
			yield(EdgeDriverSummary{}, err)
			return
		}

		for _, driver := range drivers {
			if !yield(driver, nil) {
				return
			}
		}
	}
}

// SchemaApps returns an iterator over all ST Schema apps.
func (c *Client) SchemaApps(ctx context.Context, includeAllOrganizations bool) iter.Seq2[SchemaApp, error] {
	return func(yield func(SchemaApp, error) bool) {
		select {
		case <-ctx.Done():
			yield(SchemaApp{}, ctx.Err())
			return
		default:
		}

		apps, err := c.ListSchemaApps(ctx, includeAllOrganizations)
		if err != nil {
			yield(SchemaApp{}, err)
			return
		}

		for _, app := range apps {
			if !yield(app, nil) {
				return
			}
		}
	}
}

// InstalledSchemaApps returns an iterator over all installed ST Schema apps for a location.
func (c *Client) InstalledSchemaApps(ctx context.Context, locationID string) iter.Seq2[InstalledSchemaApp, error] {
	return func(yield func(InstalledSchemaApp, error) bool) {
		select {
		case <-ctx.Done():
			yield(InstalledSchemaApp{}, ctx.Err())
			return
		default:
		}

		apps, err := c.ListInstalledSchemaApps(ctx, locationID)
		if err != nil {
			yield(InstalledSchemaApp{}, err)
			return
		}

		for _, app := range apps {
			if !yield(app, nil) {
				return
			}
		}
	}
}

// SchemaAppInvitations returns an iterator over all invitations for a schema app.
func (c *Client) SchemaAppInvitations(ctx context.Context, schemaAppID string) iter.Seq2[SchemaAppInvitation, error] {
	return func(yield func(SchemaAppInvitation, error) bool) {
		if schemaAppID == "" {
			yield(SchemaAppInvitation{}, ErrEmptySchemaAppID)
			return
		}

		select {
		case <-ctx.Done():
			yield(SchemaAppInvitation{}, ctx.Err())
			return
		default:
		}

		invites, err := c.ListSchemaAppInvitations(ctx, schemaAppID)
		if err != nil {
			yield(SchemaAppInvitation{}, err)
			return
		}

		for _, invite := range invites {
			if !yield(invite, nil) {
				return
			}
		}
	}
}

// DevicePreferences returns an iterator over all device preferences.
func (c *Client) DevicePreferences(ctx context.Context, namespace string) iter.Seq2[DevicePreference, error] {
	return func(yield func(DevicePreference, error) bool) {
		select {
		case <-ctx.Done():
			yield(DevicePreference{}, ctx.Err())
			return
		default:
		}

		prefs, err := c.ListDevicePreferences(ctx, namespace)
		if err != nil {
			yield(DevicePreference{}, err)
			return
		}

		for _, pref := range prefs {
			if !yield(pref, nil) {
				return
			}
		}
	}
}

// EnrolledChannels returns an iterator over all enrolled channels for a hub.
func (c *Client) EnrolledChannels(ctx context.Context, hubID string) iter.Seq2[EnrolledChannel, error] {
	return func(yield func(EnrolledChannel, error) bool) {
		if hubID == "" {
			yield(EnrolledChannel{}, ErrEmptyHubID)
			return
		}

		select {
		case <-ctx.Done():
			yield(EnrolledChannel{}, ctx.Err())
			return
		default:
		}

		channels, err := c.ListEnrolledChannels(ctx, hubID)
		if err != nil {
			yield(EnrolledChannel{}, err)
			return
		}

		for _, channel := range channels {
			if !yield(channel, nil) {
				return
			}
		}
	}
}

// InstalledDrivers returns an iterator over all installed drivers on a hub.
func (c *Client) InstalledDrivers(ctx context.Context, hubID, deviceID string) iter.Seq2[InstalledDriver, error] {
	return func(yield func(InstalledDriver, error) bool) {
		if hubID == "" {
			yield(InstalledDriver{}, ErrEmptyHubID)
			return
		}

		select {
		case <-ctx.Done():
			yield(InstalledDriver{}, ctx.Err())
			return
		default:
		}

		drivers, err := c.ListInstalledDrivers(ctx, hubID, deviceID)
		if err != nil {
			yield(InstalledDriver{}, err)
			return
		}

		for _, driver := range drivers {
			if !yield(driver, nil) {
				return
			}
		}
	}
}

// AssignedDrivers returns an iterator over all drivers assigned to a channel.
func (c *Client) AssignedDrivers(ctx context.Context, channelID string) iter.Seq2[DriverChannelDetails, error] {
	return func(yield func(DriverChannelDetails, error) bool) {
		if channelID == "" {
			yield(DriverChannelDetails{}, ErrEmptyChannelID)
			return
		}

		select {
		case <-ctx.Done():
			yield(DriverChannelDetails{}, ctx.Err())
			return
		default:
		}

		drivers, err := c.ListAssignedDrivers(ctx, channelID)
		if err != nil {
			yield(DriverChannelDetails{}, err)
			return
		}

		for _, driver := range drivers {
			if !yield(driver, nil) {
				return
			}
		}
	}
}
