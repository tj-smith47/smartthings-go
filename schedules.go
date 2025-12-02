package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Schedule represents a scheduled task for an installed app.
type Schedule struct {
	Name           string        `json:"name"`
	Cron           *CronSchedule `json:"cron,omitempty"`
	InstalledAppID string        `json:"installedAppId"`
	LocationID     string        `json:"locationId,omitempty"`
}

// CronSchedule represents a cron-based schedule.
type CronSchedule struct {
	Expression string `json:"expression"`
	Timezone   string `json:"timezone,omitempty"`
}

// ScheduleCreate is the request body for creating a schedule.
type ScheduleCreate struct {
	Name string        `json:"name"`
	Cron *CronSchedule `json:"cron,omitempty"`
}

// scheduleListResponse is the API response for listing schedules.
type scheduleListResponse struct {
	Items []Schedule `json:"items"`
}

// ListSchedules returns all schedules for an installed app.
func (c *Client) ListSchedules(ctx context.Context, installedAppID string) ([]Schedule, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}

	data, err := c.get(ctx, "/installedapps/"+installedAppID+"/schedules")
	if err != nil {
		return nil, err
	}

	var resp scheduleListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse schedule list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetSchedule returns a single schedule by name.
func (c *Client) GetSchedule(ctx context.Context, installedAppID, scheduleName string) (*Schedule, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}
	if scheduleName == "" {
		return nil, ErrEmptyScheduleName
	}

	data, err := c.get(ctx, "/installedapps/"+installedAppID+"/schedules/"+scheduleName)
	if err != nil {
		return nil, err
	}

	var schedule Schedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return nil, fmt.Errorf("failed to parse schedule: %w (body: %s)", err, truncatePreview(data))
	}

	return &schedule, nil
}

// CreateSchedule creates a new schedule.
func (c *Client) CreateSchedule(ctx context.Context, installedAppID string, schedule *ScheduleCreate) (*Schedule, error) {
	if installedAppID == "" {
		return nil, ErrEmptyInstalledAppID
	}
	if schedule == nil || schedule.Name == "" {
		return nil, ErrEmptyScheduleName
	}

	data, err := c.post(ctx, "/installedapps/"+installedAppID+"/schedules", schedule)
	if err != nil {
		return nil, err
	}

	var created Schedule
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created schedule: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// DeleteSchedule deletes a schedule.
func (c *Client) DeleteSchedule(ctx context.Context, installedAppID, scheduleName string) error {
	if installedAppID == "" {
		return ErrEmptyInstalledAppID
	}
	if scheduleName == "" {
		return ErrEmptyScheduleName
	}

	_, err := c.delete(ctx, "/installedapps/"+installedAppID+"/schedules/"+scheduleName)
	return err
}
