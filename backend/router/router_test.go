package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestSetupRoutesRegistersHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSetupRoutesRegistersProtectedDashboardStatsEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSetupRoutesRegistersProtectedCategoryUpdateEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r)

	req := httptest.NewRequest(http.MethodPut, "/api/categories/1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSetupRoutesRegistersProtectedCSDNPreviewEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/articles/import/csdn/preview", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body=%s", w.Code, w.Body.String())
	}
}

func TestSetupRoutesRegistersProtectedCSDNImportEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	SetupRoutes(r)

	req := httptest.NewRequest(http.MethodPost, "/api/articles/import/csdn", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d, body=%s", w.Code, w.Body.String())
	}
}
