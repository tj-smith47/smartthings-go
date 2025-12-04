package smartthings

import (
	"context"
	"sync"
)

// BatchCommand represents a command to be executed on a specific device.
type BatchCommand struct {
	DeviceID string    // Target device ID
	Commands []Command // Commands to execute on this device
}

// BatchResult contains the result of executing commands on a single device.
type BatchResult struct {
	DeviceID string // The device ID
	Error    error  // Error if execution failed, nil on success
}

// BatchConfig configures batch execution behavior.
type BatchConfig struct {
	// MaxConcurrent is the maximum number of concurrent API calls.
	// Defaults to 10 if not specified.
	MaxConcurrent int

	// StopOnError determines whether to stop processing remaining commands
	// when an error occurs. Default is false (continue processing all).
	StopOnError bool
}

// DefaultBatchConfig returns sensible defaults for batch operations.
func DefaultBatchConfig() *BatchConfig {
	return &BatchConfig{
		MaxConcurrent: 10,
		StopOnError:   false,
	}
}

// ExecuteCommandsBatch executes commands on multiple devices concurrently.
// It uses a worker pool to limit concurrent API calls and respects rate limits.
//
// Example:
//
//	batch := []BatchCommand{
//	    {DeviceID: "device1", Commands: []Command{{Component: "main", Capability: "switch", Command: "on"}}},
//	    {DeviceID: "device2", Commands: []Command{{Component: "main", Capability: "switch", Command: "off"}}},
//	}
//	results := client.ExecuteCommandsBatch(ctx, batch, nil)
//	for _, r := range results {
//	    if r.Error != nil {
//	        log.Printf("Device %s failed: %v", r.DeviceID, r.Error)
//	    }
//	}
func (c *Client) ExecuteCommandsBatch(ctx context.Context, batch []BatchCommand, cfg *BatchConfig) []BatchResult {
	if len(batch) == 0 {
		return nil
	}

	if cfg == nil {
		cfg = DefaultBatchConfig()
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 10
	}

	results := make([]BatchResult, len(batch))
	var mu sync.Mutex
	var stopped bool

	// Worker pool using semaphore pattern
	sem := make(chan struct{}, cfg.MaxConcurrent)
	var wg sync.WaitGroup

	for i, cmd := range batch {
		// Check if we should stop
		mu.Lock()
		if stopped {
			mu.Unlock()
			results[i] = BatchResult{DeviceID: cmd.DeviceID, Error: context.Canceled}
			continue
		}
		mu.Unlock()

		// Check context
		select {
		case <-ctx.Done():
			results[i] = BatchResult{DeviceID: cmd.DeviceID, Error: ctx.Err()}
			continue
		default:
		}

		wg.Add(1)
		go func(idx int, bc BatchCommand) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = BatchResult{DeviceID: bc.DeviceID, Error: ctx.Err()}
				return
			}

			// Check if stopped
			mu.Lock()
			if stopped {
				mu.Unlock()
				results[idx] = BatchResult{DeviceID: bc.DeviceID, Error: context.Canceled}
				return
			}
			mu.Unlock()

			// Execute commands
			err := c.ExecuteCommands(ctx, bc.DeviceID, bc.Commands)
			results[idx] = BatchResult{DeviceID: bc.DeviceID, Error: err}

			// Handle stop on error
			if err != nil && cfg.StopOnError {
				mu.Lock()
				stopped = true
				mu.Unlock()
			}
		}(i, cmd)
	}

	wg.Wait()
	return results
}

// ExecuteCommandBatch is a convenience wrapper for executing the same command
// on multiple devices.
//
// Example:
//
//	cmd := Command{Component: "main", Capability: "switch", Command: "on"}
//	results := client.ExecuteCommandBatch(ctx, []string{"device1", "device2", "device3"}, cmd, nil)
func (c *Client) ExecuteCommandBatch(ctx context.Context, deviceIDs []string, cmd Command, cfg *BatchConfig) []BatchResult {
	batch := make([]BatchCommand, len(deviceIDs))
	for i, id := range deviceIDs {
		batch[i] = BatchCommand{
			DeviceID: id,
			Commands: []Command{cmd},
		}
	}
	return c.ExecuteCommandsBatch(ctx, batch, cfg)
}

// BatchStatusResult contains device status fetch results.
type BatchStatusResult struct {
	DeviceID   string            // The device ID
	Components map[string]Status // Status per component (nil on error)
	Error      error             // Error if fetch failed
}

// GetDeviceStatusBatch fetches status for multiple devices concurrently.
//
// Example:
//
//	results := client.GetDeviceStatusBatch(ctx, []string{"device1", "device2"}, nil)
//	for _, r := range results {
//	    if r.Error == nil {
//	        fmt.Printf("Device %s: %v\n", r.DeviceID, r.Components)
//	    }
//	}
func (c *Client) GetDeviceStatusBatch(ctx context.Context, deviceIDs []string, cfg *BatchConfig) []BatchStatusResult {
	if len(deviceIDs) == 0 {
		return nil
	}

	if cfg == nil {
		cfg = DefaultBatchConfig()
	}
	if cfg.MaxConcurrent <= 0 {
		cfg.MaxConcurrent = 10
	}

	results := make([]BatchStatusResult, len(deviceIDs))

	// Worker pool using semaphore pattern
	sem := make(chan struct{}, cfg.MaxConcurrent)
	var wg sync.WaitGroup

	for i, deviceID := range deviceIDs {
		// Check context
		select {
		case <-ctx.Done():
			results[i] = BatchStatusResult{DeviceID: deviceID, Error: ctx.Err()}
			continue
		default:
		}

		wg.Add(1)
		go func(idx int, id string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				results[idx] = BatchStatusResult{DeviceID: id, Error: ctx.Err()}
				return
			}

			// Fetch status
			status, err := c.GetDeviceFullStatus(ctx, id)
			results[idx] = BatchStatusResult{
				DeviceID:   id,
				Components: status,
				Error:      err,
			}
		}(i, deviceID)
	}

	wg.Wait()
	return results
}
