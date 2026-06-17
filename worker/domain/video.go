package domain

import (
	"context"
	"io"

	"github.com/google/uuid"
)

const (
	StatusProcessing = "PROCESSING"
	StatusDone       = "DONE"
	StatusError      = "ERROR"
)

type VideoMessage struct {
	VideoID  uuid.UUID `json:"video_id"`
	UserID   uuid.UUID `json:"user_id"`
	S3Key    string    `json:"s3_key"`
	Filename string    `json:"filename"`
}

type NotificationMessage struct {
	UserID   uuid.UUID `json:"user_id"`
	VideoID  uuid.UUID `json:"video_id"`
	Filename string    `json:"filename"`
	Error    string    `json:"error"`
}

type VideoRepository interface {
	UpdateStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error
	UpdateDone(ctx context.Context, id uuid.UUID, zipS3Key string, frameCount int) error
}

type ObjectStorage interface {
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	UploadFile(ctx context.Context, objectKey, filePath string) error
}

// Notifier abstrai o envio de notificações de erro de processamento.
type Notifier interface {
	NotifyError(ctx context.Context, videoID, userID uuid.UUID, filename, reason string)
}
