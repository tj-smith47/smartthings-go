package smartthings

import (
	"math"
	"testing"
)

func TestGetString(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		keys   []string
		want   string
		wantOk bool
	}{
		{
			name:   "simple key",
			data:   map[string]any{"key": "value"},
			keys:   []string{"key"},
			want:   "value",
			wantOk: true,
		},
		{
			name: "nested keys",
			data: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": "deep value",
					},
				},
			},
			keys:   []string{"level1", "level2", "level3"},
			want:   "deep value",
			wantOk: true,
		},
		{
			name:   "missing key",
			data:   map[string]any{"key": "value"},
			keys:   []string{"missing"},
			want:   "",
			wantOk: false,
		},
		{
			name:   "wrong type",
			data:   map[string]any{"key": 123},
			keys:   []string{"key"},
			want:   "",
			wantOk: false,
		},
		{
			name:   "empty data",
			data:   map[string]any{},
			keys:   []string{"key"},
			want:   "",
			wantOk: false,
		},
		{
			name:   "nil data",
			data:   nil,
			keys:   []string{"key"},
			want:   "",
			wantOk: false,
		},
		{
			name:   "empty keys",
			data:   map[string]any{"key": "value"},
			keys:   []string{},
			want:   "",
			wantOk: false,
		},
		{
			name: "intermediate key not a map",
			data: map[string]any{
				"level1": "not a map",
			},
			keys:   []string{"level1", "level2"},
			want:   "",
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetString(tt.data, tt.keys...)
			if got != tt.want {
				t.Errorf("GetString() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetString() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGetInt(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		keys   []string
		want   int
		wantOk bool
	}{
		{
			name:   "float64 value",
			data:   map[string]any{"volume": float64(50)},
			keys:   []string{"volume"},
			want:   50,
			wantOk: true,
		},
		{
			name:   "int value",
			data:   map[string]any{"count": 42},
			keys:   []string{"count"},
			want:   42,
			wantOk: true,
		},
		{
			name:   "int64 value",
			data:   map[string]any{"big": int64(1000000)},
			keys:   []string{"big"},
			want:   1000000,
			wantOk: true,
		},
		{
			name: "nested value",
			data: map[string]any{
				"audioVolume": map[string]any{
					"volume": map[string]any{
						"value": float64(75),
					},
				},
			},
			keys:   []string{"audioVolume", "volume", "value"},
			want:   75,
			wantOk: true,
		},
		{
			name:   "missing key",
			data:   map[string]any{"volume": float64(50)},
			keys:   []string{"missing"},
			want:   0,
			wantOk: false,
		},
		{
			name:   "wrong type - string",
			data:   map[string]any{"volume": "fifty"},
			keys:   []string{"volume"},
			want:   0,
			wantOk: false,
		},
		{
			name:   "NaN value",
			data:   map[string]any{"value": math.NaN()},
			keys:   []string{"value"},
			want:   0,
			wantOk: false,
		},
		{
			name:   "positive infinity",
			data:   map[string]any{"value": math.Inf(1)},
			keys:   []string{"value"},
			want:   0,
			wantOk: false,
		},
		{
			name:   "negative infinity",
			data:   map[string]any{"value": math.Inf(-1)},
			keys:   []string{"value"},
			want:   0,
			wantOk: false,
		},
		{
			name:   "zero value",
			data:   map[string]any{"value": float64(0)},
			keys:   []string{"value"},
			want:   0,
			wantOk: true,
		},
		{
			name:   "negative value",
			data:   map[string]any{"value": float64(-10)},
			keys:   []string{"value"},
			want:   -10,
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetInt(tt.data, tt.keys...)
			if got != tt.want {
				t.Errorf("GetInt() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetInt() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGetFloat(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		keys   []string
		want   float64
		wantOk bool
	}{
		{
			name:   "float64 value",
			data:   map[string]any{"temp": 72.5},
			keys:   []string{"temp"},
			want:   72.5,
			wantOk: true,
		},
		{
			name:   "int value",
			data:   map[string]any{"temp": 72},
			keys:   []string{"temp"},
			want:   72.0,
			wantOk: true,
		},
		{
			name:   "int64 value",
			data:   map[string]any{"temp": int64(72)},
			keys:   []string{"temp"},
			want:   72.0,
			wantOk: true,
		},
		{
			name: "nested value",
			data: map[string]any{
				"temperatureMeasurement": map[string]any{
					"temperature": map[string]any{
						"value": 20.5,
					},
				},
			},
			keys:   []string{"temperatureMeasurement", "temperature", "value"},
			want:   20.5,
			wantOk: true,
		},
		{
			name:   "missing key",
			data:   map[string]any{"temp": 72.5},
			keys:   []string{"missing"},
			want:   0,
			wantOk: false,
		},
		{
			name:   "wrong type - string",
			data:   map[string]any{"temp": "hot"},
			keys:   []string{"temp"},
			want:   0,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetFloat(tt.data, tt.keys...)
			if got != tt.want {
				t.Errorf("GetFloat() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetFloat() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGetBool(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		keys   []string
		want   bool
		wantOk bool
	}{
		{
			name:   "true value",
			data:   map[string]any{"enabled": true},
			keys:   []string{"enabled"},
			want:   true,
			wantOk: true,
		},
		{
			name:   "false value",
			data:   map[string]any{"enabled": false},
			keys:   []string{"enabled"},
			want:   false,
			wantOk: true,
		},
		{
			name:   "missing key",
			data:   map[string]any{"enabled": true},
			keys:   []string{"missing"},
			want:   false,
			wantOk: false,
		},
		{
			name:   "wrong type - string",
			data:   map[string]any{"enabled": "true"},
			keys:   []string{"enabled"},
			want:   false,
			wantOk: false,
		},
		{
			name:   "wrong type - int",
			data:   map[string]any{"enabled": 1},
			keys:   []string{"enabled"},
			want:   false,
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetBool(tt.data, tt.keys...)
			if got != tt.want {
				t.Errorf("GetBool() got = %v, want %v", got, tt.want)
			}
			if gotOk != tt.wantOk {
				t.Errorf("GetBool() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}

func TestGetMap(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		keys   []string
		wantOk bool
	}{
		{
			name: "valid map",
			data: map[string]any{
				"nested": map[string]any{"key": "value"},
			},
			keys:   []string{"nested"},
			wantOk: true,
		},
		{
			name:   "missing key",
			data:   map[string]any{"other": map[string]any{}},
			keys:   []string{"nested"},
			wantOk: false,
		},
		{
			name:   "wrong type - string",
			data:   map[string]any{"nested": "not a map"},
			keys:   []string{"nested"},
			wantOk: false,
		},
		{
			name: "deeply nested",
			data: map[string]any{
				"a": map[string]any{
					"b": map[string]any{
						"c": map[string]any{"key": "value"},
					},
				},
			},
			keys:   []string{"a", "b", "c"},
			wantOk: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetMap(tt.data, tt.keys...)
			if gotOk != tt.wantOk {
				t.Errorf("GetMap() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && got == nil {
				t.Error("GetMap() returned nil map with ok=true")
			}
		})
	}
}

func TestGetArray(t *testing.T) {
	tests := []struct {
		name    string
		data    map[string]any
		keys    []string
		wantOk  bool
		wantLen int
	}{
		{
			name: "valid array",
			data: map[string]any{
				"items": []any{"a", "b", "c"},
			},
			keys:    []string{"items"},
			wantOk:  true,
			wantLen: 3,
		},
		{
			name:    "missing key",
			data:    map[string]any{"other": []any{}},
			keys:    []string{"items"},
			wantOk:  false,
			wantLen: 0,
		},
		{
			name:    "wrong type - string",
			data:    map[string]any{"items": "not an array"},
			keys:    []string{"items"},
			wantOk:  false,
			wantLen: 0,
		},
		{
			name: "empty array",
			data: map[string]any{
				"items": []any{},
			},
			keys:    []string{"items"},
			wantOk:  true,
			wantLen: 0,
		},
		{
			name: "nested array",
			data: map[string]any{
				"mediaInputSource": map[string]any{
					"supportedInputSources": map[string]any{
						"value": []any{"HDMI1", "HDMI2"},
					},
				},
			},
			keys:    []string{"mediaInputSource", "supportedInputSources", "value"},
			wantOk:  true,
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, gotOk := GetArray(tt.data, tt.keys...)
			if gotOk != tt.wantOk {
				t.Errorf("GetArray() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && len(got) != tt.wantLen {
				t.Errorf("GetArray() len = %v, want %v", len(got), tt.wantLen)
			}
		})
	}
}

func TestGetStringEquals(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		expected string
		keys     []string
		want     bool
	}{
		{
			name: "matches",
			data: map[string]any{
				"switch": map[string]any{
					"switch": map[string]any{
						"value": "on",
					},
				},
			},
			expected: "on",
			keys:     []string{"switch", "switch", "value"},
			want:     true,
		},
		{
			name: "does not match",
			data: map[string]any{
				"switch": map[string]any{
					"switch": map[string]any{
						"value": "off",
					},
				},
			},
			expected: "on",
			keys:     []string{"switch", "switch", "value"},
			want:     false,
		},
		{
			name:     "missing key",
			data:     map[string]any{},
			expected: "on",
			keys:     []string{"switch", "switch", "value"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringEquals(tt.data, tt.expected, tt.keys...)
			if got != tt.want {
				t.Errorf("GetStringEquals() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCelsiusToFahrenheit(t *testing.T) {
	tests := []struct {
		celsius    float64
		fahrenheit int
	}{
		{0, 32},
		{100, 212},
		{-40, -40},
		{20, 68},
		{37, 98}, // Body temperature (rounded)
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := CelsiusToFahrenheit(tt.celsius)
			if got != tt.fahrenheit {
				t.Errorf("CelsiusToFahrenheit(%v) = %v, want %v", tt.celsius, got, tt.fahrenheit)
			}
		})
	}
}

func TestCelsiusToFahrenheit_EdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		celsius    float64
		fahrenheit int
	}{
		{"NaN returns 0", math.NaN(), 0},
		{"positive Inf returns 0", math.Inf(1), 0},
		{"negative Inf returns 0", math.Inf(-1), 0},
		{"very large positive returns 0", 1e20, 0},
		{"very large negative returns 0", -1e20, 0},
		{"normal negative temp", -273, -459}, // Absolute zero (rounded)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CelsiusToFahrenheit(tt.celsius)
			if got != tt.fahrenheit {
				t.Errorf("CelsiusToFahrenheit(%v) = %v, want %v", tt.celsius, got, tt.fahrenheit)
			}
		})
	}
}

func TestFahrenheitToCelsius(t *testing.T) {
	tests := []struct {
		fahrenheit int
		celsius    float64
	}{
		{32, 0},
		{212, 100},
		{-40, -40},
		{68, 20},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			got := FahrenheitToCelsius(tt.fahrenheit)
			if got != tt.celsius {
				t.Errorf("FahrenheitToCelsius(%v) = %v, want %v", tt.fahrenheit, got, tt.celsius)
			}
		})
	}
}

func TestNavigate(t *testing.T) {
	tests := []struct {
		name   string
		data   map[string]any
		keys   []string
		wantOk bool
	}{
		{
			name:   "empty keys returns data",
			data:   map[string]any{"key": "value"},
			keys:   []string{},
			wantOk: true,
		},
		{
			name:   "nil data",
			data:   nil,
			keys:   []string{"key"},
			wantOk: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, gotOk := navigate(tt.data, tt.keys)
			if gotOk != tt.wantOk {
				t.Errorf("navigate() gotOk = %v, want %v", gotOk, tt.wantOk)
			}
		})
	}
}
