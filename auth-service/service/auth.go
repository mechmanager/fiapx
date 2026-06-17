// Package service contains the business logic for the auth-service.
package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mechmanager/fiapx/auth-service/domain"
)

// AuthService implements registration and login logic.
type AuthService struct {
	users  domain.UserRepository
	tokens *JWTManager
}

// NewAuthService creates an AuthService with its dependencies.
func NewAuthService(users domain.UserRepository, tokens *JWTManager) *AuthService {
	return &AuthService{users: users, tokens: tokens}
}

// Register creates a new user after validating email uniqueness and hashing the password.
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, err
	}
	if existing != nil {
		return nil, domain.ErrEmailAlreadyExists
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), 12)
	if err != nil {
		return nil, err
	}

	user := &domain.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login validates credentials and returns a signed JWT on success.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	email = strings.ToLower(strings.TrimSpace(email))

	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
