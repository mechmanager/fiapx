package handler_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/upload-service/domain"
	"github.com/mechmanager/fiapx/upload-service/handler"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// --- mock use case ---

type mockUploadUseCase struct {
	video *domain.Video
	err   error
}

func (m *mockUploadUseCase) Upload(_ context.Context, _ uuid.UUID, _ string, _ io.Reader, _ int64) (*domain.Video, error) {
	return m.video, m.err
}

// --- helpers ---

func newUploadRouter(svc handler.UploadUseCase) *gin.Engine {
	r := gin.New()
	h := handler.NewUploadHandler(svc)
	r.POST("/videos", h.Upload)
	r.GET("/health", handler.Health)
	return r
}

func makeMultipartRequest(fieldName, filename, content string) (*http.Request, error) {
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	part, err := w.CreateFormFile(fieldName, filename)
	if err != nil {
		return nil, err
	}
	if _, err = io.WriteString(part, content); err != nil {
		return nil, err
	}
	w.Close()
	req, err := http.NewRequest(http.MethodPost, "/videos", body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req, nil
}

// --- tests: Health ---

func TestHealth(t *testing.T) {
	r := newUploadRouter(&mockUploadUseCase{})
	req, _ := http.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

// --- tests: Upload ---

func TestUpload_MissingUserID(t *testing.T) {
	r := newUploadRouter(&mockUploadUseCase{})
	req, _ := makeMultipartRequest("video", "file.mp4", "data")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestUpload_InvalidUserID(t *testing.T) {
	r := newUploadRouter(&mockUploadUseCase{})
	req, _ := makeMultipartRequest("video", "file.mp4", "data")
	req.Header.Set("X-User-ID", "not-a-uuid")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", w.Code)
	}
}

func TestUpload_MissingFile(t *testing.T) {
	r := newUploadRouter(&mockUploadUseCase{})
	req, _ := http.NewRequest(http.MethodPost, "/videos", nil)
	req.Header.Set("X-User-ID", uuid.New().String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}

func TestUpload_InvalidFormat(t *testing.T) {
	r := newUploadRouter(&mockUploadUseCase{err: domain.ErrInvalidVideoFormat})
	req, _ := makeMultipartRequest("video", "file.txt", "data")
	req.Header.Set("X-User-ID", uuid.New().String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected 422, got %d", w.Code)
	}
}

func TestUpload_ServiceError(t *testing.T) {
	r := newUploadRouter(&mockUploadUseCase{err: errors.New("internal error")})
	req, _ := makeMultipartRequest("video", "file.mp4", "data")
	req.Header.Set("X-User-ID", uuid.New().String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", w.Code)
	}
}

func TestUpload_Success(t *testing.T) {
	vid := &domain.Video{ID: uuid.New(), Status: "PENDING"}
	r := newUploadRouter(&mockUploadUseCase{video: vid})
	req, _ := makeMultipartRequest("video", "file.mp4", "data")
	req.Header.Set("X-User-ID", uuid.New().String())
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("expected 202, got %d", w.Code)
	}
}
