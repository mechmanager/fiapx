package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/domain"
	"github.com/mechmanager/fiapx/api-gateway/handler"
	"github.com/mechmanager/fiapx/api-gateway/mocks"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newAuthRouter(uc *mocks.AuthUseCase) *gin.Engine {
	r := gin.New()
	h := handler.NewAuthHandler(uc)
	r.POST("/auth/register", h.Register)
	r.POST("/auth/login", h.Login)
	return r
}

func bodyJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}

// --- Register ---

func TestRegisterHandler_Created(t *testing.T) {
	uc := &mocks.AuthUseCase{
		RegisterFn: func(_ context.Context, name, email, _ string) (*domain.User, error) {
			return &domain.User{ID: uuid.New(), Name: name, Email: email}, nil
		},
	}
	r := newAuthRouter(uc)
	req := httptest.NewRequest(http.MethodPost, "/auth/register",
		bodyJSON(t, map[string]string{"name": "Alice", "email": "a@b.com", "password": "123456"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Errorf("esperava 201, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestRegisterHandler_BadJSON(t *testing.T) {
	r := newAuthRouter(&mocks.AuthUseCase{})
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBufferString("não-é-json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, obteve %d", w.Code)
	}
}

func TestRegisterHandler_EmailConflict(t *testing.T) {
	uc := &mocks.AuthUseCase{
		RegisterFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
			return nil, domain.ErrEmailAlreadyExists
		},
	}
	r := newAuthRouter(uc)
	req := httptest.NewRequest(http.MethodPost, "/auth/register",
		bodyJSON(t, map[string]string{"name": "A", "email": "a@b.com", "password": "123456"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("esperava 409, obteve %d", w.Code)
	}
}

func TestRegisterHandler_InternalError(t *testing.T) {
	uc := &mocks.AuthUseCase{
		RegisterFn: func(_ context.Context, _, _, _ string) (*domain.User, error) {
			return nil, errors.New("banco indisponível")
		},
	}
	r := newAuthRouter(uc)
	req := httptest.NewRequest(http.MethodPost, "/auth/register",
		bodyJSON(t, map[string]string{"name": "A", "email": "a@b.com", "password": "123456"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperava 500, obteve %d", w.Code)
	}
}

// --- Login ---

func TestLoginHandler_OK(t *testing.T) {
	uc := &mocks.AuthUseCase{
		LoginFn: func(_ context.Context, _, _ string) (string, error) {
			return "jwt.token.aqui", nil
		},
	}
	r := newAuthRouter(uc)
	req := httptest.NewRequest(http.MethodPost, "/auth/login",
		bodyJSON(t, map[string]string{"email": "a@b.com", "password": "123456"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("esperava 200, obteve %d: %s", w.Code, w.Body.String())
	}
	var body map[string]string
	json.NewDecoder(w.Body).Decode(&body)
	if body["token"] != "jwt.token.aqui" {
		t.Errorf("token incorreto: %q", body["token"])
	}
}

func TestLoginHandler_BadJSON(t *testing.T) {
	r := newAuthRouter(&mocks.AuthUseCase{})
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, obteve %d", w.Code)
	}
}

func TestLoginHandler_Unauthorized(t *testing.T) {
	uc := &mocks.AuthUseCase{
		LoginFn: func(_ context.Context, _, _ string) (string, error) {
			return "", domain.ErrInvalidCredentials
		},
	}
	r := newAuthRouter(uc)
	req := httptest.NewRequest(http.MethodPost, "/auth/login",
		bodyJSON(t, map[string]string{"email": "a@b.com", "password": "errada"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("esperava 401, obteve %d", w.Code)
	}
}

func TestLoginHandler_InternalError(t *testing.T) {
	uc := &mocks.AuthUseCase{
		LoginFn: func(_ context.Context, _, _ string) (string, error) {
			return "", errors.New("banco indisponível")
		},
	}
	r := newAuthRouter(uc)
	req := httptest.NewRequest(http.MethodPost, "/auth/login",
		bodyJSON(t, map[string]string{"email": "a@b.com", "password": "123456"}))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperava 500, obteve %d", w.Code)
	}
}
