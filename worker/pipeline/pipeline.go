// Package pipeline implementa o processamento de uma mensagem de vídeo.
// É completamente independente de RabbitMQ e pode ser testado com mocks simples.
package pipeline

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/worker/domain"
	"github.com/mechmanager/fiapx/worker/processor"
)

// VideoProcessor abstrai a execução do pipeline ffmpeg → zip.
type VideoProcessor interface {
	Process(ctx context.Context, videoReader io.Reader, videoFilename, workDir string) (*processor.Result, error)
}

// Message é a mensagem recebida da fila (mesmo layout que domain.VideoMessage).
type Message struct {
	VideoID uuid.UUID `json:"video_id"`
	UserID  uuid.UUID `json:"user_id"`
	S3Key   string    `json:"s3_key"`
}

// Pipeline encapsula a lógica de processamento de uma mensagem de vídeo.
type Pipeline struct {
	videos   domain.VideoRepository
	storage  domain.ObjectStorage
	proc     VideoProcessor
	notifier domain.Notifier
}

// New cria um Pipeline com suas dependências.
func New(
	videos domain.VideoRepository,
	stor domain.ObjectStorage,
	proc VideoProcessor,
	notifier domain.Notifier,
) *Pipeline {
	return &Pipeline{videos: videos, storage: stor, proc: proc, notifier: notifier}
}

// HandleMessage processa o corpo bruto de uma mensagem AMQP.
// Retorna (true, nil) em sucesso, (false, err) em falha de processamento e
// (false, nil) para mensagem mal formada (descartável sem requeue).
func (p *Pipeline) HandleMessage(ctx context.Context, body []byte) (ack bool, err error) {
	var m Message
	if err := json.Unmarshal(body, &m); err != nil {
		log.Printf("[WARN] mensagem inválida, descartando: %v", err)
		return false, nil
	}

	log.Printf("processando video_id=%s", m.VideoID)

	if err := p.process(ctx, m); err != nil {
		log.Printf("[ERRO] video_id=%s: %v", m.VideoID, err)
		p.videos.UpdateStatus(ctx, m.VideoID, domain.StatusError, err.Error())
		p.notifier.NotifyError(m.VideoID, m.UserID, err.Error())
		return false, err
	}

	return true, nil
}

func (p *Pipeline) process(ctx context.Context, m Message) error {
	if err := p.videos.UpdateStatus(ctx, m.VideoID, domain.StatusProcessing, ""); err != nil {
		return fmt.Errorf("erro ao atualizar status para PROCESSING: %w", err)
	}

	workDir, err := os.MkdirTemp("", "fiapx-"+m.VideoID.String())
	if err != nil {
		return fmt.Errorf("erro ao criar diretório temporário: %w", err)
	}
	defer os.RemoveAll(workDir)

	videoStream, err := p.storage.Download(ctx, m.S3Key)
	if err != nil {
		return fmt.Errorf("erro ao baixar vídeo: %w", err)
	}
	defer videoStream.Close()

	result, err := p.proc.Process(ctx, videoStream, filepath.Base(m.S3Key), workDir)
	if err != nil {
		return fmt.Errorf("erro no processamento: %w", err)
	}

	zipKey := fmt.Sprintf("zips/%s/frames.zip", m.VideoID)
	if err := p.storage.UploadFile(ctx, zipKey, result.ZipPath); err != nil {
		return fmt.Errorf("erro ao enviar zip: %w", err)
	}

	if err := p.videos.UpdateDone(ctx, m.VideoID, zipKey, result.FrameCount); err != nil {
		return fmt.Errorf("erro ao atualizar status para DONE: %w", err)
	}

	return nil
}
