package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Room represents a room within a location.
type Room struct {
	RoomID          string `json:"roomId"`
	LocationID      string `json:"locationId"`
	Name            string `json:"name"`
	BackgroundImage string `json:"backgroundImage,omitempty"`
}

// RoomCreate is the request body for creating a room.
type RoomCreate struct {
	Name string `json:"name"`
}

// RoomUpdate is the request body for updating a room.
type RoomUpdate struct {
	Name            string `json:"name,omitempty"`
	BackgroundImage string `json:"backgroundImage,omitempty"`
}

// roomListResponse is the API response for listing rooms.
type roomListResponse struct {
	Items []Room `json:"items"`
}

// ListRooms returns all rooms in a location.
func (c *Client) ListRooms(ctx context.Context, locationID string) ([]Room, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}

	data, err := c.get(ctx, "/locations/"+locationID+"/rooms")
	if err != nil {
		return nil, err
	}

	var resp roomListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse room list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetRoom returns a single room by ID.
func (c *Client) GetRoom(ctx context.Context, locationID, roomID string) (*Room, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}
	if roomID == "" {
		return nil, ErrEmptyRoomID
	}

	data, err := c.get(ctx, "/locations/"+locationID+"/rooms/"+roomID)
	if err != nil {
		return nil, err
	}

	var room Room
	if err := json.Unmarshal(data, &room); err != nil {
		return nil, fmt.Errorf("failed to parse room: %w (body: %s)", err, truncatePreview(data))
	}

	return &room, nil
}

// CreateRoom creates a new room in a location.
func (c *Client) CreateRoom(ctx context.Context, locationID string, room *RoomCreate) (*Room, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}
	if room == nil || room.Name == "" {
		return nil, ErrEmptyRoomName
	}

	data, err := c.post(ctx, "/locations/"+locationID+"/rooms", room)
	if err != nil {
		return nil, err
	}

	var created Room
	if err := json.Unmarshal(data, &created); err != nil {
		return nil, fmt.Errorf("failed to parse created room: %w (body: %s)", err, truncatePreview(data))
	}

	return &created, nil
}

// UpdateRoom updates an existing room.
func (c *Client) UpdateRoom(ctx context.Context, locationID, roomID string, update *RoomUpdate) (*Room, error) {
	if locationID == "" {
		return nil, ErrEmptyLocationID
	}
	if roomID == "" {
		return nil, ErrEmptyRoomID
	}

	data, err := c.put(ctx, "/locations/"+locationID+"/rooms/"+roomID, update)
	if err != nil {
		return nil, err
	}

	var updated Room
	if err := json.Unmarshal(data, &updated); err != nil {
		return nil, fmt.Errorf("failed to parse updated room: %w (body: %s)", err, truncatePreview(data))
	}

	return &updated, nil
}

// DeleteRoom deletes a room.
func (c *Client) DeleteRoom(ctx context.Context, locationID, roomID string) error {
	if locationID == "" {
		return ErrEmptyLocationID
	}
	if roomID == "" {
		return ErrEmptyRoomID
	}

	_, err := c.delete(ctx, "/locations/"+locationID+"/rooms/"+roomID)
	return err
}
