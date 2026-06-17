// Package domain contains the core entities and port interfaces for the upload-service.
package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// Video represents a video upload record.
type Video struct {
	ID               uuid.UUID // unique identifier
	UserID           uuid.UUID // owning user
	OriginalFilename string    // original file name supplied by the client
	S3Key            string    // object storage key
	Status           string    // e.g. PENDING, PROCESSING, DONE, ERROR
	CreatedAt        time.Time // creation timestamp
}

// VideoRepository abstracts persistence for videos.
type VideoRepository interface {
	// Create inserts a new video record (status = PENDING).
	Create(ctx context.Context, video *Video) error
}

// ObjectStorage abstracts object storage uploads.
type ObjectStorage interface {
	// Upload stores the content at the given key.
	Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) error
}

// QueuePublisher abstracts message queue publishing.
type QueuePublisher interface {
	// Publish sends a raw JSON payload to the configured queue.
	Publish(ctx context.Context, body []byte) error
}
