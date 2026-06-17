// Package domain contém as entidades e interfaces do worker.
package domain

import (
	"context"
	"io"

	"github.com/google/uuid"
)

// Status possíveis de um vídeo durante o processamento.
const (
	StatusProcessing = "PROCESSING"
	StatusDone       = "DONE"
	StatusError      = "ERROR"
)

// VideoMessage representa a mensagem recebida da fila.
type VideoMessage struct {
	VideoID uuid.UUID `json:"video_id"`
	UserID  uuid.UUID `json:"user_id"`
	S3Key   string    `json:"s3_key"`
}

// VideoRepository abstrai as atualizações de status no banco.
type VideoRepository interface {
	// UpdateStatus atualiza o status de um vídeo.
	UpdateStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error
	// UpdateDone marca o vídeo como processado com sucesso.
	UpdateDone(ctx context.Context, id uuid.UUID, zipS3Key string, frameCount int) error
}

// ObjectStorage abstrai o acesso ao object storage (MinIO/S3).
type ObjectStorage interface {
	// Download devolve um leitor para o conteúdo de um objeto.
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
	// Upload envia um objeto a partir de um arquivo local (caminho no disco).
	UploadFile(ctx context.Context, objectKey, filePath string) error
}

// Notifier abstrai o envio de notificações de erro.
type Notifier interface {
	// NotifyError registra ou envia uma notificação de falha no processamento.
	NotifyError(videoID uuid.UUID, userID uuid.UUID, reason string)
}
