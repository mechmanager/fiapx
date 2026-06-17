package handler

import (
	"context"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/domain"
	"github.com/mechmanager/fiapx/api-gateway/middleware"
)

// VideoUseCase descreve as operações de vídeo usadas pelo handler.
type VideoUseCase interface {
	Upload(ctx context.Context, userID uuid.UUID, filename string, reader io.Reader, size int64) (*domain.Video, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error)
	GetDownloadStream(ctx context.Context, userID, videoID uuid.UUID) (io.ReadCloser, error)
}

// VideoHandler expõe os endpoints de vídeo.
type VideoHandler struct {
	video VideoUseCase
}

// NewVideoHandler cria o handler de vídeos.
func NewVideoHandler(video VideoUseCase) *VideoHandler {
	return &VideoHandler{video: video}
}

// Upload trata POST /videos — recebe o arquivo e enfileira o processamento.
func (h *VideoHandler) Upload(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserIDKey).(uuid.UUID)

	// Recebe o arquivo multipart com limite de 500 MB.
	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "arquivo de vídeo ausente"})
		return
	}
	defer file.Close()

	video, err := h.video.Upload(c.Request.Context(), userID, header.Filename, file, header.Size)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidVideoFormat) {
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao processar upload"})
		return
	}

	c.JSON(http.StatusAccepted, gin.H{
		"id":     video.ID,
		"status": video.Status,
	})
}

// List trata GET /videos — lista os vídeos do usuário autenticado.
func (h *VideoHandler) List(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserIDKey).(uuid.UUID)

	videos, err := h.video.ListByUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar vídeos"})
		return
	}

	// Retorna lista vazia em vez de null quando não há vídeos.
	if videos == nil {
		videos = []*domain.Video{}
	}

	type videoResponse struct {
		ID               string `json:"id"`
		OriginalFilename string `json:"original_filename"`
		Status           string `json:"status"`
		ErrorMessage     string `json:"error_message,omitempty"`
		FrameCount       int    `json:"frame_count"`
		CreatedAt        string `json:"created_at"`
	}

	result := make([]videoResponse, len(videos))
	for i, v := range videos {
		result[i] = videoResponse{
			ID:               v.ID.String(),
			OriginalFilename: v.OriginalFilename,
			Status:           v.Status,
			ErrorMessage:     v.ErrorMessage,
			FrameCount:       v.FrameCount,
			CreatedAt:        v.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}

	c.JSON(http.StatusOK, result)
}

// Download trata GET /videos/:id/download — faz stream do zip de frames.
func (h *VideoHandler) Download(c *gin.Context) {
	userID := c.MustGet(middleware.ContextUserIDKey).(uuid.UUID)

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	stream, err := h.video.GetDownloadStream(c.Request.Context(), userID, videoID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVideoNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "vídeo não encontrado"})
		case errors.Is(err, domain.ErrVideoNotReady):
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao baixar arquivo"})
		}
		return
	}
	defer stream.Close()

	c.Header("Content-Disposition", "attachment; filename=frames_"+videoID.String()+".zip")
	c.Header("Content-Type", "application/zip")
	io.Copy(c.Writer, stream)
}
