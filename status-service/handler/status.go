// Package handler expõe os endpoints HTTP do status-service.
package handler

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/status-service/domain"
	"github.com/mechmanager/fiapx/status-service/service"
)

// StatusUseCase descreve as operações usadas pelo handler.
type StatusUseCase interface {
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error)
	GetDownloadStream(ctx context.Context, userID, videoID uuid.UUID) (io.ReadCloser, error)
}

// StatusHandler expõe os endpoints de consulta e download de vídeos.
type StatusHandler struct {
	svc StatusUseCase
}

// NewStatusHandler cria o handler de status.
func NewStatusHandler(svc StatusUseCase) *StatusHandler {
	return &StatusHandler{svc: svc}
}

// List trata GET /videos — lista os vídeos do usuário identificado pelo header X-User-ID.
func (h *StatusHandler) List(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	videos, err := h.svc.ListByUser(c.Request.Context(), userID)
	if err != nil {
		log.Printf("service=status-service event=list status=error user_id=%s err=%v", userID, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao listar vídeos"})
		return
	}

	if videos == nil {
		videos = []*domain.Video{}
	}
	log.Printf("service=status-service event=list status=ok user_id=%s count=%d", userID, len(videos))

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
func (h *StatusHandler) Download(c *gin.Context) {
	userID, ok := parseUserID(c)
	if !ok {
		return
	}

	videoID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "id inválido"})
		return
	}

	stream, err := h.svc.GetDownloadStream(c.Request.Context(), userID, videoID)
	if err != nil {
		switch {
		case errors.Is(err, domain.ErrVideoNotFound):
			log.Printf("service=status-service event=download status=not_found user_id=%s video_id=%s", userID, videoID)
			c.JSON(http.StatusNotFound, gin.H{"error": "vídeo não encontrado"})
		case errors.Is(err, domain.ErrVideoNotReady):
			log.Printf("service=status-service event=download status=not_ready user_id=%s video_id=%s", userID, videoID)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		case service.IsOwnershipError(err):
			log.Printf("service=status-service event=download status=forbidden user_id=%s video_id=%s", userID, videoID)
			c.JSON(http.StatusForbidden, gin.H{"error": "acesso negado"})
		default:
			log.Printf("service=status-service event=download status=error user_id=%s video_id=%s err=%v", userID, videoID, err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao baixar arquivo"})
		}
		return
	}
	defer stream.Close()

	log.Printf("service=status-service event=download status=ok user_id=%s video_id=%s", userID, videoID)
	c.Header("Content-Disposition", "attachment; filename=frames_"+videoID.String()+".zip")
	c.Header("Content-Type", "application/zip")
	if _, err := io.Copy(c.Writer, stream); err != nil {
		log.Printf("service=status-service event=download status=stream_error video_id=%s err=%v", videoID, err)
	}
}

// parseUserID lê e valida o header X-User-ID. Responde 401 se ausente ou inválido.
func parseUserID(c *gin.Context) (uuid.UUID, bool) {
	header := c.GetHeader("X-User-ID")
	if header == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "header X-User-ID ausente"})
		return uuid.Nil, false
	}
	userID, err := uuid.Parse(header)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "header X-User-ID inválido"})
		return uuid.Nil, false
	}
	return userID, true
}
