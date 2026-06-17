package proxy_test

import (
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
