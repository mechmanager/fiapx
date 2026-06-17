// Package repository implementa o acesso aos dados em PostgreSQL.
package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
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
	pool *pgxpool.Pool
}

// NewUserRepo cria o repositório de usuários.
func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

// FindEmailByUserID retorna o e-mail do usuário pelo ID.
// Retorna "" e nil se o usuário não for encontrado (graceful).
func (r *UserRepo) FindEmailByUserID(ctx context.Context, userID uuid.UUID) (string, error) {
	const query = `SELECT email FROM users WHERE id = $1`
	var email string
	err := r.pool.QueryRow(ctx, query, userID).Scan(&email)
	if err != nil {
		// Graceful: usuário não encontrado não é um erro fatal para notificação.
		return "", nil
	}
	return email, nil
}
