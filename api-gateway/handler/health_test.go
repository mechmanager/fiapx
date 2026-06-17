package handler_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/mechmanager/fiapx/api-gateway/handler"
)

func TestHealthHandler(t *testing.T) {
	r := gin.New()
	r.GET("/health", handler.Health)
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("esperava 200, obteve %d", w.Code)
	}
}
