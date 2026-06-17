package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User representa um usuário cadastrado no sistema.
type User struct {
	ID           uuid.UUID // identificador único
	Name         string    // nome do usuário
	Email        string    // email único usado no login
	PasswordHash string    // hash bcrypt da senha
	CreatedAt    time.Time // data de criação
}

// UserRepository abstrai o acesso aos dados de usuários.
// Implementado pelo PostgreSQL e mockado nos testes.
type UserRepository interface {
	// Create persiste um novo usuário no banco.
	Create(ctx context.Context, user *User) error
	// FindByEmail busca um usuário pelo email; retorna ErrUserNotFound se não existir.
	FindByEmail(ctx context.Context, email string) (*User, error)
}
