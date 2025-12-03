package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Organization represents a SmartThings organization.
type Organization struct {
	OrganizationID   string `json:"organizationId"`
	Name             string `json:"name"`
	Label            string `json:"label,omitempty"`
	ManufacturerName string `json:"manufacturerName,omitempty"`
	MNID             string `json:"mnid,omitempty"`
	WarehouseGroupID string `json:"warehouseGroupId,omitempty"`
	DeveloperGroupID string `json:"developerGroupId,omitempty"`
	AdminGroupID     string `json:"adminGroupId,omitempty"`
	IsDefaultUserOrg bool   `json:"isDefaultUserOrg,omitempty"`
}

// organizationListResponse is the API response for listing organizations.
type organizationListResponse struct {
	Items []Organization `json:"items"`
}

// ListOrganizations returns all organizations for the user.
func (c *Client) ListOrganizations(ctx context.Context) ([]Organization, error) {
	data, err := c.get(ctx, "/organizations")
	if err != nil {
		return nil, err
	}

	var resp organizationListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListOrganizations: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetOrganization returns a single organization by ID.
func (c *Client) GetOrganization(ctx context.Context, organizationID string) (*Organization, error) {
	if organizationID == "" {
		return nil, ErrEmptyOrganizationID
	}

	data, err := c.get(ctx, "/organizations/"+organizationID)
	if err != nil {
		return nil, err
	}

	var org Organization
	if err := json.Unmarshal(data, &org); err != nil {
		return nil, fmt.Errorf("GetOrganization: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &org, nil
}
