package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mechmanager/fiapx/api-gateway/middleware"
)

func TestRateLimiter_AllowsNormalRequests(t *testing.T) {
	limiter := middleware.NewIPRateLimiter(100)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "192.168.1.1:1234"
	rr := httptest.NewRecorder()

	limiter.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRateLimiter_UsesXForwardedFor(t *testing.T) {
	limiter := middleware.NewIPRateLimiter(100)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Forwarded-For", "10.0.0.1")
	rr := httptest.NewRecorder()

	limiter.Middleware(next).ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rr.Code)
	}
}

func TestRateLimiter_BlocksWhenLimitExceeded(t *testing.T) {
	// burst=1 significa que a 2ª requisição imediata será bloqueada
	limiter := middleware.NewIPRateLimiter(1)
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	ip := "10.0.0.99:9999"
	send := func() int {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.RemoteAddr = ip
		rr := httptest.NewRecorder()
		limiter.Middleware(next).ServeHTTP(rr, req)
		return rr.Code
	}

	// primeira passa (consome o burst)
	send()
	// segunda deve ser bloqueada
	if code := send(); code != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", code)
	}
}
