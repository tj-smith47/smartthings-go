package smartthings

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"sync"
	"time"
)

const (
	// DefaultBaseURL is the SmartThings API base URL.
	DefaultBaseURL = "https://api.smartthings.com/v1"

	// DefaultTimeout is the default HTTP request timeout.
	DefaultTimeout = 30 * time.Second
)

// RetryConfig configures automatic retry behavior for transient failures.
type RetryConfig struct {
	// MaxRetries is the maximum number of retry attempts (default: 3).
	MaxRetries int
	// InitialBackoff is the initial backoff duration (default: 100ms).
	InitialBackoff time.Duration
	// MaxBackoff is the maximum backoff duration (default: 5s).
	MaxBackoff time.Duration
	// Multiplier is the backoff multiplier (default: 2.0).
	Multiplier float64
}

// DefaultRetryConfig returns sensible retry defaults.
func DefaultRetryConfig() *RetryConfig {
	return &RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 100 * time.Millisecond,
		MaxBackoff:     5 * time.Second,
		Multiplier:     2.0,
	}
}

// RateLimitInfo contains rate limit information from API response headers.
type RateLimitInfo struct {
	Limit     int       // Maximum requests allowed in the window
	Remaining int       // Requests remaining in current window
	Reset     time.Time // When the rate limit window resets
}

// RateLimitCallback is called when rate limit headers are received.
// Can be used for monitoring or preemptive throttling.
type RateLimitCallback func(RateLimitInfo)

// Client is a SmartThings API client.
type Client struct {
	baseURL           string
	token             string
	httpClient        *http.Client
	retryConfig       *RetryConfig
	rateLimitCallback RateLimitCallback
	lastRateLimit     *RateLimitInfo
	rateLimitMu       sync.RWMutex
}

// Option configures a Client.
type Option func(*Client)

// WithBaseURL sets a custom base URL for the API.
func WithBaseURL(url string) Option {
	return func(c *Client) {
		c.baseURL = url
	}
}

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) {
		c.httpClient = client
	}
}

// WithTimeout sets the HTTP request timeout.
// This option can be applied in any order relative to other options.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		if c.httpClient == nil {
			c.httpClient = &http.Client{}
		}
		c.httpClient.Timeout = timeout
	}
}

// WithRetry enables automatic retry with the given configuration.
// Retries are attempted on rate limits (429), server errors (5xx), and timeouts.
func WithRetry(config *RetryConfig) Option {
	return func(c *Client) {
		c.retryConfig = config
	}
}

// WithRateLimitCallback sets a callback that is invoked when rate limit headers are received.
// This can be used for monitoring, logging, or preemptive throttling.
func WithRateLimitCallback(callback RateLimitCallback) Option {
	return func(c *Client) {
		c.rateLimitCallback = callback
	}
}

// NewClient creates a new SmartThings API client.
// Returns ErrEmptyToken if token is empty.
func NewClient(token string, opts ...Option) (*Client, error) {
	if token == "" {
		return nil, ErrEmptyToken
	}

	c := &Client{
		baseURL: DefaultBaseURL,
		token:   token,
		httpClient: &http.Client{
			Timeout: DefaultTimeout,
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				DisableKeepAlives:   false,
			},
		},
	}

	for _, opt := range opts {
		opt(c)
	}

	return c, nil
}

// do performs an HTTP request and returns the response body.
func (c *Client) do(ctx context.Context, method, path string, body any) ([]byte, error) {
	url := c.baseURL + path

	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Parse and store rate limit headers
	c.parseRateLimitHeaders(resp.Header)

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, c.handleError(resp.StatusCode, respBody)
	}

	return respBody, nil
}

// parseRateLimitHeaders extracts rate limit information from response headers.
func (c *Client) parseRateLimitHeaders(header http.Header) {
	limit := header.Get("X-RateLimit-Limit")
	remaining := header.Get("X-RateLimit-Remaining")
	reset := header.Get("X-RateLimit-Reset")

	// Only process if at least one header is present
	if limit == "" && remaining == "" && reset == "" {
		return
	}

	info := RateLimitInfo{}

	if limit != "" {
		if v, err := strconv.Atoi(limit); err == nil {
			info.Limit = v
		}
	}

	if remaining != "" {
		if v, err := strconv.Atoi(remaining); err == nil {
			info.Remaining = v
		}
	}

	if reset != "" {
		if v, err := strconv.ParseInt(reset, 10, 64); err == nil {
			info.Reset = time.Unix(v, 0)
		}
	}

	// Store the rate limit info
	c.rateLimitMu.Lock()
	c.lastRateLimit = &info
	c.rateLimitMu.Unlock()

	// Invoke callback if set
	if c.rateLimitCallback != nil {
		c.rateLimitCallback(info)
	}
}

