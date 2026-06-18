// Package handler contains HTTP handlers for the upload-service.
package handler

import (
	"context"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/upload-service/domain"
)

// UploadUseCase describes the upload operation used by the handler.
type UploadUseCase interface {
	Upload(ctx context.Context, userID uuid.UUID, filename string, reader io.Reader, size int64) (*domain.Video, error)
}

// UploadHandler exposes the video upload endpoint.
type UploadHandler struct {
	upload UploadUseCase
}

// NewUploadHandler creates an UploadHandler.
func NewUploadHandler(upload UploadUseCase) *UploadHandler {
	return &UploadHandler{upload: upload}
}

// Upload handles POST /videos.
// It reads the user identity from the X-User-ID header injected by the API gateway.
func (h *UploadHandler) Upload(c *gin.Context) {
	rawUserID := c.GetHeader("X-User-ID")
	if rawUserID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing X-User-ID header"})
		return
	}
	userID, err := uuid.Parse(rawUserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid X-User-ID header"})
		return
	}

	file, header, err := c.Request.FormFile("video")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
		return
	}
	defer file.Close()

	log.Printf("service=upload-service event=upload msg=\"recebendo upload de vídeo\" user_id=%s filename=%s size_bytes=%d",
		userID, header.Filename, header.Size)

	video, err := h.upload.Upload(c.Request.Context(), userID, header.Filename, file, header.Size)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidVideoFormat) {
			log.Printf("service=upload-service event=upload status=invalido msg=\"formato de vídeo inválido\" user_id=%s filename=%s", userID, header.Filename)
			c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
			return
		}
		log.Printf("service=upload-service event=upload status=erro msg=\"falha no upload\" user_id=%s filename=%s err=%v", userID, header.Filename, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "upload failed"})
		return
	}

	log.Printf("service=upload-service event=upload status=ok msg=\"vídeo recebido e enfileirado para processamento\" user_id=%s filename=%s video_id=%s",
		userID, header.Filename, video.ID)
	c.JSON(http.StatusAccepted, gin.H{
		"id":     video.ID,
		"status": video.Status,
	})
}
