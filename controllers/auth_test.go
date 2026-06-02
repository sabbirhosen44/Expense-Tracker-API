package controllers

import (
	"bytes"
	"encoding/json"
	"expense-tracker/middlewares"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

// ─── Global Setup ─────────────────────────────────────────────────────────────

// TestMain runs once before all tests.
// Routes are registered here with proper middleware configuration.
func TestMain(m *testing.M) {
	beego.BConfig.CopyRequestBody = true
	beego.BConfig.RunMode = "test"

	// Register all routes with middleware as in routers/router.go
	ns := beego.NewNamespace("/api/v1",
		// Health Route
		beego.NSRouter("/health", &HealthController{}, "get:GetHealthStatus"),

		// Auth Route Group
		beego.NSNamespace("/auth",
			beego.NSRouter("/register", &AuthController{}, "post:Register"),
			beego.NSRouter("/login", &AuthController{}, "post:Login"),
		),

		// Expenses Route Group with middleware
		beego.NSNamespace("/expenses",
			beego.NSBefore(middlewares.AuthFilter),
			beego.NSRouter("", &ExpenseController{}, "post:CreateExpense;get:ListExpenses"),
			beego.NSRouter("/:id", &ExpenseController{}, "get:GetExpense;put:UpdateExpense;delete:DeleteExpense"),
			beego.NSRouter("/summary", &ExpenseController{}, "get:GetSummary"),
		))

	beego.AddNamespace(ns)

	os.Exit(m.Run())
}

// ─── Test Helpers ─────────────────────────────────────────────────────────────

// setupTestEnv redirects model CSV to a fresh temp dir for each test.
func setupTestEnv(t *testing.T) func() {
	t.Helper()
	original, _ := os.Getwd()

	dir := t.TempDir()
	os.MkdirAll(dir+"/data", 0755)
	os.Chdir(dir)

	f, _ := os.Create(dir + "/data/users.csv")
	f.WriteString("id,name,email,password,created_at\n")
	f.Close()

	return func() { os.Chdir(original) }
}

// post sends a JSON POST request and returns the recorded response.
func post(t *testing.T, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	raw, _ := json.Marshal(body)
	r, _ := http.NewRequest("POST", path, bytes.NewReader(raw))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

// decodeBody parses JSON response body into a map.
func decodeBody(t *testing.T, w *httptest.ResponseRecorder) map[string]interface{} {
	t.Helper()
	var result map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&result); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	return result
}

// ─── Register ─────────────────────────────────────────────────────────────────

func TestRegister(t *testing.T) {
	tests := []struct {
		name        string
		payload     map[string]string
		wantStatus  int
		preRegister bool // if true, register a user first (for duplicate test)
	}{
		{
			name: "success returns 201",
			payload: map[string]string{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secure123",
			},
			wantStatus: http.StatusCreated,
		},
		{
			name: "missing name returns 400",
			payload: map[string]string{
				"email":    "alice@example.com",
				"password": "secure123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing email returns 400",
			payload: map[string]string{
				"name":     "Alice",
				"password": "secure123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "invalid email returns 400",
			payload: map[string]string{
				"name":     "Alice",
				"email":    "not-an-email",
				"password": "secure123",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "missing password returns 400",
			payload: map[string]string{
				"name":  "Alice",
				"email": "alice@example.com",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "short password returns 400",
			payload: map[string]string{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "abc",
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "duplicate email returns 409",
			payload: map[string]string{
				"name":     "Alice",
				"email":    "alice@example.com",
				"password": "secure123",
			},
			wantStatus:  http.StatusConflict,
			preRegister: true,
		},
		{
			name: "whitespace only name returns 400",
			payload: map[string]string{
				"name":     "   ",
				"email":    "alice@example.com",
				"password": "secure123",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnv(t)
			defer cleanup()

			if tt.preRegister {
				post(t, "/api/v1/auth/register", tt.payload)
			}

			w := post(t, "/api/v1/auth/register", tt.payload)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d — body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			if tt.wantStatus == http.StatusCreated && !tt.preRegister {
				body := decodeBody(t, w)
				if body["success"] != true {
					t.Errorf("expected success=true, got: %v", body["success"])
				}
			}
		})
	}
}

func TestRegister_InvalidJSON(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	r, _ := http.NewRequest("POST", "/api/v1/auth/register", bytes.NewBufferString("{invalid json}"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin(t *testing.T) {
	tests := []struct {
		name          string
		setup         func(t *testing.T) // optional: pre-register a user
		payload       map[string]string
		wantStatus    int
		checkUserData bool // if true, validate user data in response
	}{
		{
			name: "success returns 200",
			setup: func(t *testing.T) {
				post(t, "/api/v1/auth/register", map[string]string{
					"name":     "Bob",
					"email":    "bob@example.com",
					"password": "mypassword",
				})
			},
			payload: map[string]string{
				"email":    "bob@example.com",
				"password": "mypassword",
			},
			wantStatus:    http.StatusOK,
			checkUserData: true,
		},
		{
			name: "response contains user data",
			setup: func(t *testing.T) {
				post(t, "/api/v1/auth/register", map[string]string{
					"name":     "Bob",
					"email":    "bob@example.com",
					"password": "mypassword",
				})
			},
			payload: map[string]string{
				"email":    "bob@example.com",
				"password": "mypassword",
			},
			wantStatus:    http.StatusOK,
			checkUserData: true,
		},
		{
			name: "wrong password returns 401",
			setup: func(t *testing.T) {
				post(t, "/api/v1/auth/register", map[string]string{
					"name":     "Carol",
					"email":    "carol@example.com",
					"password": "rightpass",
				})
			},
			payload: map[string]string{
				"email":    "carol@example.com",
				"password": "wrongpass",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "non-existent email returns 401",
			payload: map[string]string{
				"email":    "nobody@example.com",
				"password": "whatever",
			},
			wantStatus: http.StatusUnauthorized,
		},
		{
			name: "missing fields returns 400",
			payload: map[string]string{
				"email": "carol@example.com",
				// password intentionally missing
			},
			wantStatus: http.StatusBadRequest,
		},
		{
			name: "email is case normalised",
			setup: func(t *testing.T) {
				post(t, "/api/v1/auth/register", map[string]string{
					"name":     "Dave",
					"email":    "dave@example.com",
					"password": "mypassword",
				})
			},
			payload: map[string]string{
				"email":    "DAVE@EXAMPLE.COM", // uppercase
				"password": "mypassword",
			},
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanup := setupTestEnv(t)
			defer cleanup()

			if tt.setup != nil {
				tt.setup(t)
			}

			w := post(t, "/api/v1/auth/login", tt.payload)
			if w.Code != tt.wantStatus {
				t.Errorf("expected %d, got %d — body: %s", tt.wantStatus, w.Code, w.Body.String())
			}

			if tt.checkUserData && tt.wantStatus == http.StatusOK {
				body := decodeBody(t, w)
				if body["success"] != true {
					t.Errorf("expected success=true")
				}
				data, ok := body["data"].(map[string]interface{})
				if !ok {
					t.Fatalf("expected data object in response, got: %v", body["data"])
				}
				if data["email"] != "bob@example.com" && data["email"] != "DAVE@EXAMPLE.COM" {
					t.Errorf("expected email in data, got: %v", data["email"])
				}
				if data["user_id"] == nil {
					t.Error("expected user_id in data")
				}
			}
		})
	}
}

func TestLogin_InvalidJSON(t *testing.T) {
	cleanup := setupTestEnv(t)
	defer cleanup()

	r, _ := http.NewRequest("POST", "/api/v1/auth/login", bytes.NewBufferString("{bad json"))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
