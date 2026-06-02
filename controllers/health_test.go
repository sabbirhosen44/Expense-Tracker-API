package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beego/beego/v2/server/web"
)

func TestHealthController_GetHealthStatus(t *testing.T) {
	// Register the specific route as defined in your routers/init.go
	web.Router("/api/v1/health", &HealthController{}, "get:GetHealthStatus")

	req, _ := http.NewRequest("GET", "/api/v1/health", nil)
	w := httptest.NewRecorder()

	web.BeeApp.Handlers.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("failed to decode JSON: %v", err)
	}

	if response["message"] != "Server is running!" {
		t.Errorf("expected 'Server is running!', got %v", response["message"])
	}
}
