package middlewares

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"expense-tracker/models"

	beego "github.com/beego/beego/v2/server/web"
)

// ─── Setup ────────────────────────────────────────────────────────────────────

func TestMain(m *testing.M) {
	beego.BConfig.CopyRequestBody = true
	beego.BConfig.RunMode = "test"

	// A minimal protected route — 200 if middleware passes, 401 if blocked
	beego.InsertFilter("/api/v1/protected", beego.BeforeRouter, AuthFilter)
	beego.Router("/api/v1/protected", &dummyController{}, "get:Get")

	os.Exit(m.Run())
}

// dummyController returns 200 so we can verify the middleware let the request through.
type dummyController struct {
	beego.Controller
}

func (c *dummyController) Get() {
	c.Ctx.Output.SetStatus(http.StatusOK)
	c.Data["json"] = map[string]bool{"ok": true}
	c.ServeJSON()
}

// setupMiddlewareEnv creates fresh CSV files and one real user.
func setupMiddlewareEnv(t *testing.T) (userID int, cleanup func()) {
	t.Helper()
	original, _ := os.Getwd()

	dir := t.TempDir()
	os.MkdirAll(dir+"/data", 0755)
	os.Chdir(dir)

	uf, _ := os.Create(dir + "/data/users.csv")
	uf.WriteString("id,name,email,password,created_at\n")
	uf.Close()

	ef, _ := os.Create(dir + "/data/expenses.csv")
	ef.WriteString("id,user_id,title,amount,category,note,expense_date,created_at\n")
	ef.Close()

	hash, _ := models.HashPassword("pass123")
	u := &models.User{Name: "Test", Email: "test@test.com", Password: hash}
	models.CreateUser(u)

	return u.ID, func() { os.Chdir(original) }
}

// get sends a GET to the protected route with an optional X-User-ID header value.
func get(userIDHeader string) *httptest.ResponseRecorder {
	r, _ := http.NewRequest("GET", "/api/v1/protected", nil)
	if userIDHeader != "" {
		r.Header.Set("X-User-ID", userIDHeader)
	}
	w := httptest.NewRecorder()
	beego.BeeApp.Handlers.ServeHTTP(w, r)
	return w
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestAuthFilter_ValidUserID_PassesThrough(t *testing.T) {
	userID, cleanup := setupMiddlewareEnv(t)
	defer cleanup()

	w := get(fmt.Sprint(userID))
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d — %s", w.Code, w.Body.String())
	}
}

func TestAuthFilter_MissingHeader_Returns401(t *testing.T) {
	_, cleanup := setupMiddlewareEnv(t)
	defer cleanup()

	w := get("") // no X-User-ID header at all
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthFilter_ZeroUserID_Returns401(t *testing.T) {
	_, cleanup := setupMiddlewareEnv(t)
	defer cleanup()

	w := get("0")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthFilter_NonExistentUserID_Returns401(t *testing.T) {
	_, cleanup := setupMiddlewareEnv(t)
	defer cleanup()

	w := get("9999") // valid number but no such user in CSV
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthFilter_NonNumericHeader_Returns401(t *testing.T) {
	_, cleanup := setupMiddlewareEnv(t)
	defer cleanup()

	w := get("not-a-number")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestAuthFilter_NegativeUserID_Returns401(t *testing.T) {
	_, cleanup := setupMiddlewareEnv(t)
	defer cleanup()

	w := get("-1")
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}
