// Package domain contains the core entities and repository interfaces for the auth-service.
package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// User represents a registered user.
type User struct {
	ID           uuid.UUID // unique identifier
	Name         string    // display name
	Email        string    // unique login email
	PasswordHash string    // bcrypt hash of the password
	CreatedAt    time.Time // creation timestamp
}

// UserRepository abstracts persistence operations for users.
type UserRepository interface {
	// Create persists a new user.
	Create(ctx context.Context, user *User) error
	// FindByEmail retrieves a user by email; returns ErrUserNotFound if absent.
	FindByEmail(ctx context.Context, email string) (*User, error)
}
