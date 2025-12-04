package smartthings

import (
	"context"
	"log/slog"
	"net/http"
	"time"
)

// Logger is an optional interface for structured logging.
// It uses the standard library's slog interface for compatibility.
type Logger interface {
	// LogAttrs logs a message with the given level and attributes.
	LogAttrs(ctx context.Context, level slog.Level, msg string, attrs ...slog.Attr)
}

// WithLogger configures a structured logger for the client.
// When set, the client will log API requests and responses.
//
// Example:
//
//	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
//	client, _ := st.NewClient("token", st.WithLogger(logger))
func WithLogger(logger *slog.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// LoggingTransport wraps an http.RoundTripper and logs requests/responses.
type LoggingTransport struct {
	Base   http.RoundTripper
	Logger *slog.Logger
}

// RoundTrip implements http.RoundTripper with logging.
func (t *LoggingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()

	// Log request
	if t.Logger != nil {
		t.Logger.LogAttrs(req.Context(), slog.LevelDebug, "api_request",
			slog.String("method", req.Method),
			slog.String("url", req.URL.String()),
		)
	}

	resp, err := t.Base.RoundTrip(req)
	duration := time.Since(start)

	// Log response or error
	if t.Logger != nil {
		if err != nil {
			t.Logger.LogAttrs(req.Context(), slog.LevelError, "api_error",
				slog.String("method", req.Method),
				slog.String("url", req.URL.String()),
				slog.Duration("duration", duration),
				slog.String("error", err.Error()),
			)
		} else {
			level := slog.LevelDebug
			if resp.StatusCode >= 400 {
				level = slog.LevelWarn
			}
			if resp.StatusCode >= 500 {
				level = slog.LevelError
			}

			attrs := []slog.Attr{
				slog.String("method", req.Method),
				slog.String("url", req.URL.String()),
				slog.Int("status", resp.StatusCode),
				slog.Duration("duration", duration),
			}

			// Add rate limit info if present
			if remaining := resp.Header.Get("X-RateLimit-Remaining"); remaining != "" {
				attrs = append(attrs, slog.String("rate_limit_remaining", remaining))
			}

			t.Logger.LogAttrs(req.Context(), level, "api_response", attrs...)
		}
	}

	return resp, err
}

// LogRequest logs an API request. This is the low-level logging method
// used internally and can be used for custom request logging.
func (c *Client) LogRequest(ctx context.Context, method, path string) {
	if c.logger == nil {
		return
	}
	c.logger.LogAttrs(ctx, slog.LevelDebug, "api_request",
		slog.String("method", method),
		slog.String("path", path),
	)
}

// LogResponse logs an API response. This is the low-level logging method
// used internally and can be used for custom response logging.
func (c *Client) LogResponse(ctx context.Context, method, path string, statusCode int, duration time.Duration, err error) {
	if c.logger == nil {
		return
	}

	level := slog.LevelDebug
	if statusCode >= 400 {
		level = slog.LevelWarn
	}
	if statusCode >= 500 || err != nil {
		level = slog.LevelError
	}

	attrs := []slog.Attr{
		slog.String("method", method),
		slog.String("path", path),
		slog.Int("status", statusCode),
		slog.Duration("duration", duration),
	}

	if err != nil {
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	c.logger.LogAttrs(ctx, level, "api_response", attrs...)
}

// LogRateLimit logs rate limit information at info level.
// Useful for monitoring API usage.
func (c *Client) LogRateLimit(ctx context.Context, info RateLimitInfo) {
	if c.logger == nil {
		return
	}
	c.logger.LogAttrs(ctx, slog.LevelInfo, "rate_limit",
		slog.Int("limit", info.Limit),
		slog.Int("remaining", info.Remaining),
		slog.Time("reset", info.Reset),
	)
}

// LogDeviceCommand logs a device command execution.
func (c *Client) LogDeviceCommand(ctx context.Context, deviceID string, capability, command string, err error) {
	if c.logger == nil {
		return
	}

	level := slog.LevelInfo
	msg := "device_command"
	attrs := []slog.Attr{
		slog.String("device_id", deviceID),
		slog.String("capability", capability),
		slog.String("command", command),
	}

	if err != nil {
		level = slog.LevelError
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	c.logger.LogAttrs(ctx, level, msg, attrs...)
}

// LogWebhookEvent logs an incoming webhook event.
func LogWebhookEvent(logger *slog.Logger, ctx context.Context, event *WebhookEvent, err error) {
	if logger == nil {
		return
	}

	level := slog.LevelInfo
	attrs := []slog.Attr{
		slog.String("lifecycle", string(event.Lifecycle)),
	}

	if event.AppID != "" {
		attrs = append(attrs, slog.String("app_id", event.AppID))
	}

	// LocationID is in InstallData.InstalledApp, UpdateData.InstalledApp, or EventData.InstalledApp
	if event.InstallData != nil && event.InstallData.InstalledApp.LocationID != "" {
		attrs = append(attrs, slog.String("location_id", event.InstallData.InstalledApp.LocationID))
	} else if event.UpdateData != nil && event.UpdateData.InstalledApp.LocationID != "" {
		attrs = append(attrs, slog.String("location_id", event.UpdateData.InstalledApp.LocationID))
	} else if event.EventData != nil && event.EventData.InstalledApp.LocationID != "" {
		attrs = append(attrs, slog.String("location_id", event.EventData.InstalledApp.LocationID))
	}

	if err != nil {
		level = slog.LevelError
		attrs = append(attrs, slog.String("error", err.Error()))
	}

	logger.LogAttrs(ctx, level, "webhook_event", attrs...)
}

// NewLoggingClient creates a client with request/response logging enabled.
// This is a convenience function that wraps the HTTP transport with logging.
//
// Example:
//
//	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
//	client, err := st.NewLoggingClient("token", logger)
func NewLoggingClient(token string, logger *slog.Logger, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	// Create transport with logging
	transport := &LoggingTransport{
		Base: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 10,
			IdleConnTimeout:     90 * time.Second,
			DisableKeepAlives:   false,
			ForceAttemptHTTP2:   true,
		},
		Logger: logger,
	}

	httpClient := &http.Client{
		Timeout:   DefaultTimeout,
		Transport: transport,
	}

	// Prepend WithHTTPClient and WithLogger to options
	allOpts := append([]Option{WithHTTPClient(httpClient), WithLogger(logger)}, opts...)

	return NewClient(token, allOpts...)
}
