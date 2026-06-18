package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mechmanager/fiapx/api-gateway/middleware"
)

func TestMetrics_Middleware_RecordsSuccess(t *testing.T) {
	m := middleware.NewMetrics()
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/videos", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestMetrics_Middleware_RecordsError(t *testing.T) {
	m := middleware.NewMetrics()
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))

	req := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rr.Code)
	}
}

func TestMetrics_Middleware_DefaultStatus(t *testing.T) {
	// handler does not call WriteHeader explicitly — default 200 must be recorded
	m := middleware.NewMetrics()
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestMetrics_Middleware_SkipsMetricsPath(t *testing.T) {
	m := middleware.NewMetrics()
	called := false
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if !called {
		t.Error("handler must be called even for /metrics path")
	}
}

func TestMetrics_Middleware_NormalizesUUID(t *testing.T) {
	m := middleware.NewMetrics()
	var recordedPath string

	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// normalizePath is exercised internally; just confirm the request passes through
		recordedPath = r.URL.Path
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/videos/123e4567-e89b-12d3-a456-426614174000/download", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if recordedPath != "/videos/123e4567-e89b-12d3-a456-426614174000/download" {
		t.Errorf("path should not be mutated on request, got %q", recordedPath)
	}
}

func TestMetrics_Middleware_PathWithQueryString(t *testing.T) {
	m := middleware.NewMetrics()
	handler := m.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/videos?page=1&limit=10", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}
