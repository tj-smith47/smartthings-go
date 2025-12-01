package smartthings

import (
	"errors"
	"testing"
)

func TestAPIError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *APIError
		wantMsg string
	}{
		{
			name: "with request ID",
			err: &APIError{
				StatusCode: 500,
				Message:    "Internal server error",
				RequestID:  "abc123",
			},
			wantMsg: "smartthings: API error 500: Internal server error (request_id: abc123)",
		},
		{
			name: "without request ID",
			err: &APIError{
				StatusCode: 400,
				Message:    "Bad request",
			},
			wantMsg: "smartthings: API error 400: Bad request",
		},
		{
			name: "empty message",
			err: &APIError{
				StatusCode: 503,
			},
			wantMsg: "smartthings: API error 503: ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.wantMsg {
				t.Errorf("APIError.Error() = %q, want %q", got, tt.wantMsg)
			}
		})
	}
}

func TestIsUnauthorized(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrUnauthorized",
			err:  ErrUnauthorized,
			want: true,
		},
		{
			name: "wrapped ErrUnauthorized",
			err:  errors.Join(errors.New("context"), ErrUnauthorized),
			want: true,
		},
		{
			name: "APIError with 401",
			err:  &APIError{StatusCode: 401, Message: "token expired"},
			want: true,
		},
		{
			name: "APIError with other status",
			err:  &APIError{StatusCode: 403, Message: "forbidden"},
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "ErrNotFound",
			err:  ErrNotFound,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsUnauthorized(tt.err)
			if got != tt.want {
				t.Errorf("IsUnauthorized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrNotFound",
			err:  ErrNotFound,
			want: true,
		},
		{
			name: "wrapped ErrNotFound",
			err:  errors.Join(errors.New("context"), ErrNotFound),
			want: true,
		},
		{
			name: "APIError with 404",
			err:  &APIError{StatusCode: 404, Message: "device not found"},
			want: true,
		},
		{
			name: "APIError with other status",
			err:  &APIError{StatusCode: 500, Message: "internal error"},
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
		{
			name: "ErrUnauthorized",
			err:  ErrUnauthorized,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsNotFound(tt.err)
			if got != tt.want {
				t.Errorf("IsNotFound() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsRateLimited(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{
			name: "ErrRateLimited",
			err:  ErrRateLimited,
			want: true,
		},
		{
			name: "wrapped ErrRateLimited",
			err:  errors.Join(errors.New("context"), ErrRateLimited),
			want: true,
		},
		{
			name: "APIError with 429",
			err:  &APIError{StatusCode: 429, Message: "too many requests"},
			want: true,
		},
		{
			name: "APIError with other status",
			err:  &APIError{StatusCode: 503, Message: "service unavailable"},
			want: false,
		},
		{
			name: "other error",
			err:  errors.New("some error"),
			want: false,
		},
		{
			name: "nil error",
			err:  nil,
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsRateLimited(tt.err)
			if got != tt.want {
				t.Errorf("IsRateLimited() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	// Test that error constants are distinct
	if errors.Is(ErrUnauthorized, ErrNotFound) {
		t.Error("ErrUnauthorized should not match ErrNotFound")
	}
	if errors.Is(ErrUnauthorized, ErrRateLimited) {
		t.Error("ErrUnauthorized should not match ErrRateLimited")
	}
	if errors.Is(ErrNotFound, ErrRateLimited) {
		t.Error("ErrNotFound should not match ErrRateLimited")
	}
	if errors.Is(ErrUnauthorized, ErrDeviceOffline) {
		t.Error("ErrUnauthorized should not match ErrDeviceOffline")
	}

	// Test error messages
	if ErrUnauthorized.Error() == "" {
		t.Error("ErrUnauthorized should have a message")
	}
	if ErrNotFound.Error() == "" {
		t.Error("ErrNotFound should have a message")
	}
	if ErrRateLimited.Error() == "" {
		t.Error("ErrRateLimited should have a message")
	}
	if ErrDeviceOffline.Error() == "" {
		t.Error("ErrDeviceOffline should have a message")
	}
}
