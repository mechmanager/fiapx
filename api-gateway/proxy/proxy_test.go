package proxy_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/mechmanager/fiapx/api-gateway/proxy"
)

func TestNew_ValidURL(t *testing.T) {
	p, err := proxy.New("http://auth-service:8081")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil proxy")
	}
}

func TestNew_InvalidURL(t *testing.T) {
	_, err := proxy.New("://invalid-url")
	if err == nil {
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestNew_EmptyURL(t *testing.T) {
	p, err := proxy.New("")
	if err != nil {
		t.Fatalf("empty URL parses fine: %v", err)
	}
	if p == nil {
		t.Fatal("expected non-nil proxy")
	}
}

func TestNew_DirectorSetsSchemeAndHost(t *testing.T) {
	p, err := proxy.New("http://backend:8081")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, "/some/path", nil)
	req.Header.Set("X-Forwarded-Host", "original-host")
	p.Director(req)

	if req.URL.Scheme != "http" {
		t.Errorf("URL.Scheme = %q, want %q", req.URL.Scheme, "http")
	}
	if req.URL.Host != "backend:8081" {
		t.Errorf("URL.Host = %q, want %q", req.URL.Host, "backend:8081")
	}
	if req.Host != "backend:8081" {
		t.Errorf("Host = %q, want %q", req.Host, "backend:8081")
	}
	if req.Header.Get("X-Forwarded-Host") != "" {
		t.Error("X-Forwarded-Host should be deleted by Director")
	}
}

func TestNew_ModifyResponseReturnsNil(t *testing.T) {
	p, err := proxy.New("http://backend:8081")
	if err != nil {
		t.Fatal(err)
	}

	resp := &http.Response{Header: make(http.Header)}
	if err := p.ModifyResponse(resp); err != nil {
		t.Errorf("ModifyResponse returned unexpected error: %v", err)
	}
}
