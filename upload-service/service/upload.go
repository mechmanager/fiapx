// Package service contains the business logic for the upload-service.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/upload-service/domain"
)

// allowed video extensions.
var allowedExtensions = map[string]bool{
	".mp4":  true,
	".avi":  true,
	".mov":  true,
	".mkv":  true,
	".wmv":  true,
	".flv":  true,
	".webm": true,
}

// uploadMessage is the JSON payload published to the queue.
type uploadMessage struct {
	VideoID  string `json:"video_id"`
	UserID   string `json:"user_id"`
	S3Key    string `json:"s3_key"`
	Filename string `json:"filename"`
}

// UploadService orchestrates video upload: validates, stores, persists, and notifies.
type UploadService struct {
	videos    domain.VideoRepository
	storage   domain.ObjectStorage
	publisher domain.QueuePublisher
}

// NewUploadService creates an UploadService with its dependencies.
func NewUploadService(videos domain.VideoRepository, storage domain.ObjectStorage, publisher domain.QueuePublisher) *UploadService {
	return &UploadService{videos: videos, storage: storage, publisher: publisher}
}

// Upload validates the file extension, stores the object, persists the record, and
// publishes an upload event. Returns the created Video on success.
func (s *UploadService) Upload(ctx context.Context, userID uuid.UUID, filename string, reader io.Reader, size int64) (*domain.Video, error) {
	ext := strings.ToLower(filepath.Ext(filename))
	if !allowedExtensions[ext] {
		return nil, domain.ErrInvalidVideoFormat
	}

	videoID := uuid.New()
	s3Key := fmt.Sprintf("videos/%s/%s", videoID, filename)

	if err := s.storage.Upload(ctx, s3Key, reader, size, "video/"+strings.TrimPrefix(ext, ".")); err != nil {
		return nil, fmt.Errorf("storage upload failed: %w", err)
	}

	video := &domain.Video{
		ID:               videoID,
		UserID:           userID,
		OriginalFilename: filename,
		S3Key:            s3Key,
		Status:           "PENDING",
	}

	if err := s.videos.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("failed to persist video record: %w", err)
	}

	msg := uploadMessage{
		VideoID:  videoID.String(),
		UserID:   userID.String(),
		S3Key:    s3Key,
		Filename: filename,
	}
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal queue message: %w", err)
	}

	if err := s.publisher.Publish(ctx, payload); err != nil {
		return nil, fmt.Errorf("failed to publish queue message: %w", err)
	}

	return video, nil
}
