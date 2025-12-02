package smartthings

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestClient_ListRules(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rules" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/rules")
			}
			if r.URL.Query().Get("locationId") != "loc-123" {
				t.Errorf("locationId query = %q, want %q", r.URL.Query().Get("locationId"), "loc-123")
			}
			resp := ruleListResponse{
				Items: []Rule{
					{ID: "rule-1", Name: "Turn on lights at sunset"},
					{ID: "rule-2", Name: "Lock doors at night"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rules, err := client.ListRules(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 2 {
			t.Errorf("got %d rules, want 2", len(rules))
		}
		if rules[0].Name != "Turn on lights at sunset" {
			t.Errorf("rules[0].Name = %q, want %q", rules[0].Name, "Turn on lights at sunset")
		}
	})

	t.Run("empty location ID returns all rules", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Should not have locationId query when empty
			if r.URL.Query().Get("locationId") != "" {
				t.Errorf("locationId query should be empty, got %q", r.URL.Query().Get("locationId"))
			}
			resp := ruleListResponse{
				Items: []Rule{
					{ID: "rule-1", Name: "Global Rule"},
				},
			}
			json.NewEncoder(w).Encode(resp)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rules, err := client.ListRules(context.Background(), "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 1 {
			t.Errorf("got %d rules, want 1", len(rules))
		}
	})

	t.Run("empty list", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			json.NewEncoder(w).Encode(ruleListResponse{Items: []Rule{}})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rules, err := client.ListRules(context.Background(), "loc-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(rules) != 0 {
			t.Errorf("got %d rules, want 0", len(rules))
		}
	})
}

func TestClient_GetRule(t *testing.T) {
	t.Run("successful response", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rules/rule-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/rules/rule-123")
			}
			rule := Rule{
				ID:   "rule-123",
				Name: "Motion Light",
			}
			json.NewEncoder(w).Encode(rule)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rule, err := client.GetRule(context.Background(), "rule-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.ID != "rule-123" {
			t.Errorf("ID = %q, want %q", rule.ID, "rule-123")
		}
		if rule.Name != "Motion Light" {
			t.Errorf("Name = %q, want %q", rule.Name, "Motion Light")
		}
	})

	t.Run("empty rule ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.GetRule(context.Background(), "")
		if err != ErrEmptyRuleID {
			t.Errorf("expected ErrEmptyRuleID, got %v", err)
		}
	})

	t.Run("rule not found", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		_, err := client.GetRule(context.Background(), "missing")
		if !IsNotFound(err) {
			t.Errorf("expected not found error, got %v", err)
		}
	})
}

func TestClient_CreateRule(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rules" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/rules")
			}
			if r.URL.Query().Get("locationId") != "loc-123" {
				t.Errorf("locationId query = %q, want %q", r.URL.Query().Get("locationId"), "loc-123")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}

			var req RuleCreate
			json.NewDecoder(r.Body).Decode(&req)
			if req.Name != "New Rule" {
				t.Errorf("Name = %q, want %q", req.Name, "New Rule")
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(Rule{
				ID:   "new-rule-123",
				Name: req.Name,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rule, err := client.CreateRule(context.Background(), "loc-123", &RuleCreate{
			Name: "New Rule",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.ID != "new-rule-123" {
			t.Errorf("ID = %q, want %q", rule.ID, "new-rule-123")
		}
	})

	t.Run("empty location ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateRule(context.Background(), "", &RuleCreate{Name: "Test"})
		if err != ErrEmptyLocationID {
			t.Errorf("expected ErrEmptyLocationID, got %v", err)
		}
	})

	t.Run("nil rule", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateRule(context.Background(), "loc-123", nil)
		if err != ErrEmptyRuleName {
			t.Errorf("expected ErrEmptyRuleName, got %v", err)
		}
	})

	t.Run("empty name", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.CreateRule(context.Background(), "loc-123", &RuleCreate{Name: ""})
		if err != ErrEmptyRuleName {
			t.Errorf("expected ErrEmptyRuleName, got %v", err)
		}
	})
}

func TestClient_UpdateRule(t *testing.T) {
	t.Run("successful update", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rules/rule-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/rules/rule-123")
			}
			if r.Method != http.MethodPut {
				t.Errorf("method = %q, want PUT", r.Method)
			}

			var req RuleUpdate
			json.NewDecoder(r.Body).Decode(&req)

			json.NewEncoder(w).Encode(Rule{
				ID:   "rule-123",
				Name: req.Name,
			})
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		rule, err := client.UpdateRule(context.Background(), "rule-123", &RuleUpdate{
			Name: "Updated Rule",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if rule.Name != "Updated Rule" {
			t.Errorf("Name = %q, want %q", rule.Name, "Updated Rule")
		}
	})

	t.Run("empty rule ID", func(t *testing.T) {
		client, _ := NewClient("token")
		_, err := client.UpdateRule(context.Background(), "", &RuleUpdate{Name: "Test"})
		if err != ErrEmptyRuleID {
			t.Errorf("expected ErrEmptyRuleID, got %v", err)
		}
	})
}

func TestClient_DeleteRule(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rules/rule-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/rules/rule-123")
			}
			if r.Method != http.MethodDelete {
				t.Errorf("method = %q, want DELETE", r.Method)
			}
			w.WriteHeader(http.StatusNoContent)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.DeleteRule(context.Background(), "rule-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty rule ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.DeleteRule(context.Background(), "")
		if err != ErrEmptyRuleID {
			t.Errorf("expected ErrEmptyRuleID, got %v", err)
		}
	})
}

func TestClient_ExecuteRule(t *testing.T) {
	t.Run("successful execution", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/rules/execute/rule-123" {
				t.Errorf("path = %q, want %q", r.URL.Path, "/rules/execute/rule-123")
			}
			if r.Method != http.MethodPost {
				t.Errorf("method = %q, want POST", r.Method)
			}
			w.WriteHeader(http.StatusOK)
		}))
		defer server.Close()

		client, _ := NewClient("token", WithBaseURL(server.URL))
		err := client.ExecuteRule(context.Background(), "rule-123")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("empty rule ID", func(t *testing.T) {
		client, _ := NewClient("token")
		err := client.ExecuteRule(context.Background(), "")
		if err != ErrEmptyRuleID {
			t.Errorf("expected ErrEmptyRuleID, got %v", err)
		}
	})
}
