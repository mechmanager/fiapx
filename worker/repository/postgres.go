// Package repository implementa o acesso ao banco de dados para o worker.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// VideoRepo implementa domain.VideoRepository usando PostgreSQL.
type VideoRepo struct {
	pool *pgxpool.Pool
}

// NewVideoRepo cria o repositório de vídeos.
func NewVideoRepo(pool *pgxpool.Pool) *VideoRepo {
	return &VideoRepo{pool: pool}
}

// NewPool cria um pool de conexões com o PostgreSQL e valida a conectividade.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// UpdateStatus atualiza o status e a mensagem de erro de um vídeo.
func (r *VideoRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status, errorMsg string) error {
	const query = `
		UPDATE videos
		SET status = $2, error_message = NULLIF($3, ''), updated_at = NOW()
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, status, errorMsg)
	return err
}

// UpdateDone marca o vídeo como processado com sucesso, gravando a chave do zip e o total de frames.
func (r *VideoRepo) UpdateDone(ctx context.Context, id uuid.UUID, zipS3Key string, frameCount int) error {
	const query = `
		UPDATE videos
		SET status = 'DONE', zip_s3_key = $2, frame_count = $3, updated_at = NOW()
		WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, id, zipS3Key, frameCount)
	return err
}
