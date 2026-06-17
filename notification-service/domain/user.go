// Package domain contém os tipos e interfaces de domínio do notification-service.
package domain

import (
	"context"

	"github.com/google/uuid"
)

// User representa um usuário do sistema.
type User struct {
	ID    uuid.UUID
	Email string
}

// UserRepository define as operações de leitura de usuários no banco de dados.
type UserRepository interface {
	FindEmailByUserID(ctx context.Context, userID uuid.UUID) (string, error)
}
