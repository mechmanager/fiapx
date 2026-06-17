// Package repository implementa o acesso aos dados em PostgreSQL.
package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mechmanager/fiapx/api-gateway/domain"
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

// UserRepo implementa domain.UserRepository usando PostgreSQL.
type UserRepo struct {
	db DB
}

// NewUserRepo cria o repositório de usuários.
func NewUserRepo(db DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create insere um novo usuário no banco.
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (id, name, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())`
	_, err := r.db.Exec(ctx, query, user.ID, user.Name, user.Email, user.PasswordHash)
	return err
}

// FindByEmail busca um usuário pelo email; retorna ErrUserNotFound se não existir.
func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*domain.User, error) {
	const query = `
		SELECT id, name, email, password_hash, created_at
		FROM users WHERE email = $1`
	row := r.db.QueryRow(ctx, query, email)
	user := &domain.User{}
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}
	return user, nil
}

// VideoRepo implementa domain.VideoRepository usando PostgreSQL.
type VideoRepo struct {
	db DB
}

// NewVideoRepo cria o repositório de vídeos.
func NewVideoRepo(db DB) *VideoRepo {
	return &VideoRepo{db: db}
}

// Create insere um novo vídeo no banco com status PENDING.
func (r *VideoRepo) Create(ctx context.Context, video *domain.Video) error {
	const query = `
		INSERT INTO videos (id, user_id, original_filename, s3_key, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())`
	_, err := r.db.Exec(ctx, query,
		video.ID, video.UserID, video.OriginalFilename, video.S3Key, video.Status)
	return err
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

// FindByID busca um vídeo pelo ID.
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
