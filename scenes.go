package smartthings

import (
	"context"
	"encoding/json"
	"fmt"
)

// Scene represents a SmartThings scene.
type Scene struct {
	SceneID          string `json:"sceneId"`
	SceneName        string `json:"sceneName"`
	SceneIcon        string `json:"sceneIcon,omitempty"`
	SceneColor       string `json:"sceneColor,omitempty"`
	LocationID       string `json:"locationId"`
	CreatedBy        string `json:"createdBy,omitempty"`
	CreatedDate      string `json:"createdDate,omitempty"`
	LastUpdatedDate  string `json:"lastUpdatedDate,omitempty"`
	LastExecutedDate string `json:"lastExecutedDate,omitempty"`
	Editable         bool   `json:"editable,omitempty"`
	APIOnly          bool   `json:"apiOnly,omitempty"`
}

// sceneListResponse is the API response for listing scenes.
type sceneListResponse struct {
	Items []Scene `json:"items"`
}

// ListScenes returns all scenes for a location.
// If locationID is empty, returns scenes for all locations.
func (c *Client) ListScenes(ctx context.Context, locationID string) ([]Scene, error) {
	path := "/scenes"
	if locationID != "" {
		path += "?locationId=" + locationID
	}

	data, err := c.get(ctx, path)
	if err != nil {
		return nil, err
	}

	var resp sceneListResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("failed to parse scene list: %w (body: %s)", err, truncatePreview(data))
	}

	return resp.Items, nil
}

// GetScene returns a single scene by ID.
func (c *Client) GetScene(ctx context.Context, sceneID string) (*Scene, error) {
	if sceneID == "" {
		return nil, ErrEmptySceneID
	}

	data, err := c.get(ctx, "/scenes/"+sceneID)
	if err != nil {
		return nil, err
	}

	var scene Scene
	if err := json.Unmarshal(data, &scene); err != nil {
		return nil, fmt.Errorf("failed to parse scene: %w (body: %s)", err, truncatePreview(data))
	}

	return &scene, nil
}

// ExecuteScene executes a scene.
func (c *Client) ExecuteScene(ctx context.Context, sceneID string) error {
	if sceneID == "" {
		return ErrEmptySceneID
	}

	_, err := c.post(ctx, "/scenes/"+sceneID+"/execute", nil)
	return err
}
