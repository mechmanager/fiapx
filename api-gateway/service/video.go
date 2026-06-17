// Package service contém a lógica de negócio do api-gateway.
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/domain"
)

// VideoUseCase expõe as operações de vídeo disponíveis para os handlers.
type VideoUseCase interface {
	Upload(ctx context.Context, userID uuid.UUID, filename string, reader io.Reader, size int64) (*domain.Video, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error)
	GetDownloadStream(ctx context.Context, userID, videoID uuid.UUID) (io.ReadCloser, error)
}

// videoMessage é a mensagem publicada na fila para o worker processar.
type videoMessage struct {
	VideoID uuid.UUID `json:"video_id"`
	UserID  uuid.UUID `json:"user_id"`
	S3Key   string    `json:"s3_key"`
}

// VideoService orquestra upload, listagem e download de vídeos.
type VideoService struct {
	videos    domain.VideoRepository
	storage   domain.ObjectStorage
	publisher domain.QueuePublisher
}

// NewVideoService cria o serviço de vídeos com suas dependências.
func NewVideoService(
	videos domain.VideoRepository,
	storage domain.ObjectStorage,
	publisher domain.QueuePublisher,
) *VideoService {
	return &VideoService{videos: videos, storage: storage, publisher: publisher}
}

// Upload salva o vídeo no MinIO, persiste o registro com status PENDING
// e publica a mensagem na fila para processamento assíncrono.
func (s *VideoService) Upload(ctx context.Context, userID uuid.UUID, filename string, reader io.Reader, size int64) (*domain.Video, error) {
	// Valida a extensão do arquivo antes de qualquer operação.
	if !isValidVideoFile(filename) {
		return nil, domain.ErrInvalidVideoFormat
	}

	videoID := uuid.New()
	// Chave única no object storage para evitar colisões.
	s3Key := fmt.Sprintf("videos/%s/%s", videoID, filename)

	// Envia o vídeo para o MinIO.
	if err := s.storage.Upload(ctx, s3Key, reader, size, "video/mp4"); err != nil {
		return nil, fmt.Errorf("erro ao armazenar vídeo: %w", err)
	}

	// Registra o vídeo no banco com status inicial PENDING.
	video := &domain.Video{
		ID:               videoID,
		UserID:           userID,
		OriginalFilename: filename,
		S3Key:            s3Key,
		Status:           domain.StatusPending,
	}
	if err := s.videos.Create(ctx, video); err != nil {
		return nil, fmt.Errorf("erro ao registrar vídeo: %w", err)
	}

	// Publica na fila para o worker processar de forma assíncrona.
	msg := videoMessage{VideoID: videoID, UserID: userID, S3Key: s3Key}
	payload, err := json.Marshal(msg)
	if err != nil {
		return nil, fmt.Errorf("erro ao serializar mensagem: %w", err)
	}
	if err := s.publisher.Publish(ctx, payload); err != nil {
		return nil, fmt.Errorf("erro ao publicar na fila: %w", err)
	}

	return video, nil
}

// ListByUser devolve os vídeos pertencentes ao usuário autenticado.
func (s *VideoService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error) {
	return s.videos.ListByUser(ctx, userID)
}

// GetDownloadStream valida a propriedade do vídeo e devolve o stream do zip.
func (s *VideoService) GetDownloadStream(ctx context.Context, userID, videoID uuid.UUID) (io.ReadCloser, error) {
	video, err := s.videos.FindByID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	// Garante que apenas o dono do vídeo pode baixá-lo.
	if video.UserID != userID {
		return nil, domain.ErrVideoNotFound
	}

	// Vídeo ainda não foi processado.
	if video.Status != domain.StatusDone || video.ZipS3Key == "" {
		return nil, domain.ErrVideoNotReady
	}

	return s.storage.Download(ctx, video.ZipS3Key)
}

// isValidVideoFile aceita as extensões de vídeo suportadas pelo ffmpeg.
func isValidVideoFile(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	valid := map[string]bool{
		".mp4": true, ".avi": true, ".mov": true,
		".mkv": true, ".wmv": true, ".flv": true, ".webm": true,
	}
	return valid[ext]
}
