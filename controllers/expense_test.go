package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// ─── Expense Test Helpers ─────────────────────────────────────────────────────

// setupExpenseEnv creates fresh CSV files and returns a cleanup func.
func setupExpenseEnv(t *testing.T) func() {
	t.Helper()
	original, _ := os.Getwd()

	dir := t.TempDir()
	os.MkdirAll(dir+"/data", 0755)
	os.Chdir(dir)

	uf, _ := os.Create(dir + "/data/users.csv")
	uf.WriteString("id,name,email,password,created_at\n")
	// Create test users 1 and 2 (middleware validates users exist)
	uf.WriteString("1,Test User 1,test1@example.com,hashed_password,2024-01-01T00:00:00Z\n")
	uf.WriteString("2,Test User 2,test2@example.com,hashed_password,2024-01-01T00:00:00Z\n")
	uf.Close()

	ef, _ := os.Create(dir + "/data/expenses.csv")
	ef.WriteString("id,user_id,title,amount,category,note,expense_date,created_at\n")
	ef.Close()

	return func() { os.Chdir(original) }
}

// authPost sends a POST with X-User-ID header (bypasses JWT, uses your simple header auth).
func authPost(t *testing.T, path string, userID int, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	r, _ := http.NewRequest("POST", path, bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

// authGet sends a GET with X-User-ID header.
func authGet(t *testing.T, path string, userID int) *httptest.ResponseRecorder {
	t.Helper()
	r, _ := http.NewRequest("GET", path, nil)
	r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

// authPut sends a PUT with X-User-ID header.
func authPut(t *testing.T, path string, userID int, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	r, _ := http.NewRequest("PUT", path, bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

// authDelete sends a DELETE with X-User-ID header.
func authDelete(t *testing.T, path string, userID int) *httptest.ResponseRecorder {
	t.Helper()
	r, _ := http.NewRequest("DELETE", path, nil)
	r.Header.Set("X-User-ID", fmt.Sprintf("%d", userID))
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

// validExpensePayload returns a correct expense body.
func validExpensePayload() map[string]interface{} {
	return map[string]interface{}{
		"title":        "Lunch",
		"amount":       12.50,
		"category":     "Food",
		"note":         "test",
		"expense_date": "2024-06-01",
	}
}

// createExpenseForUser is a shortcut that POSTs a valid expense and returns its ID.
func createExpenseForUser(t *testing.T, userID int) int {
	t.Helper()
	w := authPost(t, "/api/v1/expenses", userID, validExpensePayload())
	if w.Code != http.StatusCreated {
		t.Fatalf("setup: expected 201, got %d — %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	data := resp["data"].(map[string]interface{})
	return int(data["id"].(float64))
}

// ─── CreateExpense ────────────────────────────────────────────────────────────

func TestCreateExpense(t *testing.T) {
	tests := []struct {
		name       string
		payloadFn  func() map[string]interface{}
		wantStatus int
	}{
		{
			name: "success returns 201",
			payloadFn: func() map[string]interface{} {
				return validExpensePayload()
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing title returns 400",
			payloadFn: func() map[string]interface{} {
				payload := validExpensePayload()
				delete(payload, "title")
				return payload
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "zero amount returns 400",
			payloadFn: func() map[string]interface{} {
				payload := validExpensePayload()
				payload["amount"] = 0
				return payload
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "negative amount returns 400",
			payloadFn: func() map[string]interface{} {
				payload := validExpensePayload()
				payload["amount"] = -5.0
				return payload
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid category returns 400",
			payloadFn: func() map[string]interface{} {
				payload := validExpensePayload()
				payload["category"] = "Gambling"
				return payload
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid date format returns 400",
			payloadFn: func() map[string]interface{} {
				payload := validExpensePayload()
				payload["expense_date"] = "01-06-2024" // wrong format
				return payload
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing date returns 400",
			payloadFn: func() map[string]interface{} {
				payload := validExpensePayload()
				delete(payload, "expense_date")
				return payload
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseEnv(t)
			defer cleanup()

			w := authPost(t, "/api/v1/expenses", 1, tt.payloadFn())
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d — %s", tt.wantStatus, w.Code, w.Body.String())
			}

			if tt.wantStatus == http.StatusCreated {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				data := resp["data"].(map[string]interface{})
				if data["id"] == nil {
					t.Error("expected id in response data")
				}
			}
		})
	}
}

// ─── ListExpenses ─────────────────────────────────────────────────────────────

func TestListExpenses(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T)
		path          string
		wantStatus    int
		checkCount    bool
		expectedCount int
	}{
		{
			name: "returns own expenses",
			setup: func(t *testing.T) {
				authPost(t, "/api/v1/expenses", 1, validExpensePayload())
				authPost(t, "/api/v1/expenses", 1, validExpensePayload())
			},
			path:          "/api/v1/expenses",
			wantStatus:    http.StatusOK,
			checkCount:    true,
			expectedCount: 2,
		},
		{
			name:       "invalid sort_by returns 400",
			path:       "/api/v1/expenses?sort_by=invalid",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid sort_order returns 400",
			path:       "/api/v1/expenses?sort_order=random",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid date_from returns 400",
			path:       "/api/v1/expenses?date_from=not-a-date",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:          "empty list returns empty array",
			path:          "/api/v1/expenses",
			wantStatus:    http.StatusOK,
			checkCount:    true,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseEnv(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			w := authGet(t, tt.path, 1)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.checkCount && tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				data := resp["data"].([]interface{})
				if len(data) != tt.expectedCount {
					t.Errorf("expected %d items, got %d", tt.expectedCount, len(data))
				}
			}
		})
	}
}

// ─── GetExpense ───────────────────────────────────────────────────────────────

func TestGetExpense(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) int // returns expense ID
		userID     int
		pathFn     func(id int) string
		wantStatus int
	}{
		{
			name: "found returns 200",
			setup: func(t *testing.T) int {
				return createExpenseForUser(t, 1)
			},
			userID: 1,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found returns 404",
			setup: func(t *testing.T) int {
				return 9999
			},
			userID: 1,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid ID returns 400",
			setup: func(t *testing.T) int {
				return 0 // not used
			},
			userID: 1,
			pathFn: func(id int) string {
				return "/api/v1/expenses/abc"
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "other users expense returns 404",
			setup: func(t *testing.T) int {
				return createExpenseForUser(t, 1)
			},
			userID: 2,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseEnv(t)
			defer cleanup()

			id := tt.setup(t)
			path := tt.pathFn(id)
			w := authGet(t, path, tt.userID)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// ─── UpdateExpense ────────────────────────────────────────────────────────────

func TestUpdateExpense(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) int
		userID     int
		pathFn     func(id int) string
		bodyFn     func() map[string]interface{}
		wantStatus int
	}{
		{
			name: "success returns 200",
			setup: func(t *testing.T) int {
				return createExpenseForUser(t, 1)
			},
			userID: 1,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			bodyFn: func() map[string]interface{} {
				payload := validExpensePayload()
				payload["title"] = "Dinner"
				payload["amount"] = 25.0
				return payload
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found returns 404",
			setup: func(t *testing.T) int {
				return 9999
			},
			userID: 1,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			bodyFn: func() map[string]interface{} {
				return validExpensePayload()
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid ID returns 400",
			setup: func(t *testing.T) int {
				return 0 // not used
			},
			userID: 1,
			pathFn: func(id int) string {
				return "/api/v1/expenses/abc"
			},
			bodyFn: func() map[string]interface{} {
				return validExpensePayload()
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseEnv(t)
			defer cleanup()

			id := tt.setup(t)
			path := tt.pathFn(id)
			w := authPut(t, path, tt.userID, tt.bodyFn())
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d — %s", tt.wantStatus, w.Code, w.Body.String())
			}
		})
	}
}

func TestUpdateExpense_InvalidBody(t *testing.T) {
	cleanup := setupExpenseEnv(t)
	defer cleanup()

	id := createExpenseForUser(t, 1)
	r, _ := http.NewRequest("PUT", fmt.Sprintf("/api/v1/expenses/%d", id), bytes.NewBufferString("{bad json"))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("X-User-ID", "1")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ─── DeleteExpense ────────────────────────────────────────────────────────────

func TestDeleteExpense(t *testing.T) {
	tests := []struct {
		name       string
		setup      func(t *testing.T) int
		userID     int
		pathFn     func(id int) string
		wantStatus int
	}{
		{
			name: "success returns 200",
			setup: func(t *testing.T) int {
				return createExpenseForUser(t, 1)
			},
			userID: 1,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "not found returns 404",
			setup: func(t *testing.T) int {
				return 9999
			},
			userID: 1,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			wantStatus: http.StatusNotFound,
		},
		{
			name: "invalid ID returns 400",
			setup: func(t *testing.T) int {
				return 0 // not used
			},
			userID: 1,
			pathFn: func(id int) string {
				return "/api/v1/expenses/abc"
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "other users expense returns 404",
			setup: func(t *testing.T) int {
				return createExpenseForUser(t, 1)
			},
			userID: 2,
			pathFn: func(id int) string {
				return fmt.Sprintf("/api/v1/expenses/%d", id)
			},
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseEnv(t)
			defer cleanup()

			id := tt.setup(t)
			path := tt.pathFn(id)
			w := authDelete(t, path, tt.userID)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}

// ─── GetSummary ───────────────────────────────────────────────────────────────

func TestGetSummary(t *testing.T) {
	tests := []struct {
		name             string
		setup            func(t *testing.T)
		path             string
		wantStatus       int
		checkTotalAmount bool
	}{
		{
			name: "returns 200",
			setup: func(t *testing.T) {
				authPost(t, "/api/v1/expenses", 1, validExpensePayload())
			},
			path:             "/api/v1/expenses/summary",
			wantStatus:       http.StatusOK,
			checkTotalAmount: true,
		},
		{
			name:       "invalid date_from returns 400",
			path:       "/api/v1/expenses/summary?date_from=bad-date",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "invalid date_to returns 400",
			path:       "/api/v1/expenses/summary?date_to=bad-date",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupExpenseEnv(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			w := authGet(t, tt.path, 1)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d", tt.wantStatus, w.Code)
			}

			if tt.checkTotalAmount && tt.wantStatus == http.StatusOK {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				data := resp["data"].(map[string]interface{})
				if data["total_amount"] == nil {
					t.Error("expected total_amount in summary response")
				}
			}
		})
	}
}
