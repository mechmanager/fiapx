// Package repository implementa o acesso aos dados em PostgreSQL.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mechmanager/fiapx/status-service/domain"
)

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

// VideoRepo implementa domain.VideoRepository usando PostgreSQL.
type VideoRepo struct {
	db DB
}

// NewVideoRepo cria o repositório de vídeos.
func NewVideoRepo(db DB) *VideoRepo {
	return &VideoRepo{db: db}
}

// ListByUser retorna os vídeos de um usuário ordenados do mais recente ao mais antigo.
func (r *VideoRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Video, error) {
	const query = `
		SELECT id, user_id, original_filename, s3_key, status,
		       COALESCE(error_message,''), COALESCE(zip_s3_key,''),
		       frame_count, created_at, updated_at
		FROM videos WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var videos []*domain.Video
	for rows.Next() {
		v := &domain.Video{}
		if err := rows.Scan(
			&v.ID, &v.UserID, &v.OriginalFilename, &v.S3Key, &v.Status,
			&v.ErrorMessage, &v.ZipS3Key, &v.FrameCount, &v.CreatedAt, &v.UpdatedAt,
		); err != nil {
			return nil, err
		}
		videos = append(videos, v)
	}
	return videos, rows.Err()
}

// Delete remove o registro do banco. Retorna ErrVideoNotFound se não existir.
func (r *VideoRepo) Delete(ctx context.Context, id uuid.UUID) error {
	tag, err := r.db.Exec(ctx, `DELETE FROM videos WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrVideoNotFound
	}
	return nil
}

// FindByID busca um vídeo pelo ID; retorna ErrVideoNotFound se não existir.
func (r *VideoRepo) FindByID(ctx context.Context, id uuid.UUID) (*domain.Video, error) {
	const query = `
		SELECT id, user_id, original_filename, s3_key, status,
		       COALESCE(error_message,''), COALESCE(zip_s3_key,''),
		       frame_count, created_at, updated_at
		FROM videos WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)
	v := &domain.Video{}
	err := row.Scan(
		&v.ID, &v.UserID, &v.OriginalFilename, &v.S3Key, &v.Status,
		&v.ErrorMessage, &v.ZipS3Key, &v.FrameCount, &v.CreatedAt, &v.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrVideoNotFound
	}
	if err != nil {
		return nil, err
	}
	return v, nil
}
