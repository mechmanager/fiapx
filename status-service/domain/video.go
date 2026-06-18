// Package domain contém os tipos e interfaces de domínio do status-service.
package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// Video representa um vídeo enviado por um usuário.
type Video struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	OriginalFilename string
	S3Key            string
	Status           string // PENDING, PROCESSING, DONE, ERROR
	ErrorMessage     string
	ZipS3Key         string
	FrameCount       int
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// VideoRepository define as operações de leitura de vídeos no banco de dados.
type VideoRepository interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Video, error)
	FindByID(ctx context.Context, id uuid.UUID) (*Video, error)
	Delete(ctx context.Context, id uuid.UUID) error
}

// ObjectStorage define as operações de acesso ao object storage.
type ObjectStorage interface {
	Download(ctx context.Context, key string) (io.ReadCloser, error)
	Delete(ctx context.Context, keys ...string) error
}

// StatusCache define as operações de cache de status de vídeos.
type StatusCache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key, value string, ttl time.Duration) error
}
