package handler_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/domain"
	"github.com/mechmanager/fiapx/api-gateway/handler"
	"github.com/mechmanager/fiapx/api-gateway/middleware"
	"github.com/mechmanager/fiapx/api-gateway/mocks"
)

func newVideoRouter(uc *mocks.VideoUseCase) *gin.Engine {
	r := gin.New()
	h := handler.NewVideoHandler(uc)
	// Injeta um userID fixo no contexto para simular o middleware JWT.
	injectUser := func(c *gin.Context) {
		c.Set(middleware.ContextUserIDKey, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		c.Next()
	}
	r.POST("/videos", injectUser, h.Upload)
	r.GET("/videos", injectUser, h.List)
	r.GET("/videos/:id/download", injectUser, h.Download)
	return r
}

func multipartVideo(t *testing.T, filename, content string) (*bytes.Buffer, string) {
	t.Helper()
	body := &bytes.Buffer{}
	w := multipart.NewWriter(body)
	fw, err := w.CreateFormFile("video", filename)
	if err != nil {
		t.Fatalf("CreateFormFile: %v", err)
	}
	fw.Write([]byte(content))
	w.Close()
	return body, w.FormDataContentType()
}

// --- Upload ---

func TestVideoUploadHandler_Accepted(t *testing.T) {
	userID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	uc := &mocks.VideoUseCase{
		UploadFn: func(_ context.Context, _ uuid.UUID, _ string, _ io.Reader, _ int64) (*domain.Video, error) {
			return &domain.Video{ID: uuid.New(), UserID: userID, Status: domain.StatusPending}, nil
		},
	}
	r := newVideoRouter(uc)
	body, ct := multipartVideo(t, "clip.mp4", "conteudo")
	req := httptest.NewRequest(http.MethodPost, "/videos", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusAccepted {
		t.Errorf("esperava 202, obteve %d: %s", w.Code, w.Body.String())
	}
}

func TestVideoUploadHandler_MissingFile(t *testing.T) {
	r := newVideoRouter(&mocks.VideoUseCase{})
	req := httptest.NewRequest(http.MethodPost, "/videos", strings.NewReader(""))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, obteve %d", w.Code)
	}
}

func TestVideoUploadHandler_InvalidFormat(t *testing.T) {
	uc := &mocks.VideoUseCase{
		UploadFn: func(_ context.Context, _ uuid.UUID, _ string, _ io.Reader, _ int64) (*domain.Video, error) {
			return nil, domain.ErrInvalidVideoFormat
		},
	}
	r := newVideoRouter(uc)
	body, ct := multipartVideo(t, "doc.pdf", "data")
	req := httptest.NewRequest(http.MethodPost, "/videos", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("esperava 422, obteve %d", w.Code)
	}
}

func TestVideoUploadHandler_InternalError(t *testing.T) {
	uc := &mocks.VideoUseCase{
		UploadFn: func(_ context.Context, _ uuid.UUID, _ string, _ io.Reader, _ int64) (*domain.Video, error) {
			return nil, errors.New("falha interna")
		},
	}
	r := newVideoRouter(uc)
	body, ct := multipartVideo(t, "v.mp4", "data")
	req := httptest.NewRequest(http.MethodPost, "/videos", body)
	req.Header.Set("Content-Type", ct)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperava 500, obteve %d", w.Code)
	}
}

// --- List ---

func TestVideoListHandler_OK(t *testing.T) {
	uc := &mocks.VideoUseCase{
		ListByUserFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Video, error) {
			return []*domain.Video{
				{ID: uuid.New(), Status: domain.StatusDone, CreatedAt: time.Now()},
			}, nil
		},
	}
	r := newVideoRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/videos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("esperava 200, obteve %d", w.Code)
	}
}

func TestVideoListHandler_Empty(t *testing.T) {
	uc := &mocks.VideoUseCase{
		ListByUserFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Video, error) {
			return nil, nil
		},
	}
	r := newVideoRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/videos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("esperava 200, obteve %d", w.Code)
	}
	if w.Body.String() == "null" {
		t.Error("resposta não deve ser null para lista vazia")
	}
}

func TestVideoListHandler_Error(t *testing.T) {
	uc := &mocks.VideoUseCase{
		ListByUserFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Video, error) {
			return nil, errors.New("erro")
		},
	}
	r := newVideoRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/videos", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperava 500, obteve %d", w.Code)
	}
}

// --- Download ---

func TestVideoDownloadHandler_OK(t *testing.T) {
	videoID := uuid.New()
	uc := &mocks.VideoUseCase{
		GetDownloadStreamFn: func(_ context.Context, _, _ uuid.UUID) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("zip-content")), nil
		},
	}
	r := newVideoRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/videos/"+videoID.String()+"/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Errorf("esperava 200, obteve %d", w.Code)
	}
}

func TestVideoDownloadHandler_InvalidUUID(t *testing.T) {
	r := newVideoRouter(&mocks.VideoUseCase{})
	req := httptest.NewRequest(http.MethodGet, "/videos/nao-e-uuid/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusBadRequest {
		t.Errorf("esperava 400, obteve %d", w.Code)
	}
}

func TestVideoDownloadHandler_NotFound(t *testing.T) {
	uc := &mocks.VideoUseCase{
		GetDownloadStreamFn: func(_ context.Context, _, _ uuid.UUID) (io.ReadCloser, error) {
			return nil, domain.ErrVideoNotFound
		},
	}
	r := newVideoRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/videos/"+uuid.New().String()+"/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Errorf("esperava 404, obteve %d", w.Code)
	}
}

func TestVideoDownloadHandler_NotReady(t *testing.T) {
	uc := &mocks.VideoUseCase{
		GetDownloadStreamFn: func(_ context.Context, _, _ uuid.UUID) (io.ReadCloser, error) {
			return nil, domain.ErrVideoNotReady
		},
	}
	r := newVideoRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/videos/"+uuid.New().String()+"/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusConflict {
		t.Errorf("esperava 409, obteve %d", w.Code)
	}
}

func TestVideoDownloadHandler_InternalError(t *testing.T) {
	uc := &mocks.VideoUseCase{
		GetDownloadStreamFn: func(_ context.Context, _, _ uuid.UUID) (io.ReadCloser, error) {
			return nil, errors.New("erro inesperado")
		},
	}
	r := newVideoRouter(uc)
	req := httptest.NewRequest(http.MethodGet, "/videos/"+uuid.New().String()+"/download", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusInternalServerError {
		t.Errorf("esperava 500, obteve %d", w.Code)
	}
}
