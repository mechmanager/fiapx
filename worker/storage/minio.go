// Package storage implementa o acesso ao object storage (MinIO/S3) para o worker.
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
func (s *MinIOStorage) Download(ctx context.Context, objectKey string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucket, objectKey, minio.GetObjectOptions{})
}

// UploadFile envia um arquivo do disco para o MinIO pelo caminho local.
// FPutObject determina o tamanho automaticamente.
func (s *MinIOStorage) UploadFile(ctx context.Context, objectKey, filePath string) error {
	_, err := s.client.FPutObject(ctx, s.bucket, objectKey, filePath, minio.PutObjectOptions{
		ContentType: "application/zip",
	})
	return err
}
