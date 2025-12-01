package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetTVStatus(t *testing.T) {
	t.Run("full status", func(t *testing.T) {
		status := Status{
			"switch": map[string]any{
				"switch": map[string]any{"value": "on"},
			},
			"audioVolume": map[string]any{
				"volume": map[string]any{"value": float64(25)},
			},
			"audioMute": map[string]any{
				"mute": map[string]any{"value": "muted"},
			},
			"mediaInputSource": map[string]any{
				"inputSource": map[string]any{"value": "HDMI1"},
			},
		}

		result := GetTVStatus(status)
		if result.Power != "on" {
			t.Errorf("Power = %q, want %q", result.Power, "on")
		}
		if result.Volume != 25 {
			t.Errorf("Volume = %d, want 25", result.Volume)
		}
		if !result.Muted {
			t.Error("Muted should be true")
		}
		if result.InputSource != "HDMI1" {
			t.Errorf("InputSource = %q, want %q", result.InputSource, "HDMI1")
		}
	})

	t.Run("default values for missing fields", func(t *testing.T) {
		result := GetTVStatus(Status{})
		if result.Power != "off" {
			t.Errorf("Power = %q, want %q", result.Power, "off")
		}
		if result.Volume != 0 {
			t.Errorf("Volume = %d, want 0", result.Volume)
		}
		if result.Muted {
			t.Error("Muted should be false by default")
		}
		if result.InputSource != "" {
			t.Errorf("InputSource = %q, want empty", result.InputSource)
		}
	})

	t.Run("unmuted state", func(t *testing.T) {
		status := Status{
			"audioMute": map[string]any{
				"mute": map[string]any{"value": "unmuted"},
			},
		}
		result := GetTVStatus(status)
		if result.Muted {
			t.Error("Muted should be false for 'unmuted'")
		}
	})
}

func TestClient_FetchTVStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := Status{
			"switch": map[string]any{
				"switch": map[string]any{"value": "on"},
			},
			"audioVolume": map[string]any{
				"volume": map[string]any{"value": float64(50)},
			},
		}
		json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	result, err := client.FetchTVStatus(context.Background(), "tv-device")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Power != "on" {
		t.Errorf("Power = %q, want %q", result.Power, "on")
	}
	if result.Volume != 50 {
		t.Errorf("Volume = %d, want 50", result.Volume)
	}
}

