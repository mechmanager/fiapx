// Package storage implementa o acesso ao object storage (MinIO/S3).
package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// MinIOStorage implementa domain.ObjectStorage usando MinIO.
type MinIOStorage struct {
	client *minio.Client
	bucket string
}

// NewMinIOStorage cria o cliente MinIO.
func NewMinIOStorage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOStorage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &MinIOStorage{client: client, bucket: bucket}, nil
}

// Download devolve um leitor para o conteúdo de um objeto armazenado.
func (s *MinIOStorage) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	return obj, nil
}

// Delete remove um ou mais objetos do bucket. Erros individuais são ignorados (best-effort).
func (s *MinIOStorage) Delete(ctx context.Context, keys ...string) error {
	for _, key := range keys {
		if key == "" {
			continue
		}
		_ = s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	}
	return nil
}
