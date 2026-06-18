package handler_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/status-service/domain"
	"github.com/mechmanager/fiapx/status-service/handler"
	"github.com/mechmanager/fiapx/status-service/service"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- mock use case ---

type mockStatusUseCase struct {
	videos []*domain.Video
	stream io.ReadCloser
	err    error
}

func (m *mockStatusUseCase) ListByUser(_ context.Context, _ uuid.UUID) ([]*domain.Video, error) {
	return m.videos, m.err
}

func (m *mockStatusUseCase) GetDownloadStream(_ context.Context, _, _ uuid.UUID) (io.ReadCloser, error) {
	return m.stream, m.err
}

func (m *mockStatusUseCase) DeleteVideo(_ context.Context, _, _ uuid.UUID) error { return m.err }

// --- helpers ---

func newRouter(svc handler.StatusUseCase) *gin.Engine {
	r := gin.New()
	h := handler.NewStatusHandler(svc)
	r.GET("/videos", h.List)
	r.GET("/videos/:id/download", h.Download)
	r.DELETE("/videos/:id", h.Delete)
	r.GET("/health", handler.Health)
	return r
}

func perform(r *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req, _ := http.NewRequest(method, path, nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

// --- tests: Health ---

func TestHealth(t *testing.T) {
	r := newRouter(&mockStatusUseCase{})
	w := perform(r, "GET", "/health", nil)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- tests: List ---

func TestList_MissingUserID(t *testing.T) {
	r := newRouter(&mockStatusUseCase{videos: []*domain.Video{}})
	w := perform(r, "GET", "/videos", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestList_InvalidUserID(t *testing.T) {
	r := newRouter(&mockStatusUseCase{videos: []*domain.Video{}})
	w := perform(r, "GET", "/videos", map[string]string{"X-User-ID": "not-a-uuid"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestList_EmptyResult(t *testing.T) {
	r := newRouter(&mockStatusUseCase{videos: []*domain.Video{}})
	w := perform(r, "GET", "/videos", map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestList_WithVideos(t *testing.T) {
	v := &domain.Video{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		OriginalFilename: "sample.mp4",
		Status:           "DONE",
	}
	r := newRouter(&mockStatusUseCase{videos: []*domain.Video{v}})
	w := perform(r, "GET", "/videos", map[string]string{"X-User-ID": v.UserID.String()})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestList_ServiceError(t *testing.T) {
	r := newRouter(&mockStatusUseCase{err: errors.New("db error")})
	w := perform(r, "GET", "/videos", map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

// --- tests: Download ---

func TestDownload_MissingUserID(t *testing.T) {
	r := newRouter(&mockStatusUseCase{})
	w := perform(r, "GET", "/videos/"+uuid.New().String()+"/download", nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestDownload_InvalidVideoID(t *testing.T) {
	r := newRouter(&mockStatusUseCase{})
	w := perform(r, "GET", "/videos/not-a-uuid/download", map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDownload_Success(t *testing.T) {
	stream := io.NopCloser(strings.NewReader("zip data"))
	r := newRouter(&mockStatusUseCase{stream: stream})
	w := perform(r, "GET", "/videos/"+uuid.New().String()+"/download", map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestDownload_NotFound(t *testing.T) {
	r := newRouter(&mockStatusUseCase{err: domain.ErrVideoNotFound})
	w := perform(r, "GET", "/videos/"+uuid.New().String()+"/download", map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDownload_NotReady(t *testing.T) {
	r := newRouter(&mockStatusUseCase{err: domain.ErrVideoNotReady})
	w := perform(r, "GET", "/videos/"+uuid.New().String()+"/download", map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", w.Code)
	}
}

// --- tests: Delete ---

func TestDelete_MissingUserID(t *testing.T) {
	r := newRouter(&mockStatusUseCase{})
	w := perform(r, "DELETE", "/videos/"+uuid.New().String(), nil)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestDelete_InvalidVideoID(t *testing.T) {
	r := newRouter(&mockStatusUseCase{})
	w := perform(r, "DELETE", "/videos/not-a-uuid", map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestDelete_Success(t *testing.T) {
	r := newRouter(&mockStatusUseCase{})
	w := perform(r, "DELETE", "/videos/"+uuid.New().String(), map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", w.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	r := newRouter(&mockStatusUseCase{err: domain.ErrVideoNotFound})
	w := perform(r, "DELETE", "/videos/"+uuid.New().String(), map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}

func TestDelete_Forbidden(t *testing.T) {
	r := newRouter(&mockStatusUseCase{err: service.NewOwnershipError()})
	w := perform(r, "DELETE", "/videos/"+uuid.New().String(), map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", w.Code)
	}
}

func TestDelete_ServerError(t *testing.T) {
	r := newRouter(&mockStatusUseCase{err: errors.New("db error")})
	w := perform(r, "DELETE", "/videos/"+uuid.New().String(), map[string]string{"X-User-ID": uuid.New().String()})
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}
