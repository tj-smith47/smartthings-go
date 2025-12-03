package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// SchemaAppInvitation represents an invitation to use a Schema app.
type SchemaAppInvitation struct {
	ID          string `json:"id"`
	SchemaAppID string `json:"schemaAppId"`
	Description string `json:"description,omitempty"`
	Expiration  string `json:"expiration,omitempty"`
	AcceptURL   string `json:"acceptUrl,omitempty"`
	DeclineURL  string `json:"declineUrl,omitempty"`
	ShortCode   string `json:"shortCode,omitempty"`
	Acceptances int    `json:"acceptances,omitempty"`
	AcceptLimit int    `json:"acceptLimit,omitempty"`
}

// SchemaAppInvitationCreate is the request body for creating an invitation.
type SchemaAppInvitationCreate struct {
	SchemaAppID string `json:"schemaAppId"`
	Description string `json:"description,omitempty"`
	AcceptLimit int    `json:"acceptLimit,omitempty"`
}

// SchemaAppInvitationID is the response from creating an invitation.
type SchemaAppInvitationID struct {
	InvitationID string `json:"invitationId"`
}

// schemaAppInvitationListResponse is the API response for listing invitations.
type schemaAppInvitationListResponse struct {
	Items []SchemaAppInvitation `json:"items"`
}

// CreateSchemaAppInvitation creates an invitation for a Schema app.
func (c *Client) CreateSchemaAppInvitation(ctx context.Context, invitation *SchemaAppInvitationCreate) (*SchemaAppInvitationID, error) {
	if invitation == nil || invitation.SchemaAppID == "" {
		return nil, ErrEmptySchemaAppID
	}

	data, err := c.post(ctx, "/invites/schemaApp", invitation)
	if err != nil {
		return nil, err
	}

	var resp SchemaAppInvitationID
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("CreateSchemaAppInvitation: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return &resp, nil
}

// ListSchemaAppInvitations returns all invitations for a Schema app.
func (c *Client) ListSchemaAppInvitations(ctx context.Context, schemaAppID string) ([]SchemaAppInvitation, error) {
	if schemaAppID == "" {
		return nil, ErrEmptySchemaAppID
	}

	data, err := c.get(ctx, "/invites/schemaApp?schemaAppId="+schemaAppID)
	if err != nil {
		return nil, err
	}

	var resp schemaAppInvitationListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("ListSchemaAppInvitations: parse response: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// RevokeSchemaAppInvitation revokes an invitation.
func (c *Client) RevokeSchemaAppInvitation(ctx context.Context, invitationID string) error {
	if invitationID == "" {
		return ErrEmptyInvitationID
	}

	_, err := c.delete(ctx, "/invites/schemaApp/"+invitationID)
	return err
}
