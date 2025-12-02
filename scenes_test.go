package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListScenes(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/scenes" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/scenes")
			}
			if r.URL.Query().Get("locationId") != "loc-123" {
				t.Errorf("locationId query = %q, want %q", r.URL.Query().Get("locationId"), "loc-123")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := sceneListResponse{
				Items: []Scene{
					{SceneID: "scene-1", SceneName: "Good Morning", LocationID: "loc-123"},
					{SceneID: "scene-2", SceneName: "Good Night", LocationID: "loc-123"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		scenes, err := client.ListScenes(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(scenes) != 2 {
			t.Errorf("got %d scenes, want 2", len(scenes))
		}
		if scenes[0].SceneName != "Good Morning" {
			t.Errorf("scenes[0].SceneName = %q, want %q", scenes[0].SceneName, "Good Morning")
		}
	})

	t.Run("empty location ID returns all scenes", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Query().Get("locationId") != "" {
				t.Errorf("locationId query should be empty, got %q", r.URL.Query().Get("locationId"))
			}
			resp := sceneListResponse{
				Items: []Scene{
					{SceneID: "scene-1", SceneName: "Scene 1"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		scenes, err := client.ListScenes(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(scenes) != 1 {
			t.Errorf("got %d scenes, want 1", len(scenes))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(sceneListResponse{Items: []Scene{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		scenes, err := client.ListScenes(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(scenes) != 0 {
			t.Errorf("got %d scenes, want 0", len(scenes))
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("not json"))
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.ListScenes(context.Background(), "loc-123")
		if err == nil {
			t.Fatal("expected error for invalid JSON")
		}
	})
}

func TestClient_GetScene(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/scenes/scene-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/scenes/scene-123")
			}
			scene := Scene{
				SceneID:    "scene-123",
				SceneName:  "Movie Time",
				LocationID: "loc-456",
				SceneIcon:  "movie",
				SceneColor: "#FF0000",
			}
			json.NewEncoder(w).Encode(scene)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		scene, err := client.GetScene(context.Background(), "scene-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if scene.SceneID != "scene-123" {
			t.Errorf("SceneID = %q, want %q", scene.SceneID, "scene-123")
		}
		if scene.SceneName != "Movie Time" {
			t.Errorf("SceneName = %q, want %q", scene.SceneName, "Movie Time")
		}
	})

	t.Run("empty scene ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetScene(context.Background(), "")
		if err != ErrEmptySceneID {
			t.Errorf("expected ErrEmptySceneID, got %v", err)
		}
	})

	t.Run("scene not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetScene(context.Background(), "missing-scene")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_ExecuteScene(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/scenes/scene-123/execute" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/scenes/scene-123/execute")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.ExecuteScene(context.Background(), "scene-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty scene ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.ExecuteScene(context.Background(), "")
		if err != ErrEmptySceneID {
			t.Errorf("expected ErrEmptySceneID, got %v", err)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusUnauthorized)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.ExecuteScene(context.Background(), "scene-123")
		if !IsUnauthorized(err) {
			t.Errorf("expected unauthorized error, got %v", err)
		}
	})
}
