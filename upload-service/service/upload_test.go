package service_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/upload-service/domain"
	"github.com/mechmanager/fiapx/upload-service/service"
)

// --- mocks ---

type mockRepo struct{ err error }

func (m *mockRepo) Create(_ context.Context, _ *domain.Video) error { return m.err }

type mockStorage struct{ err error }

func (m *mockStorage) Upload(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error {
	return m.err
}

type mockPublisher struct{ err error }

func (m *mockPublisher) Publish(_ context.Context, _ []byte) error { return m.err }

// --- helper ---

func newSvc(repo domain.VideoRepository, stor domain.ObjectStorage, pub domain.QueuePublisher) *service.UploadService {
	return service.NewUploadService(repo, stor, pub)
}

// --- tests ---

func TestUpload_InvalidExtension(t *testing.T) {
	svc := newSvc(&mockRepo{}, &mockStorage{}, &mockPublisher{})
	_, err := svc.Upload(context.Background(), uuid.New(), "file.txt", strings.NewReader(""), 0)
	if !errors.Is(err, domain.ErrInvalidVideoFormat) {
		t.Errorf("expected ErrInvalidVideoFormat, got %v", err)
	}
}

func TestUpload_ValidExtensions(t *testing.T) {
	exts := []string{"video.mp4", "video.avi", "video.mov", "video.mkv", "video.wmv", "video.flv", "video.webm"}
	for _, name := range exts {
		t.Run(name, func(t *testing.T) {
			svc := newSvc(&mockRepo{}, &mockStorage{}, &mockPublisher{})
			v, err := svc.Upload(context.Background(), uuid.New(), name, strings.NewReader("data"), 4)
			if err != nil {
				t.Errorf("unexpected error for %s: %v", name, err)
			}
			if v == nil {
				t.Error("expected non-nil video")
			}
		})
	}
}

func TestUpload_StorageError(t *testing.T) {
	svc := newSvc(&mockRepo{}, &mockStorage{err: errors.New("storage fail")}, &mockPublisher{})
	_, err := svc.Upload(context.Background(), uuid.New(), "file.mp4", strings.NewReader(""), 0)
	if err == nil {
		t.Fatal("expected storage error")
	}
}

func TestUpload_RepoError(t *testing.T) {
	svc := newSvc(&mockRepo{err: errors.New("db fail")}, &mockStorage{}, &mockPublisher{})
	_, err := svc.Upload(context.Background(), uuid.New(), "file.mp4", strings.NewReader(""), 0)
	if err == nil {
		t.Fatal("expected repo error")
	}
}

func TestUpload_PublisherError(t *testing.T) {
	svc := newSvc(&mockRepo{}, &mockStorage{}, &mockPublisher{err: errors.New("queue fail")})
	_, err := svc.Upload(context.Background(), uuid.New(), "file.mp4", strings.NewReader(""), 0)
	if err == nil {
		t.Fatal("expected publisher error")
	}
}

func TestUpload_Success_HasPendingStatus(t *testing.T) {
	svc := newSvc(&mockRepo{}, &mockStorage{}, &mockPublisher{})
	v, err := svc.Upload(context.Background(), uuid.New(), "file.mp4", strings.NewReader("data"), 4)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if v.Status != "PENDING" {
		t.Errorf("expected PENDING status, got %q", v.Status)
	}
}
