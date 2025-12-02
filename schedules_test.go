package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListSchedules(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/schedules" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/schedules")
			}
			if r.Method != http.MethodGet {
				t.Errorf("method = %q, want GET", r.Method)
			}
			resp := scheduleListResponse{
				Items: []Schedule{
					{Name: "daily-check", InstalledAppID: "app-123"},
					{Name: "weekly-report", InstalledAppID: "app-123"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		schedules, err := client.ListSchedules(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(schedules) != 2 {
			t.Errorf("got %d schedules, want 2", len(schedules))
		}
		if schedules[0].Name != "daily-check" {
			t.Errorf("schedules[0].Name = %q, want %q", schedules[0].Name, "daily-check")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.ListSchedules(context.Background(), "")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(scheduleListResponse{Items: []Schedule{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		schedules, err := client.ListSchedules(context.Background(), "app-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(schedules) != 0 {
			t.Errorf("got %d schedules, want 0", len(schedules))
		}
	})
}

func TestClient_GetSchedule(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/schedules/daily-check" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/schedules/daily-check")
			}
			schedule := Schedule{
				Name:           "daily-check",
				InstalledAppID: "app-123",
				Cron: &CronSchedule{
					Expression: "0 8 * * *",
					Timezone:   "America/New_York",
				},
			}
			json.NewEncoder(w).Encode(schedule)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		schedule, err := client.GetSchedule(context.Background(), "app-123", "daily-check")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if schedule.Name != "daily-check" {
			t.Errorf("Name = %q, want %q", schedule.Name, "daily-check")
		}
		if schedule.Cron == nil {
			t.Fatal("Cron is nil")
		}
		if schedule.Cron.Expression != "0 8 * * *" {
			t.Errorf("Cron.Expression = %q, want %q", schedule.Cron.Expression, "0 8 * * *")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetSchedule(context.Background(), "", "daily-check")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("empty schedule name", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetSchedule(context.Background(), "app-123", "")
		if err != ErrEmptyScheduleName {
			t.Errorf("expected ErrEmptyScheduleName, got %v", err)
		}
	})

	t.Run("schedule not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetSchedule(context.Background(), "app-123", "missing")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_CreateSchedule(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/schedules" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/schedules")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req ScheduleCreate
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "new-schedule" {
				t.Errorf("Name = %q, want %q", req.Name, "new-schedule")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Schedule{
				Name:           req.Name,
				InstalledAppID: "app-123",
				Cron:           req.Cron,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		schedule, err := client.CreateSchedule(context.Background(), "app-123", &ScheduleCreate{
			Name: "new-schedule",
			Cron: &CronSchedule{
				Expression: "0 9 * * 1",
				Timezone:   "UTC",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if schedule.Name != "new-schedule" {
			t.Errorf("Name = %q, want %q", schedule.Name, "new-schedule")
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateSchedule(context.Background(), "", &ScheduleCreate{Name: "test"})
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("nil schedule", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateSchedule(context.Background(), "app-123", nil)
		if err != ErrEmptyScheduleName {
			t.Errorf("expected ErrEmptyScheduleName, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateSchedule(context.Background(), "app-123", &ScheduleCreate{Name: ""})
		if err != ErrEmptyScheduleName {
			t.Errorf("expected ErrEmptyScheduleName, got %v", err)
		}
	})
}

func TestClient_DeleteSchedule(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/installedapps/app-123/schedules/daily-check" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/installedapps/app-123/schedules/daily-check")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteSchedule(context.Background(), "app-123", "daily-check")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty installed app ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteSchedule(context.Background(), "", "daily-check")
		if err != ErrEmptyInstalledAppID {
			t.Errorf("expected ErrEmptyInstalledAppID, got %v", err)
		}
	})

	t.Run("empty schedule name", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteSchedule(context.Background(), "app-123", "")
		if err != ErrEmptyScheduleName {
			t.Errorf("expected ErrEmptyScheduleName, got %v", err)
		}
	})
}