// handleError converts HTTP error responses to appropriate errors.
func (c *Client) handleError(statusCode int, body []byte) error {
	switch statusCode {
	case http.StatusUnauthorized:
		return ErrUnauthorized
	case http.StatusNotFound:
		return ErrNotFound
	case http.StatusTooManyRequests:
		return ErrRateLimited
	case http.StatusServiceUnavailable:
		return ErrDeviceOffline
	default:
		// Try to extract error message from response
		var errResp struct {
			RequestID string `json:"requestId"`
			Error     struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		if err := json.Unmarshal(body, &errResp); err == nil && errResp.Error.Message != "" {
			return &APIError{
				StatusCode: statusCode,
				Message:    errResp.Error.Message,
				RequestID:  errResp.RequestID,
			}
		}
		return &APIError{
			StatusCode: statusCode,
			Message:    string(body),
		}
	}
}

// SetToken updates the client's bearer token.
// This is useful for OAuth clients that need to refresh tokens.
func (c *Client) SetToken(token string) {
	c.token = token
}

// Token returns the current bearer token.
func (c *Client) Token() string {
	return c.token
}

// RateLimitInfo returns the most recent rate limit information from API responses.
// Returns nil if no rate limit headers have been received yet.
func (c *Client) RateLimitInfo() *RateLimitInfo {
	c.rateLimitMu.RLock()
	defer c.rateLimitMu.RUnlock()
	if c.lastRateLimit == nil {
		return nil
	}
	// Return a copy to prevent race conditions
	info := *c.lastRateLimit
	return &info
}

// get performs a GET request.
func (c *Client) get(ctx context.Context, path string) ([]byte, error) {
	return c.doWithRetry(ctx, http.MethodGet, path, nil)
}

// post performs a POST request.
func (c *Client) post(ctx context.Context, path string, body any) ([]byte, error) {
	return c.doWithRetry(ctx, http.MethodPost, path, body)
}

// put performs a PUT request.
func (c *Client) put(ctx context.Context, path string, body any) ([]byte, error) {
	return c.doWithRetry(ctx, http.MethodPut, path, body)
}

// patch performs a PATCH request.
func (c *Client) patch(ctx context.Context, path string, body any) ([]byte, error) {
	return c.doWithRetry(ctx, http.MethodPatch, path, body)
}

// delete performs a DELETE request.
func (c *Client) delete(ctx context.Context, path string) ([]byte, error) {
	return c.doWithRetry(ctx, http.MethodDelete, path, nil)
}

// doWithRetry performs a request with automatic retry on transient failures.
func (c *Client) doWithRetry(ctx context.Context, method, path string, body any) ([]byte, error) {
	if c.retryConfig == nil {
		return c.do(ctx, method, path, body)
	}

	var lastErr error
	backoff := c.retryConfig.InitialBackoff

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		data, err := c.do(ctx, method, path, body)
		if err == nil {
			return data, nil
		}

		// Only retry on transient errors
		if !c.isRetryable(err) {
			return nil, err
		}

		lastErr = err

		if attempt < c.retryConfig.MaxRetries {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
				backoff = time.Duration(float64(backoff) * c.retryConfig.Multiplier)
				if backoff > c.retryConfig.MaxBackoff {
					backoff = c.retryConfig.MaxBackoff
				}
			}
		}
	}

	return nil, lastErr
}

// isRetryable returns true if the error is a transient failure worth retrying.
func (c *Client) isRetryable(err error) bool {
	if IsRateLimited(err) {
		return true
	}
	if IsTimeout(err) {
		return true
	}
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		// Retry on 5xx server errors
		return apiErr.StatusCode >= 500 && apiErr.StatusCode < 600
	}
	return false
}
