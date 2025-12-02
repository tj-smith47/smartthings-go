package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListRooms(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/rooms" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/rooms")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := roomListResponse{
				Items: []Room{
					{RoomID: "room-1", Name: "Living Room", LocationID: "loc-123"},
					{RoomID: "room-2", Name: "Kitchen", LocationID: "loc-123"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rooms, err := client.ListRooms(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rooms) != 2 {
			t.Errorf("got %d rooms, want 2", len(rooms))
		}
		if rooms[0].Name != "Living Room" {
			t.Errorf("rooms[0].Name = %q, want %q", rooms[0].Name, "Living Room")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.ListRooms(context.Background(), "")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(roomListResponse{Items: []Room{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rooms, err := client.ListRooms(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rooms) != 0 {
			t.Errorf("got %d rooms, want 0", len(rooms))
		}
	})
}

func TestClient_GetRoom(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/rooms/room-456" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/rooms/room-456")
			}
			room := Room{
				RoomID:          "room-456",
				LocationID:      "loc-123",
				Name:            "Bedroom",
				BackgroundImage: "https://example.com/bg.jpg",
			}
			json.NewEncoder(w).Encode(room)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		room, err := client.GetRoom(context.Background(), "loc-123", "room-456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if room.RoomID != "room-456" {
			t.Errorf("RoomID = %q, want %q", room.RoomID, "room-456")
		}
		if room.Name != "Bedroom" {
			t.Errorf("Name = %q, want %q", room.Name, "Bedroom")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetRoom(context.Background(), "", "room-456")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("empty room ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetRoom(context.Background(), "loc-123", "")
		if err != ErrEmptyRoomID {
			t.Errorf("expected ErrEmptyRoomID, got %v", err)
		}
	})

	t.Run("room not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetRoom(context.Background(), "loc-123", "missing")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_CreateRoom(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/rooms" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/rooms")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req RoomCreate
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "New Room" {
				t.Errorf("Name = %q, want %q", req.Name, "New Room")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Room{
				RoomID:     "new-room-123",
				LocationID: "loc-123",
				Name:       req.Name,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		room, err := client.CreateRoom(context.Background(), "loc-123", &RoomCreate{
			Name: "New Room",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if room.RoomID != "new-room-123" {
			t.Errorf("RoomID = %q, want %q", room.RoomID, "new-room-123")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateRoom(context.Background(), "", &RoomCreate{Name: "Test"})
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("nil room", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateRoom(context.Background(), "loc-123", nil)
		if err != ErrEmptyRoomName {
			t.Errorf("expected ErrEmptyRoomName, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateRoom(context.Background(), "loc-123", &RoomCreate{Name: ""})
		if err != ErrEmptyRoomName {
			t.Errorf("expected ErrEmptyRoomName, got %v", err)
		}
	})
}

func TestClient_UpdateRoom(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/rooms/room-456" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/rooms/room-456")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req RoomUpdate
			json.NewDecoder(r.Body).Decode(&req)

			json.NewEncoder(w).Encode(Room{
				RoomID:     "room-456",
				LocationID: "loc-123",
				Name:       req.Name,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		room, err := client.UpdateRoom(context.Background(), "loc-123", "room-456", &RoomUpdate{
			Name: "Updated Room",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if room.Name != "Updated Room" {
			t.Errorf("Name = %q, want %q", room.Name, "Updated Room")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateRoom(context.Background(), "", "room-456", &RoomUpdate{Name: "Test"})
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("empty room ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateRoom(context.Background(), "loc-123", "", &RoomUpdate{Name: "Test"})
		if err != ErrEmptyRoomID {
			t.Errorf("expected ErrEmptyRoomID, got %v", err)
		}
	})
}

func TestClient_DeleteRoom(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/locations/loc-123/rooms/room-456" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/locations/loc-123/rooms/room-456")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteRoom(context.Background(), "loc-123", "room-456")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteRoom(context.Background(), "", "room-456")
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("empty room ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteRoom(context.Background(), "loc-123", "")
		if err != ErrEmptyRoomID {
			t.Errorf("expected ErrEmptyRoomID, got %v", err)
		}
	})
}
