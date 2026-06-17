// Package repository implements PostgreSQL-backed data access for the auth-service.
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mechmanager/fiapx/auth-service/domain"
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

// UserRepo implements domain.UserRepository using PostgreSQL.
type UserRepo struct {
	db DB
}

// NewUserRepo creates a new UserRepo.
func NewUserRepo(db DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create inserts a new user into the database.
func (r *UserRepo) Create(ctx context.Context, user *domain.User) error {
	const query = `
		INSERT INTO users (id, name, email, password_hash, created_at)
		VALUES ($1, $2, $3, $4, NOW())`
	_, err := r.db.Exec(ctx, query, user.ID, user.Name, user.Email, user.PasswordHash)
	return err
}

// FindByEmail retrieves a user by email; returns ErrUserNotFound on pgx.ErrNoRows.
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
