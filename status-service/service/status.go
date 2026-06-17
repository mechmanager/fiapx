// Package service implementa a lógica de negócio do status-service.
package service

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/status-service/domain"
)

const statusCacheTTL = 30 * time.Second

// StatusService implementa as operações de consulta e download de vídeos.
type StatusService struct {
	repo    domain.VideoRepository
	storage domain.ObjectStorage
	cache   domain.StatusCache
}

// NewStatusService cria o serviço de status.
func NewStatusService(repo domain.VideoRepository, storage domain.ObjectStorage, cache domain.StatusCache) *StatusService {
	return &StatusService{repo: repo, storage: storage, cache: cache}
}

// ListByUser retorna os vídeos do usuário com status possivelmente servido do cache.
func (s *StatusService) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error) {
	videos, err := s.repo.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	for _, v := range videos {
		key := fmt.Sprintf("video:%s:status", v.ID.String())
		cached, err := s.cache.Get(ctx, key)
		if err == nil && cached != "" {
			v.Status = cached
		} else {
			// Cache miss: popula o cache com o status atual do banco.
			_ = s.cache.Set(ctx, key, v.Status, statusCacheTTL)
		}
	}

	return videos, nil
}

// GetDownloadStream valida a propriedade do vídeo e retorna o stream do zip de frames.
func (s *StatusService) GetDownloadStream(ctx context.Context, userID, videoID uuid.UUID) (io.ReadCloser, error) {
	video, err := s.repo.FindByID(ctx, videoID)
	if err != nil {
		return nil, err
	}

	if video.UserID != userID {
		return nil, &ownershipError{}
	}

	if video.Status != "DONE" || video.ZipS3Key == "" {
		return nil, domain.ErrVideoNotReady
	}

	return s.storage.Download(ctx, video.ZipS3Key)
}

// ownershipError indica que o vídeo pertence a outro usuário.
type ownershipError struct{}

func (e *ownershipError) Error() string { return "acesso negado" }

// StatusCode devolve o HTTP status code adequado para o erro.
func StatusCode(err error) int {
	if err == nil {
		return http.StatusOK
	}
	switch err.(type) {
	case *ownershipError:
		return http.StatusForbidden
	}
	return http.StatusInternalServerError
}

// IsOwnershipError informa se o erro é de propriedade.
func IsOwnershipError(err error) bool {
	_, ok := err.(*ownershipError)
	return ok
}
