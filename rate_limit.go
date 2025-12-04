package smartthings

import (
	"context"
	"strconv"
	"time"
)

// RateLimitError provides detailed information about a rate limit response.
// It includes the recommended wait time from the Retry-After header if available.
type RateLimitError struct {
	// RetryAfter is the recommended wait duration from the Retry-After header.
	// Zero if the header was not present.
	RetryAfter time.Duration

	// Info contains the rate limit headers from the response.
	Info *RateLimitInfo
}

// Error implements the error interface.
func (e *RateLimitError) Error() string {
	if e.RetryAfter > 0 {
		return "smartthings: rate limited (retry after " + e.RetryAfter.String() + ")"
	}
	return "smartthings: rate limited"
}

// Is allows errors.Is() to match ErrRateLimited.
func (e *RateLimitError) Is(target error) bool {
	return target == ErrRateLimited
}

// parseRetryAfter parses the Retry-After header value.
// It handles both delta-seconds (e.g., "120") and HTTP-date formats.
func parseRetryAfter(value string) time.Duration {
	if value == "" {
		return 0
	}

	// Try parsing as seconds first (most common)
	if seconds, err := strconv.Atoi(value); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	// Try parsing as HTTP-date (RFC 1123)
	if t, err := time.Parse(time.RFC1123, value); err == nil {
		delta := time.Until(t)
		if delta > 0 {
			return delta
		}
	}

	return 0
}

// WaitForRateLimit blocks until the rate limit window resets.
// It returns immediately if no rate limit info is available or if the reset time has passed.
// The context can be used to cancel the wait.
//
// Example:
//
//	for {
//	    err := client.ExecuteCommand(ctx, deviceID, cmd)
//	    if errors.Is(err, ErrRateLimited) {
//	        if err := client.WaitForRateLimit(ctx); err != nil {
//	            return err // Context canceled
//	        }
//	        continue // Retry
//	    }
//	    break
//	}
func (c *Client) WaitForRateLimit(ctx context.Context) error {
	info := c.RateLimitInfo()
	if info == nil {
		return nil
	}

	// Calculate wait time
	waitDuration := time.Until(info.Reset)
	if waitDuration <= 0 {
		return nil
	}

	// Cap wait at reasonable maximum (5 minutes)
	if waitDuration > 5*time.Minute {
		waitDuration = 5 * time.Minute
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDuration):
		return nil
	}
}

// WaitForRateLimitErr waits based on a RateLimitError's RetryAfter duration.
// If the error is not a RateLimitError, it returns immediately.
//
// Example:
//
//	err := client.ExecuteCommand(ctx, deviceID, cmd)
//	if err != nil {
//	    if waitErr := client.WaitForRateLimitErr(ctx, err); waitErr != nil {
//	        return waitErr // Context canceled
//	    }
//	    // Retry the command
//	}
func (c *Client) WaitForRateLimitErr(ctx context.Context, err error) error {
	rle, ok := err.(*RateLimitError)
	if !ok {
		return nil
	}

	waitDuration := rle.RetryAfter
	if waitDuration <= 0 {
		// Fall back to rate limit info
		return c.WaitForRateLimit(ctx)
	}

	// Cap wait at reasonable maximum (5 minutes)
	if waitDuration > 5*time.Minute {
		waitDuration = 5 * time.Minute
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(waitDuration):
		return nil
	}
}

// ShouldThrottle returns true if the rate limit remaining is below the given threshold.
// This enables preemptive throttling before hitting rate limits.
//
// Example:
//
//	if client.ShouldThrottle(10) {
//	    // Remaining requests below threshold, slow down
//	    time.Sleep(time.Second)
//	}
func (c *Client) ShouldThrottle(threshold int) bool {
	info := c.RateLimitInfo()
	if info == nil {
		return false
	}
	return info.Remaining < threshold && info.Remaining >= 0
}

// RemainingRequests returns the number of remaining requests in the current rate limit window.
// Returns -1 if no rate limit info is available.
func (c *Client) RemainingRequests() int {
	info := c.RateLimitInfo()
	if info == nil {
		return -1
	}
	return info.Remaining
}

// RateLimitResetTime returns the time when the rate limit window resets.
// Returns zero time if no rate limit info is available.
func (c *Client) RateLimitResetTime() time.Time {
	info := c.RateLimitInfo()
	if info == nil {
		return time.Time{}
	}
	return info.Reset
}

// RateLimitThrottler provides automatic rate limiting for bulk operations.
// It tracks API calls and automatically throttles when approaching limits.
type RateLimitThrottler struct {
	client    *Client
	threshold int           // Start throttling when remaining < threshold
	delay     time.Duration // Delay between calls when throttling
}

// NewRateLimitThrottler creates a throttler that monitors rate limits
// and automatically slows down when approaching the limit.
//
// Example:
//
//	throttler := NewRateLimitThrottler(client, 10, 100*time.Millisecond)
//	for _, deviceID := range devices {
//	    throttler.Wait(ctx) // Automatically waits if needed
//	    client.ExecuteCommand(ctx, deviceID, cmd)
//	}
func NewRateLimitThrottler(client *Client, threshold int, delay time.Duration) *RateLimitThrottler {
	return &RateLimitThrottler{
		client:    client,
		threshold: threshold,
		delay:     delay,
	}
}

// Wait blocks if the rate limit is approaching the threshold.
// Returns immediately if plenty of requests remain.
func (t *RateLimitThrottler) Wait(ctx context.Context) error {
	if !t.client.ShouldThrottle(t.threshold) {
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(t.delay):
		return nil
	}
}

// WaitUntilReset blocks until the rate limit window resets.
// Useful when you've hit the rate limit and want to wait for a full reset.
func (t *RateLimitThrottler) WaitUntilReset(ctx context.Context) error {
	return t.client.WaitForRateLimit(ctx)
}
