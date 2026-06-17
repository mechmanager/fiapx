// Package repository implements PostgreSQL-backed data access for the upload-service.
package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mechmanager/fiapx/upload-service/domain"
)

// NewPool creates a connection pool to PostgreSQL and validates connectivity.
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

// VideoRepo implements domain.VideoRepository using PostgreSQL.
type VideoRepo struct {
	db DB
}

// NewVideoRepo creates a new VideoRepo.
func NewVideoRepo(db DB) *VideoRepo {
	return &VideoRepo{db: db}
}

// Create inserts a new video record with status PENDING.
func (r *VideoRepo) Create(ctx context.Context, video *domain.Video) error {
	const query = `
		INSERT INTO videos (id, user_id, original_filename, s3_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
	_, err := r.db.Exec(ctx, query,
		video.ID, video.UserID, video.OriginalFilename, video.S3Key, video.Status)
	return err
}
