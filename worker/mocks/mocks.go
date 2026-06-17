package mocks

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// --- VideoRepository ---

type VideoRepository struct {
	UpdateStatusFn func(ctx context.Context, id uuid.UUID, status, errorMsg string) error
	UpdateDoneFn   func(ctx context.Context, id uuid.UUID, zipS3Key string, frameCount int) error
}

func (m *VideoRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error {
	return m.UpdateStatusFn(ctx, id, status, errorMsg)
}
func (m *VideoRepository) UpdateDone(ctx context.Context, id uuid.UUID, zipS3Key string, frameCount int) error {
	return m.UpdateDoneFn(ctx, id, zipS3Key, frameCount)
}

// --- ObjectStorage ---

type ObjectStorage struct {
	DownloadFn   func(ctx context.Context, key string) (io.ReadCloser, error)
	UploadFileFn func(ctx context.Context, key, path string) error
}

func (m *ObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.DownloadFn(ctx, key)
}
func (m *ObjectStorage) UploadFile(ctx context.Context, key, path string) error {
	return m.UploadFileFn(ctx, key, path)
}

// --- Notifier ---

type Notifier struct {
	NotifyErrorFn func(ctx context.Context, videoID, userID uuid.UUID, filename, reason string)
}

func (m *Notifier) NotifyError(ctx context.Context, videoID, userID uuid.UUID, filename, reason string) {
	if m.NotifyErrorFn != nil {
		m.NotifyErrorFn(ctx, videoID, userID, filename, reason)
	}
}