func TestClient_SetTVPower(t *testing.T) {
	t.Run("turn on", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if len(req.Commands) != 1 {
				t.Fatalf("expected 1 command, got %d", len(req.Commands))
			}
			if req.Commands[0].Command != "on" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "on")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.SetTVPower(context.Background(), "tv-device", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("turn off", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Command != "off" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "off")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.SetTVPower(context.Background(), "tv-device", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestClient_SetTVVolume(t *testing.T) {
	tests := []struct {
		name       string
		volume     int
		wantVolume int
	}{
		{"normal volume", 50, 50},
		{"zero volume", 0, 0},
		{"max volume", 100, 100},
		{"negative clamped", -10, 0},
		{"over 100 clamped", 150, 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req CommandRequest
				json.NewDecoder(r.Body).Decode(&req)
				if len(req.Commands[0].Arguments) != 1 {
					t.Fatalf("expected 1 argument, got %d", len(req.Commands[0].Arguments))
				}
				vol := int(req.Commands[0].Arguments[0].(float64))
				if vol != tt.wantVolume {
					t.Errorf("volume = %d, want %d", vol, tt.wantVolume)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client, _ := NewClient("token", WithBaseURL(server.URL))
			err := client.SetTVVolume(context.Background(), "tv-device", tt.volume)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_SetTVMute(t *testing.T) {
	t.Run("mute", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Command != "mute" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "mute")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.SetTVMute(context.Background(), "tv-device", true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("unmute", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Command != "unmute" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "unmute")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.SetTVMute(context.Background(), "tv-device", false)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestClient_SetTVInput(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CommandRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Commands[0].Capability != "mediaInputSource" {
			t.Errorf("capability = %q, want %q", req.Commands[0].Capability, "mediaInputSource")
		}
		if len(req.Commands[0].Arguments) != 1 {
			t.Fatalf("expected 1 argument, got %d", len(req.Commands[0].Arguments))
		}
		if req.Commands[0].Arguments[0] != "HDMI2" {
			t.Errorf("input = %v, want %q", req.Commands[0].Arguments[0], "HDMI2")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	err := client.SetTVInput(context.Background(), "tv-device", "HDMI2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTVInputs(t *testing.T) {
	t.Run("Samsung format", func(t *testing.T) {
		status := Status{
			"samsungvd.mediaInputSource": map[string]any{
				"supportedInputSourcesMap": map[string]any{
					"value": []any{
						map[string]any{"id": "HDMI1", "name": "Cable Box"},
						map[string]any{"id": "HDMI2", "name": "Game Console"},
						map[string]any{"id": "HDMI3"}, // No name
					},
				},
			},
		}

		inputs := GetTVInputs(status)
		if len(inputs) != 3 {
			t.Fatalf("got %d inputs, want 3", len(inputs))
		}
		if inputs[0].ID != "HDMI1" || inputs[0].Name != "Cable Box" {
			t.Errorf("inputs[0] = %+v, want {HDMI1, Cable Box}", inputs[0])
		}
		if inputs[2].Name != "HDMI3" {
			t.Errorf("inputs[2].Name = %q, want %q (should default to ID)", inputs[2].Name, "HDMI3")
		}
	})

	t.Run("legacy format", func(t *testing.T) {
		status := Status{
			"mediaInputSource": map[string]any{
				"supportedInputSources": map[string]any{
					"value": []any{"HDMI1", "HDMI2", "USB"},
				},
			},
		}

		inputs := GetTVInputs(status)
		if len(inputs) != 3 {
			t.Fatalf("got %d inputs, want 3", len(inputs))
		}
		if inputs[0].ID != "HDMI1" || inputs[0].Name != "HDMI1" {
			t.Errorf("inputs[0] = %+v, want {HDMI1, HDMI1}", inputs[0])
		}
	})

	t.Run("empty status", func(t *testing.T) {
		inputs := GetTVInputs(Status{})
		if len(inputs) != 0 {
			t.Errorf("got %d inputs, want 0", len(inputs))
		}
	})

	t.Run("Samsung format takes precedence", func(t *testing.T) {
		status := Status{
			"samsungvd.mediaInputSource": map[string]any{
				"supportedInputSourcesMap": map[string]any{
					"value": []any{
						map[string]any{"id": "Samsung1"},
					},
				},
			},
			"mediaInputSource": map[string]any{
				"supportedInputSources": map[string]any{
					"value": []any{"Legacy1", "Legacy2"},
				},
			},
		}

		inputs := GetTVInputs(status)
		if len(inputs) != 1 {
			t.Fatalf("got %d inputs, want 1", len(inputs))
		}
		if inputs[0].ID != "Samsung1" {
			t.Errorf("inputs[0].ID = %q, want %q", inputs[0].ID, "Samsung1")
		}
	})
}

func TestClient_FetchTVInputs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := Status{
			"mediaInputSource": map[string]any{
				"supportedInputSources": map[string]any{
					"value": []any{"HDMI1", "HDMI2"},
				},
			},
		}
		json.NewEncoder(w).Encode(status)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	inputs, err := client.FetchTVInputs(context.Background(), "tv-device")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(inputs) != 2 {
		t.Errorf("got %d inputs, want 2", len(inputs))
	}
}

func TestClient_SendTVKey(t *testing.T) {
	tests := []struct {
		key     string
		wantKey string
	}{
		{"UP", "UP"},
		{"DOWN", "DOWN"},
		{"ENTER", "KEY_ENTER"}, // Special mapping
		{"BACK", "BACK"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req CommandRequest
				json.NewDecoder(r.Body).Decode(&req)
				if req.Commands[0].Arguments[0] != tt.wantKey {
					t.Errorf("key = %v, want %q", req.Commands[0].Arguments[0], tt.wantKey)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client, _ := NewClient("token", WithBaseURL(server.URL))
			err := client.SendTVKey(context.Background(), "tv-device", tt.key)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestClient_SetTVChannel(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CommandRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Commands[0].Capability != "tvChannel" {
			t.Errorf("capability = %q, want %q", req.Commands[0].Capability, "tvChannel")
		}
		if req.Commands[0].Command != "setTvChannel" {
			t.Errorf("command = %q, want %q", req.Commands[0].Command, "setTvChannel")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	err := client.SetTVChannel(context.Background(), "tv-device", 123)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_TVChannelUpDown(t *testing.T) {
	t.Run("channel up", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Command != "channelUp" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "channelUp")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.TVChannelUp(context.Background(), "tv-device")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("channel down", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Command != "channelDown" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "channelDown")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.TVChannelDown(context.Background(), "tv-device")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestClient_TVVolumeUpDown(t *testing.T) {
	t.Run("volume up", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Command != "volumeUp" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "volumeUp")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.TVVolumeUp(context.Background(), "tv-device")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("volume down", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var req CommandRequest
			json.NewDecoder(r.Body).Decode(&req)
			if req.Commands[0].Command != "volumeDown" {
				t.Errorf("command = %q, want %q", req.Commands[0].Command, "volumeDown")
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.TVVolumeDown(context.Background(), "tv-device")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestClient_TVPlaybackControls(t *testing.T) {
	tests := []struct {
		name    string
		method  func(*Client, context.Context, string) error
		wantCmd string
	}{
		{"play", (*Client).TVPlay, "play"},
		{"pause", (*Client).TVPause, "pause"},
		{"stop", (*Client).TVStop, "stop"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				var req CommandRequest
				json.NewDecoder(r.Body).Decode(&req)
				if req.Commands[0].Command != tt.wantCmd {
					t.Errorf("command = %q, want %q", req.Commands[0].Command, tt.wantCmd)
				}
				w.WriteHeader(http.StatusOK)
			}))
			defer server.Close()

			client, _ := NewClient("token", WithBaseURL(server.URL))
			err := tt.method(client, context.Background(), "tv-device")
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestGetPictureModes(t *testing.T) {
	t.Run("with modes", func(t *testing.T) {
		status := Status{
			"custom.picturemode": map[string]any{
				"supportedPictureModes": map[string]any{
					"value": []any{"Standard", "Dynamic", "Movie", "Natural"},
				},
			},
		}

		modes := GetPictureModes(status)
		if len(modes) != 4 {
			t.Fatalf("got %d modes, want 4", len(modes))
		}
		if modes[0] != "Standard" {
			t.Errorf("modes[0] = %q, want %q", modes[0], "Standard")
		}
	})

	t.Run("empty", func(t *testing.T) {
		modes := GetPictureModes(Status{})
		if modes != nil {
			t.Errorf("expected nil, got %v", modes)
		}
	})
}

func TestGetCurrentPictureMode(t *testing.T) {
	status := Status{
		"custom.picturemode": map[string]any{
			"pictureMode": map[string]any{
				"value": "Movie",
			},
		},
	}

	mode := GetCurrentPictureMode(status)
	if mode != "Movie" {
		t.Errorf("mode = %q, want %q", mode, "Movie")
	}
}

func TestGetSoundModes(t *testing.T) {
	status := Status{
		"custom.soundmode": map[string]any{
			"supportedSoundModes": map[string]any{
				"value": []any{"Standard", "Amplify", "Optimized"},
			},
		},
	}

	modes := GetSoundModes(status)
	if len(modes) != 3 {
		t.Fatalf("got %d modes, want 3", len(modes))
	}
}

func TestGetCurrentSoundMode(t *testing.T) {
	status := Status{
		"custom.soundmode": map[string]any{
			"soundMode": map[string]any{
				"value": "Amplify",
			},
		},
	}

	mode := GetCurrentSoundMode(status)
	if mode != "Amplify" {
		t.Errorf("mode = %q, want %q", mode, "Amplify")
	}
}

func TestClient_SetPictureMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CommandRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Commands[0].Capability != "custom.picturemode" {
			t.Errorf("capability = %q, want %q", req.Commands[0].Capability, "custom.picturemode")
		}
		if req.Commands[0].Command != "setPictureMode" {
			t.Errorf("command = %q, want %q", req.Commands[0].Command, "setPictureMode")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	err := client.SetPictureMode(context.Background(), "tv-device", "Movie")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_SetSoundMode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CommandRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Commands[0].Capability != "custom.soundmode" {
			t.Errorf("capability = %q, want %q", req.Commands[0].Capability, "custom.soundmode")
		}
		if req.Commands[0].Command != "setSoundMode" {
			t.Errorf("command = %q, want %q", req.Commands[0].Command, "setSoundMode")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	err := client.SetSoundMode(context.Background(), "tv-device", "Amplify")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestClient_LaunchTVApp(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req CommandRequest
		json.NewDecoder(r.Body).Decode(&req)
		if req.Commands[0].Capability != "custom.launchapp" {
			t.Errorf("capability = %q, want %q", req.Commands[0].Capability, "custom.launchapp")
		}
		if req.Commands[0].Arguments[0] != "Netflix" {
			t.Errorf("app = %v, want %q", req.Commands[0].Arguments[0], "Netflix")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client, _ := NewClient("token", WithBaseURL(server.URL))
	err := client.LaunchTVApp(context.Background(), "tv-device", "Netflix")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetTVApps(t *testing.T) {
	t.Run("with apps", func(t *testing.T) {
		status := Status{
			"custom.launchapp": map[string]any{
				"supportedAppIds": map[string]any{
					"value": []any{
						map[string]any{"id": "netflix", "name": "Netflix"},
						map[string]any{"id": "youtube", "name": "YouTube"},
						map[string]any{"id": "hulu"}, // No name
					},
				},
			},
		}

		apps := GetTVApps(status)
		if len(apps) != 3 {
			t.Fatalf("got %d apps, want 3", len(apps))
		}
		if apps[0].ID != "netflix" || apps[0].Name != "Netflix" {
			t.Errorf("apps[0] = %+v, want {netflix, Netflix}", apps[0])
		}
		if apps[2].Name != "hulu" {
			t.Errorf("apps[2].Name = %q, want %q (should default to ID)", apps[2].Name, "hulu")
		}
	})

	t.Run("empty", func(t *testing.T) {
		apps := GetTVApps(Status{})
		if len(apps) != 0 {
			t.Errorf("got %d apps, want 0", len(apps))
		}
	})
}

func TestCommonTVApps(t *testing.T) {
	apps := CommonTVApps()
	if len(apps) == 0 {
		t.Fatal("CommonTVApps returned empty list")
	}

	// Check for Netflix as a basic validation
	foundNetflix := false
	for _, app := range apps {
		if app.ID == "Netflix" {
			foundNetflix = true
			break
		}
	}
	if !foundNetflix {
		t.Error("CommonTVApps should include Netflix")
	}
}
