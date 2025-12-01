package smartthings

import (
	"context"
)

// TV Control Methods
// These methods provide high-level TV control functionality.
// Most TV methods require the device ID to be passed as a parameter.

// GetTVStatus extracts TV status from a device status response.
// This is a helper that parses the status without making an API call.
func GetTVStatus(status Status) *TVStatus {
	tvStatus := &TVStatus{
		Power:       "off",
		Volume:      0,
		Muted:       false,
		InputSource: "",
	}

	// Extract power state: status["switch"]["switch"]["value"]
	if power, ok := GetString(status, "switch", "switch", "value"); ok {
		tvStatus.Power = power
	}

	// Extract volume: status["audioVolume"]["volume"]["value"]
	if volume, ok := GetInt(status, "audioVolume", "volume", "value"); ok {
		tvStatus.Volume = volume
	}

	// Extract mute state: status["audioMute"]["mute"]["value"]
	if mute, ok := GetString(status, "audioMute", "mute", "value"); ok {
		tvStatus.Muted = mute == "muted"
	}

	// Extract input source: status["mediaInputSource"]["inputSource"]["value"]
	if input, ok := GetString(status, "mediaInputSource", "inputSource", "value"); ok {
		tvStatus.InputSource = input
	}

	return tvStatus
}

// FetchTVStatus fetches and parses the current TV status from the API.
func (c *Client) FetchTVStatus(ctx context.Context, deviceID string) (*TVStatus, error) {
	status, err := c.GetDeviceStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return GetTVStatus(status), nil
}

// SetTVPower turns the TV on or off.
func (c *Client) SetTVPower(ctx context.Context, deviceID string, on bool) error {
	command := "off"
	if on {
		command = "on"
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("switch", command))
}

// SetTVVolume sets the TV volume (0-100).
func (c *Client) SetTVVolume(ctx context.Context, deviceID string, volume int) error {
	volume = max(0, min(volume, 100))
	return c.ExecuteCommand(ctx, deviceID, NewCommand("audioVolume", "setVolume", volume))
}

// SetTVMute sets the TV mute state.
func (c *Client) SetTVMute(ctx context.Context, deviceID string, muted bool) error {
	command := "unmute"
	if muted {
		command = "mute"
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("audioMute", command))
}

// SetTVInput sets the TV input source.
func (c *Client) SetTVInput(ctx context.Context, deviceID, inputID string) error {
	if inputID == "" {
		return ErrEmptyInputID
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("mediaInputSource", "setInputSource", inputID))
}

// GetTVInputs extracts available TV inputs from a device status.
func GetTVInputs(status Status) []TVInput {
	// First try Samsung-specific path: samsungvd.mediaInputSource.supportedInputSourcesMap
	if arr, ok := GetArray(status, "samsungvd.mediaInputSource", "supportedInputSourcesMap", "value"); ok {
		inputs := make([]TVInput, 0, len(arr))
		for _, v := range arr {
			if inputMap, ok := v.(map[string]any); ok {
				id, _ := inputMap["id"].(string)
				name, _ := inputMap["name"].(string)
				if id != "" {
					if name == "" {
						name = id
					}
					inputs = append(inputs, TVInput{ID: id, Name: name})
				}
			}
		}
		if len(inputs) > 0 {
			return inputs
		}
	}

	// Fallback to legacy path: mediaInputSource.supportedInputSources
	if arr, ok := GetArray(status, "mediaInputSource", "supportedInputSources", "value"); ok {
		inputs := make([]TVInput, 0, len(arr))
		for _, v := range arr {
			if s, ok := v.(string); ok {
				inputs = append(inputs, TVInput{ID: s, Name: s})
			}
		}
		return inputs
	}

	return nil
}

// FetchTVInputs fetches and returns available TV inputs from the API.
func (c *Client) FetchTVInputs(ctx context.Context, deviceID string) ([]TVInput, error) {
	status, err := c.GetDeviceStatus(ctx, deviceID)
	if err != nil {
		return nil, err
	}
	return GetTVInputs(status), nil
}

// SendTVKey sends a remote control key press to the TV.
// Common keys: UP, DOWN, LEFT, RIGHT, ENTER, BACK, HOME, MENU, EXIT
func (c *Client) SendTVKey(ctx context.Context, deviceID, key string) error {
	if key == "" {
		return ErrEmptyKey
	}
	// Map common keys to Samsung format
	mappedKey := key
	if key == "ENTER" {
		mappedKey = "KEY_ENTER"
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("samsungvd.remoteControl", "send", mappedKey))
}

// SetTVChannel sets the TV channel directly.
func (c *Client) SetTVChannel(ctx context.Context, deviceID string, channel int) error {
	if channel < 0 {
		return ErrInvalidChannel
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("tvChannel", "setTvChannel", channel))
}

