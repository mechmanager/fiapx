package service

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"github.com/mechmanager/fiapx/api-gateway/domain"
)

// AuthService implementa as regras de autenticação (cadastro e login).
type AuthService struct {
	users  domain.UserRepository // acesso aos dados de usuários
	tokens *JWTManager           // geração de tokens JWT
}

// NewAuthService cria o serviço de autenticação com suas dependências.
func NewAuthService(users domain.UserRepository, tokens *JWTManager) *AuthService {
	return &AuthService{
		users:  users,
		tokens: tokens,
	}
}

// Register cadastra um novo usuário, validando email único e aplicando hash na senha.
func (s *AuthService) Register(ctx context.Context, name, email, password string) (*domain.User, error) {
	// Normaliza o email para evitar duplicidade por diferença de caixa/espaços.
	email = strings.ToLower(strings.TrimSpace(email))

	// Verifica se já existe usuário com o mesmo email.
	existing, err := s.users.FindByEmail(ctx, email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		// Falha inesperada no repositório.
		return nil, err
	}
	if existing != nil {
		// Email já está em uso.
		return nil, domain.ErrEmailAlreadyExists
	}

	// Gera o hash bcrypt da senha em texto puro.
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Monta a entidade de usuário com um novo identificador.
	user := &domain.User{
		ID:           uuid.New(),
		Name:         name,
		Email:        email,
		PasswordHash: string(hash),
	}

	// Persiste o usuário no repositório.
	if err := s.users.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login valida as credenciais e devolve um token JWT em caso de sucesso.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	// Normaliza o email da mesma forma que no cadastro.
	email = strings.ToLower(strings.TrimSpace(email))

	// Busca o usuário pelo email.
	user, err := s.users.FindByEmail(ctx, email)
	if err != nil {
		// Não diferencia "não encontrado" de "senha errada" para evitar enumeração de usuários.
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrInvalidCredentials
		}
		return "", err
	}

	// Compara a senha informada com o hash armazenado.
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", domain.ErrInvalidCredentials
	}

	// Gera o token JWT para o usuário autenticado.
	token, err := s.tokens.Generate(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}
