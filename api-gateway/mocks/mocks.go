// Package mocks contém implementações falsas das interfaces de domínio para uso em testes.
package mocks

import (
	"context"
	"io"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/domain"
)

// --- UserRepository ---

type UserRepository struct {
	CreateFn      func(ctx context.Context, user *domain.User) error
	FindByEmailFn func(ctx context.Context, email string) (*domain.User, error)
}

func (m *UserRepository) Create(ctx context.Context, user *domain.User) error {
	return m.CreateFn(ctx, user)
}
func (m *UserRepository) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.FindByEmailFn(ctx, email)
}

// --- VideoRepository ---

type VideoRepository struct {
	CreateFn     func(ctx context.Context, video *domain.Video) error
	ListByUserFn func(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error)
	FindByIDFn   func(ctx context.Context, id uuid.UUID) (*domain.Video, error)
}

func (m *VideoRepository) Create(ctx context.Context, video *domain.Video) error {
	return m.CreateFn(ctx, video)
}
func (m *VideoRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error) {
	return m.ListByUserFn(ctx, userID)
}
func (m *VideoRepository) FindByID(ctx context.Context, id uuid.UUID) (*domain.Video, error) {
	return m.FindByIDFn(ctx, id)
}

// --- ObjectStorage ---

type ObjectStorage struct {
	UploadFn   func(ctx context.Context, key string, r io.Reader, size int64, ct string) error
	DownloadFn func(ctx context.Context, key string) (io.ReadCloser, error)
}

func (m *ObjectStorage) Upload(ctx context.Context, key string, r io.Reader, size int64, ct string) error {
	return m.UploadFn(ctx, key, r, size, ct)
}
func (m *ObjectStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	return m.DownloadFn(ctx, key)
}

// --- QueuePublisher ---

type QueuePublisher struct {
	PublishFn func(ctx context.Context, body []byte) error
}

func (m *QueuePublisher) Publish(ctx context.Context, body []byte) error {
	return m.PublishFn(ctx, body)
}

// --- Authenticator (para middleware) ---

type Authenticator struct {
	ValidateFn func(token string) (uuid.UUID, error)
}

func (m *Authenticator) Validate(token string) (uuid.UUID, error) {
	return m.ValidateFn(token)
}

// --- AuthUseCase (para handler) ---

type AuthUseCase struct {
	RegisterFn func(ctx context.Context, name, email, password string) (*domain.User, error)
	LoginFn    func(ctx context.Context, email, password string) (string, error)
}

func (m *AuthUseCase) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	return m.RegisterFn(ctx, name, email, password)
}
func (m *AuthUseCase) Login(ctx context.Context, email, password string) (string, error) {
	return m.LoginFn(ctx, email, password)
}

// --- VideoUseCase (para handler) ---

type VideoUseCase struct {
	UploadFn            func(ctx context.Context, userID uuid.UUID, filename string, r io.Reader, size int64) (*domain.Video, error)
	ListByUserFn        func(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error)
	GetDownloadStreamFn func(ctx context.Context, userID, videoID uuid.UUID) (io.ReadCloser, error)
}

func (m *VideoUseCase) Upload(ctx context.Context, userID uuid.UUID, filename string, r io.Reader, size int64) (*domain.Video, error) {
	return m.UploadFn(ctx, userID, filename, r, size)
}
func (m *VideoUseCase) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error) {
	return m.ListByUserFn(ctx, userID)
}
func (m *VideoUseCase) GetDownloadStream(ctx context.Context, userID, videoID uuid.UUID) (io.ReadCloser, error) {
	return m.GetDownloadStreamFn(ctx, userID, videoID)
}
