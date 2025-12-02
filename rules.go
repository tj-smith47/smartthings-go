package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Rule represents an automation rule.
type Rule struct {
	ID                string       `json:"id"`
	Name              string       `json:"name"`
	Actions           []RuleAction `json:"actions"`
	TimeZoneID        string       `json:"timeZoneId,omitempty"`
	OwnerID           string       `json:"ownerId,omitempty"`
	OwnerType         string       `json:"ownerType,omitempty"`
	ExecutionLocation string       `json:"executionLocation,omitempty"`
	DateCreated       string       `json:"dateCreated,omitempty"`
	DateUpdated       string       `json:"dateUpdated,omitempty"`
}

// RuleAction represents an action within a rule.
type RuleAction struct {
	If       *RuleCondition `json:"if,omitempty"`
	Sleep    *RuleSleep     `json:"sleep,omitempty"`
	Command  *RuleCommand   `json:"command,omitempty"`
	Every    *RuleEvery     `json:"every,omitempty"`
	Location *RuleLocation  `json:"location,omitempty"`
}

// RuleCondition represents a conditional in a rule.
type RuleCondition struct {
	Equals      map[string]interface{} `json:"equals,omitempty"`
	GreaterThan map[string]interface{} `json:"greaterThan,omitempty"`
	LessThan    map[string]interface{} `json:"lessThan,omitempty"`
	And         []RuleCondition        `json:"and,omitempty"`
	Or          []RuleCondition        `json:"or,omitempty"`
	Then        []RuleAction           `json:"then,omitempty"`
	Else        []RuleAction           `json:"else,omitempty"`
}

// RuleSleep represents a delay in a rule.
type RuleSleep struct {
	Duration int `json:"duration"` // seconds
}

// RuleCommand represents a command in a rule.
type RuleCommand struct {
	Devices    []RuleDeviceCommand `json:"devices,omitempty"`
	Component  string              `json:"component,omitempty"`
	Capability string              `json:"capability,omitempty"`
	Command    string              `json:"command,omitempty"`
	Arguments  []interface{}       `json:"arguments,omitempty"`
}

// RuleDeviceCommand represents a device command in a rule.
type RuleDeviceCommand struct {
	DeviceID   string        `json:"deviceId"`
	Component  string        `json:"component,omitempty"`
	Capability string        `json:"capability"`
	Command    string        `json:"command"`
	Arguments  []interface{} `json:"arguments,omitempty"`
}

// RuleEvery represents a time-based trigger.
type RuleEvery struct {
	Specific *RuleSpecificTime `json:"specific,omitempty"`
	Interval *RuleInterval     `json:"interval,omitempty"`
	Actions  []RuleAction      `json:"actions,omitempty"`
}

// RuleSpecificTime represents a specific time trigger.
type RuleSpecificTime struct {
	Reference string      `json:"reference,omitempty"` // "Sunrise", "Sunset", "Now"
	Offset    *RuleOffset `json:"offset,omitempty"`
}

// RuleOffset represents a time offset.
type RuleOffset struct {
	Value struct {
		Integer int `json:"integer,omitempty"`
	} `json:"value,omitempty"`
	Unit string `json:"unit,omitempty"` // "Minute", "Hour"
}

// RuleInterval represents a recurring interval.
type RuleInterval struct {
	Value struct {
		Integer int `json:"integer,omitempty"`
	} `json:"value,omitempty"`
	Unit string `json:"unit,omitempty"` // "Minute", "Hour", "Day"
}

// RuleLocation represents a location mode change.
type RuleLocation struct {
	Mode string `json:"mode,omitempty"`
}

// RuleCreate is the request body for creating a rule.
type RuleCreate struct {
	Name    string       `json:"name"`
	Actions []RuleAction `json:"actions"`
}

// RuleUpdate is the request body for updating a rule.
type RuleUpdate struct {
	Name    string       `json:"name,omitempty"`
	Actions []RuleAction `json:"actions,omitempty"`
}

// ruleListResponse is the API response for listing rules.
type ruleListResponse struct {
	Items []Rule `json:"items"`
}

// ListRules returns all rules for a location.
func (c *Client) ListRules(ctx context.Context, locationID string) ([]Rule, error) {
	path := "/rules"
	if locationID != "" {
		path += "?locationId=" + locationID
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp ruleListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse rule list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetRule returns a single rule by ID.
func (c *Client) GetRule(ctx context.Context, ruleID string) (*Rule, error) {
	if ruleID == "" {
		return nil, ErrEmptyRuleID
	}

	data, err := c.get(ctx, "/rules/"+ruleID)
	if err != nil {
		return nil, err
	}

	var rule Rule
	if err := json.Unmarshal(data, &rule); err != nil {
		return nil, fmt.Errorf("failed to parse rule: %w (body: %s)", err, truncatePreview(data))
	}

	return &rule, nil
}

// CreateRule creates a new rule.
func (c *Client) CreateRule(ctx context.Context, locationID string, rule *RuleCreate) (*Rule, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}
	if rule == nil || rule.Name == "" {
		return nil, ErrEmptyRuleName
	}

	data, err := c.post(ctx, "/rules?locationId="+locationID, rule)
	if err != nil {
		return nil, err
	}

	var created Rule
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created rule: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// UpdateRule updates an existing rule.
func (c *Client) UpdateRule(ctx context.Context, ruleID string, rule *RuleUpdate) (*Rule, error) {
	if ruleID == "" {
		return nil, ErrEmptyRuleID
	}

	data, err := c.put(ctx, "/rules/"+ruleID, rule)
	if err != nil {
		return nil, err
	}

	var updated Rule
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated rule: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// DeleteRule deletes a rule.
func (c *Client) DeleteRule(ctx context.Context, ruleID string) error {
	if ruleID == "" {
		return ErrEmptyRuleID
	}

	_, err := c.delete(ctx, "/rules/"+ruleID)
	return err
}

// ExecuteRule manually executes a rule.
func (c *Client) ExecuteRule(ctx context.Context, ruleID string) error {
	if ruleID == "" {
		return ErrEmptyRuleID
	}

	_, err := c.post(ctx, "/rules/execute/"+ruleID, nil)
	return err
}
