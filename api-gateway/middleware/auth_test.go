package middleware_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/middleware"
	"github.com/mechmanager/fiapx/api-gateway/mocks"
)

func init() { gin.SetMode(gin.TestMode) }

func newProtectedRouter(auth *mocks.Authenticator) *gin.Engine {
	r := gin.New()
	r.GET("/protegido", middleware.JWTAuth(auth), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	return r
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	r := newProtectedRouter(&mocks.Authenticator{})
	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperava 401, obteve %d", w.Code)
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	r := newProtectedRouter(&mocks.Authenticator{})
	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "tokendiretosemprefixo")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperava 401, obteve %d", w.Code)
	}
}

func TestJWTAuth_InvalidToken(t *testing.T) {
	auth := &mocks.Authenticator{
		ValidateFn: func(_ string) (uuid.UUID, error) {
			return uuid.Nil, errors.New("token inválido")
		},
	}
	r := newProtectedRouter(auth)
	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer token-invalido")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperava 401, obteve %d", w.Code)
	}
}

func TestJWTAuth_ValidToken(t *testing.T) {
	userID := uuid.New()
	auth := &mocks.Authenticator{
		ValidateFn: func(_ string) (uuid.UUID, error) {
			return userID, nil
		},
	}
	r := newProtectedRouter(auth)
	req := httptest.NewRequest(http.MethodGet, "/protegido", nil)
	req.Header.Set("Authorization", "Bearer token-valido")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("esperava 200, obteve %d", w.Code)
	}
}
