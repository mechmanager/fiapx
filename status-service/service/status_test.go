package service_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/status-service/domain"
	"github.com/mechmanager/fiapx/status-service/service"
)

// --- mocks ---

type mockRepo struct {
	videos []*domain.Video
	err    error
}

func (m *mockRepo) ListByUser(_ context.Context, _ uuid.UUID) ([]*domain.Video, error) {
	return m.videos, m.err
}

func (m *mockRepo) FindByID(_ context.Context, id uuid.UUID) (*domain.Video, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, v := range m.videos {
		if v.ID == id {
			return v, nil
		}
	}
	return nil, domain.ErrVideoNotFound
}

func (m *mockRepo) Delete(_ context.Context, _ uuid.UUID) error { return m.err }

type mockStorage struct {
	stream io.ReadCloser
	err    error
}

func (m *mockStorage) Download(_ context.Context, _ string) (io.ReadCloser, error) {
	return m.stream, m.err
}

func (m *mockStorage) Delete(_ context.Context, _ ...string) error { return m.err }

type mockCache struct {
	data map[string]string
	err  error
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	v, ok := m.data[key]
	if !ok {
		return "", errors.New("not found")
	}
	return v, nil
}

func (m *mockCache) Set(_ context.Context, key, value string, _ time.Duration) error {
	if m.data == nil {
		m.data = make(map[string]string)
	}
	m.data[key] = value
	return nil
}

// --- helpers ---

func newVideo(status string) *domain.Video {
	return &domain.Video{
		ID:               uuid.New(),
		UserID:           uuid.New(),
		OriginalFilename: "test.mp4",
		Status:           status,
		ZipS3Key:         "videos/zip/test.zip",
	}
}

// --- tests: ListByUser ---

func TestListByUser_RepoError(t *testing.T) {
	svc := service.NewStatusService(
		&mockRepo{err: errors.New("db error")},
		&mockStorage{},
		&mockCache{},
	)
	_, err := svc.ListByUser(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected repo error")
	}
}

func TestListByUser_CacheHit(t *testing.T) {
	v := newVideo("PENDING")
	key := "video:" + v.ID.String() + ":status"
	cache := &mockCache{data: map[string]string{key: "DONE"}}

	svc := service.NewStatusService(
		&mockRepo{videos: []*domain.Video{v}},
		&mockStorage{},
		cache,
	)
	videos, err := svc.ListByUser(context.Background(), v.UserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if videos[0].Status != "DONE" {
		t.Errorf("expected cache status DONE, got %q", videos[0].Status)
	}
}

func TestListByUser_CacheMiss_PopulatesCache(t *testing.T) {
	v := newVideo("PROCESSING")
	cache := &mockCache{data: make(map[string]string)}

	svc := service.NewStatusService(
		&mockRepo{videos: []*domain.Video{v}},
		&mockStorage{},
		cache,
	)
	videos, err := svc.ListByUser(context.Background(), v.UserID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if videos[0].Status != "PROCESSING" {
		t.Errorf("expected status PROCESSING, got %q", videos[0].Status)
	}
}

// --- tests: GetDownloadStream ---

func TestGetDownloadStream_Success(t *testing.T) {
	v := newVideo("DONE")
	stream := io.NopCloser(strings.NewReader("zip content"))

	svc := service.NewStatusService(
		&mockRepo{videos: []*domain.Video{v}},
		&mockStorage{stream: stream},
		&mockCache{},
	)
	rc, err := svc.GetDownloadStream(context.Background(), v.UserID, v.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer rc.Close()
}

func TestGetDownloadStream_RepoError(t *testing.T) {
	svc := service.NewStatusService(
		&mockRepo{err: errors.New("db error")},
		&mockStorage{},
		&mockCache{},
	)
	_, err := svc.GetDownloadStream(context.Background(), uuid.New(), uuid.New())
	if err == nil {
		t.Fatal("expected repo error")
	}
}

func TestGetDownloadStream_WrongOwner(t *testing.T) {
	v := newVideo("DONE")
	svc := service.NewStatusService(
		&mockRepo{videos: []*domain.Video{v}},
		&mockStorage{},
		&mockCache{},
	)
	_, err := svc.GetDownloadStream(context.Background(), uuid.New(), v.ID)
	if err == nil {
		t.Fatal("expected ownership error")
	}
	if !service.IsOwnershipError(err) {
		t.Errorf("expected ownershipError, got %T", err)
	}
}

func TestGetDownloadStream_NotReady(t *testing.T) {
	v := newVideo("PROCESSING")
	v.ZipS3Key = ""
	svc := service.NewStatusService(
		&mockRepo{videos: []*domain.Video{v}},
		&mockStorage{},
		&mockCache{},
	)
	_, err := svc.GetDownloadStream(context.Background(), v.UserID, v.ID)
	if !errors.Is(err, domain.ErrVideoNotReady) {
		t.Errorf("expected ErrVideoNotReady, got %v", err)
	}
}

func TestGetDownloadStream_NotFound(t *testing.T) {
	svc := service.NewStatusService(
		&mockRepo{videos: []*domain.Video{}},
		&mockStorage{},
		&mockCache{},
	)
	_, err := svc.GetDownloadStream(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrVideoNotFound) {
		t.Errorf("expected ErrVideoNotFound, got %v", err)
	}
}

// --- tests: StatusCode / IsOwnershipError ---

func TestStatusCode_Nil(t *testing.T) {
	if code := service.StatusCode(nil); code != 200 {
		t.Errorf("expected 200, got %d", code)
	}
}

func TestStatusCode_OtherError(t *testing.T) {
	if code := service.StatusCode(errors.New("generic")); code != 500 {
		t.Errorf("expected 500, got %d", code)
	}
}
