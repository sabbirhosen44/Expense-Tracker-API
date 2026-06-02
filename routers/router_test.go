package routers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	beego "github.com/beego/beego/v2/server/web"
)

func TestHealthRouteRegistered(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("health route is not registered")
	}
}

func TestRegisterRouteRegistered(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/v1/auth/register", nil)
	w := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("register route is not registered")
	}
}

func TestLoginRouteRegistered(t *testing.T) {
	req, _ := http.NewRequest("POST", "/api/v1/auth/login", nil)
	w := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("login route is not registered")
	}
}

func TestExpensesRouteRegistered(t *testing.T) {
	req, _ := http.NewRequest("GET", "/api/v1/expenses", nil)
	w := httptest.NewRecorder()

	beego.BeeApp.Handlers.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Fatal("expenses route is not registered")
	}
}
