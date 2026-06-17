package service_test

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/domain"
	"github.com/mechmanager/fiapx/api-gateway/mocks"
	"github.com/mechmanager/fiapx/api-gateway/service"
)

func buildVideoService(
	videosFn *mocks.VideoRepository,
	storageFn *mocks.ObjectStorage,
	pubFn *mocks.QueuePublisher,
) *service.VideoService {
	return service.NewVideoService(videosFn, storageFn, pubFn)
}

// --- Upload ---

func TestVideoUpload_HappyPath(t *testing.T) {
	svc := buildVideoService(
		&mocks.VideoRepository{
			CreateFn: func(_ context.Context, _ *domain.Video) error { return nil },
		},
		&mocks.ObjectStorage{
			UploadFn: func(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error { return nil },
		},
		&mocks.QueuePublisher{
			PublishFn: func(_ context.Context, _ []byte) error { return nil },
		},
	)
	video, err := svc.Upload(context.Background(), uuid.New(), "clip.mp4", strings.NewReader("data"), 4)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if video.Status != domain.StatusPending {
		t.Errorf("status esperado PENDING, obteve %q", video.Status)
	}
}

func TestVideoUpload_InvalidFormat(t *testing.T) {
	svc := buildVideoService(
		&mocks.VideoRepository{},
		&mocks.ObjectStorage{},
		&mocks.QueuePublisher{},
	)
	_, err := svc.Upload(context.Background(), uuid.New(), "arquivo.pdf", strings.NewReader(""), 0)
	if !errors.Is(err, domain.ErrInvalidVideoFormat) {
		t.Fatalf("esperava ErrInvalidVideoFormat, obteve %v", err)
	}
}

func TestVideoUpload_StorageError(t *testing.T) {
	svc := buildVideoService(
		&mocks.VideoRepository{},
		&mocks.ObjectStorage{
			UploadFn: func(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error {
				return errors.New("s3 indisponível")
			},
		},
		&mocks.QueuePublisher{},
	)
	_, err := svc.Upload(context.Background(), uuid.New(), "v.mp4", strings.NewReader(""), 0)
	if err == nil {
		t.Fatal("esperava erro de storage")
	}
}

func TestVideoUpload_RepoError(t *testing.T) {
	svc := buildVideoService(
		&mocks.VideoRepository{
			CreateFn: func(_ context.Context, _ *domain.Video) error {
				return errors.New("banco indisponível")
			},
		},
		&mocks.ObjectStorage{
			UploadFn: func(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error { return nil },
		},
		&mocks.QueuePublisher{},
	)
	_, err := svc.Upload(context.Background(), uuid.New(), "v.mp4", strings.NewReader(""), 0)
	if err == nil {
		t.Fatal("esperava erro de repositório")
	}
}

func TestVideoUpload_QueueError(t *testing.T) {
	svc := buildVideoService(
		&mocks.VideoRepository{
			CreateFn: func(_ context.Context, _ *domain.Video) error { return nil },
		},
		&mocks.ObjectStorage{
			UploadFn: func(_ context.Context, _ string, _ io.Reader, _ int64, _ string) error { return nil },
		},
		&mocks.QueuePublisher{
			PublishFn: func(_ context.Context, _ []byte) error {
				return errors.New("rabbitmq indisponível")
			},
		},
	)
	_, err := svc.Upload(context.Background(), uuid.New(), "v.mp4", strings.NewReader(""), 0)
	if err == nil {
		t.Fatal("esperava erro de fila")
	}
}

// --- ListByUser ---

func TestVideoListByUser_HappyPath(t *testing.T) {
	userID := uuid.New()
	svc := buildVideoService(
		&mocks.VideoRepository{
			ListByUserFn: func(_ context.Context, id uuid.UUID) ([]*domain.Video, error) {
				return []*domain.Video{{ID: uuid.New(), UserID: id}}, nil
			},
		},
		&mocks.ObjectStorage{},
		&mocks.QueuePublisher{},
	)
	videos, err := svc.ListByUser(context.Background(), userID)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if len(videos) != 1 {
		t.Errorf("esperava 1 vídeo, obteve %d", len(videos))
	}
}

func TestVideoListByUser_Error(t *testing.T) {
	svc := buildVideoService(
		&mocks.VideoRepository{
			ListByUserFn: func(_ context.Context, _ uuid.UUID) ([]*domain.Video, error) {
				return nil, errors.New("erro")
			},
		},
		&mocks.ObjectStorage{},
		&mocks.QueuePublisher{},
	)
	_, err := svc.ListByUser(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("esperava erro")
	}
}

// --- GetDownloadStream ---

func TestGetDownloadStream_HappyPath(t *testing.T) {
	userID := uuid.New()
	videoID := uuid.New()
	svc := buildVideoService(
		&mocks.VideoRepository{
			FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Video, error) {
				return &domain.Video{
					ID: videoID, UserID: userID,
					Status: domain.StatusDone, ZipS3Key: "zips/x.zip",
					CreatedAt: time.Now(), UpdatedAt: time.Now(),
				}, nil
			},
		},
		&mocks.ObjectStorage{
			DownloadFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
				return io.NopCloser(strings.NewReader("zip")), nil
			},
		},
		&mocks.QueuePublisher{},
	)
	rc, err := svc.GetDownloadStream(context.Background(), userID, videoID)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	rc.Close()
}

func TestGetDownloadStream_VideoNotFound(t *testing.T) {
	svc := buildVideoService(
		&mocks.VideoRepository{
			FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Video, error) {
				return nil, domain.ErrVideoNotFound
			},
		},
		&mocks.ObjectStorage{},
		&mocks.QueuePublisher{},
	)
	_, err := svc.GetDownloadStream(context.Background(), uuid.New(), uuid.New())
	if !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("esperava ErrVideoNotFound, obteve %v", err)
	}
}

func TestGetDownloadStream_NotOwner(t *testing.T) {
	videoID := uuid.New()
	svc := buildVideoService(
		&mocks.VideoRepository{
			FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Video, error) {
				return &domain.Video{ID: videoID, UserID: uuid.New(), Status: domain.StatusDone, ZipS3Key: "k"}, nil
			},
		},
		&mocks.ObjectStorage{},
		&mocks.QueuePublisher{},
	)
	// Outro usuário tentando baixar.
	_, err := svc.GetDownloadStream(context.Background(), uuid.New(), videoID)
	if !errors.Is(err, domain.ErrVideoNotFound) {
		t.Fatalf("esperava ErrVideoNotFound, obteve %v", err)
	}
}

func TestGetDownloadStream_NotReady(t *testing.T) {
	userID := uuid.New()
	videoID := uuid.New()
	svc := buildVideoService(
		&mocks.VideoRepository{
			FindByIDFn: func(_ context.Context, _ uuid.UUID) (*domain.Video, error) {
				return &domain.Video{ID: videoID, UserID: userID, Status: domain.StatusPending}, nil
			},
		},
		&mocks.ObjectStorage{},
		&mocks.QueuePublisher{},
	)
	_, err := svc.GetDownloadStream(context.Background(), userID, videoID)
	if !errors.Is(err, domain.ErrVideoNotReady) {
		t.Fatalf("esperava ErrVideoNotReady, obteve %v", err)
	}
}