// TVChannelUp increases the TV channel.
func (c *Client) TVChannelUp(ctx context.Context, deviceID string) error {
	return c.ExecuteCommand(ctx, deviceID, NewCommand("tvChannel", "channelUp"))
}

// TVChannelDown decreases the TV channel.
func (c *Client) TVChannelDown(ctx context.Context, deviceID string) error {
	return c.ExecuteCommand(ctx, deviceID, NewCommand("tvChannel", "channelDown"))
}

// TVVolumeUp increases the TV volume.
func (c *Client) TVVolumeUp(ctx context.Context, deviceID string) error {
	return c.ExecuteCommand(ctx, deviceID, NewCommand("audioVolume", "volumeUp"))
}

// TVVolumeDown decreases the TV volume.
func (c *Client) TVVolumeDown(ctx context.Context, deviceID string) error {
	return c.ExecuteCommand(ctx, deviceID, NewCommand("audioVolume", "volumeDown"))
}

// TVPlay sends a play command to the TV.
func (c *Client) TVPlay(ctx context.Context, deviceID string) error {
	return c.ExecuteCommand(ctx, deviceID, NewCommand("mediaPlayback", "play"))
}

// TVPause sends a pause command to the TV.
func (c *Client) TVPause(ctx context.Context, deviceID string) error {
	return c.ExecuteCommand(ctx, deviceID, NewCommand("mediaPlayback", "pause"))
}

// TVStop sends a stop command to the TV.
func (c *Client) TVStop(ctx context.Context, deviceID string) error {
	return c.ExecuteCommand(ctx, deviceID, NewCommand("mediaPlayback", "stop"))
}

// extractStrings extracts a string array from a status path.
func extractStrings(status Status, keys ...string) []string {
	arr, ok := GetArray(status, keys...)
	if !ok {
		return nil
	}
	result := make([]string, 0, len(arr))
	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}
	return result
}

// Picture Mode Methods

// SetPictureMode sets the TV picture mode.
func (c *Client) SetPictureMode(ctx context.Context, deviceID, mode string) error {
	if mode == "" {
		return ErrEmptyMode
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("custom.picturemode", "setPictureMode", mode))
}

// GetPictureModes extracts available picture modes from a device status.
func GetPictureModes(status Status) []string {
	return extractStrings(status, "custom.picturemode", "supportedPictureModes", "value")
}

// GetCurrentPictureMode extracts the current picture mode from a device status.
func GetCurrentPictureMode(status Status) string {
	mode, _ := GetString(status, "custom.picturemode", "pictureMode", "value")
	return mode
}

// Sound Mode Methods

// SetSoundMode sets the TV sound mode.
func (c *Client) SetSoundMode(ctx context.Context, deviceID, mode string) error {
	if mode == "" {
		return ErrEmptyMode
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("custom.soundmode", "setSoundMode", mode))
}

// GetSoundModes extracts available sound modes from a device status.
func GetSoundModes(status Status) []string {
	return extractStrings(status, "custom.soundmode", "supportedSoundModes", "value")
}

// GetCurrentSoundMode extracts the current sound mode from a device status.
func GetCurrentSoundMode(status Status) string {
	mode, _ := GetString(status, "custom.soundmode", "soundMode", "value")
	return mode
}

// TV App Methods

// LaunchTVApp launches an app on the TV.
func (c *Client) LaunchTVApp(ctx context.Context, deviceID, appID string) error {
	if appID == "" {
		return ErrEmptyAppID
	}
	return c.ExecuteCommand(ctx, deviceID, NewCommand("custom.launchapp", "launchApp", appID))
}

// GetTVApps extracts available apps from a device status.
func GetTVApps(status Status) []TVApp {
	var apps []TVApp

	// Extract from custom.launchapp.supportedAppIds
	if arr, ok := GetArray(status, "custom.launchapp", "supportedAppIds", "value"); ok {
		for _, v := range arr {
			if appMap, ok := v.(map[string]any); ok {
				id, _ := appMap["id"].(string)
				name, _ := appMap["name"].(string)
				if id != "" {
					if name == "" {
						name = id
					}
					apps = append(apps, TVApp{ID: id, Name: name})
				}
			}
		}
	}

	return apps
}

// CommonTVApps returns a list of commonly available Samsung TV apps.
// Use this as a fallback when apps can't be retrieved from the device.
func CommonTVApps() []TVApp {
	return []TVApp{
		{ID: "Netflix", Name: "Netflix"},
		{ID: "YouTube", Name: "YouTube"},
		{ID: "Prime Video", Name: "Prime Video"},
		{ID: "Disney+", Name: "Disney+"},
		{ID: "Hulu", Name: "Hulu"},
		{ID: "HBO Max", Name: "Max"},
		{ID: "Apple TV", Name: "Apple TV"},
		{ID: "Peacock", Name: "Peacock"},
		{ID: "Paramount+", Name: "Paramount+"},
		{ID: "Spotify", Name: "Spotify"},
		{ID: "Plex", Name: "Plex"},
	}
}
