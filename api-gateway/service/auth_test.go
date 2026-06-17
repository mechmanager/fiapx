package service_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/domain"
	"github.com/mechmanager/fiapx/api-gateway/mocks"
	"github.com/mechmanager/fiapx/api-gateway/service"
)

func newJWT() *service.JWTManager {
	return service.NewJWTManager("segredo-de-teste", 1)
}

// --- Register ---

func TestRegister_HappyPath(t *testing.T) {
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
		CreateFn: func(_ context.Context, u *domain.User) error {
			return nil
		},
	}
	svc := service.NewAuthService(repo, newJWT())
	user, err := svc.Register(context.Background(), "Alice", "alice@example.com", "senha123")
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if user.Email != "alice@example.com" {
		t.Errorf("email incorreto: %q", user.Email)
	}
}

func TestRegister_EmailDuplicate(t *testing.T) {
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return &domain.User{ID: uuid.New()}, nil
		},
	}
	svc := service.NewAuthService(repo, newJWT())
	_, err := svc.Register(context.Background(), "Alice", "alice@example.com", "senha123")
	if !errors.Is(err, domain.ErrEmailAlreadyExists) {
		t.Fatalf("esperava ErrEmailAlreadyExists, obteve %v", err)
	}
}

func TestRegister_FindByEmailError(t *testing.T) {
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, errors.New("falha no banco")
		},
	}
	svc := service.NewAuthService(repo, newJWT())
	_, err := svc.Register(context.Background(), "Alice", "alice@example.com", "senha123")
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestRegister_CreateError(t *testing.T) {
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
		CreateFn: func(_ context.Context, _ *domain.User) error {
			return errors.New("falha ao inserir")
		},
	}
	svc := service.NewAuthService(repo, newJWT())
	_, err := svc.Register(context.Background(), "Alice", "alice@example.com", "senha123")
	if err == nil {
		t.Fatal("esperava erro")
	}
}

// --- Login ---

func TestLogin_HappyPath(t *testing.T) {
	// Pré-cadastra um usuário com hash real de bcrypt para o Login poder verificar.
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
		CreateFn: func(_ context.Context, _ *domain.User) error { return nil },
	}
	jwtM := newJWT()
	svc := service.NewAuthService(repo, jwtM)

	// Cadastra para obter o hash.
	var savedUser *domain.User
	repo.CreateFn = func(_ context.Context, u *domain.User) error {
		savedUser = u
		return nil
	}
	_, _ = svc.Register(context.Background(), "Alice", "alice@example.com", "senha123")

	// Agora faz o login usando o hash salvo.
	repo.FindByEmailFn = func(_ context.Context, _ string) (*domain.User, error) {
		return savedUser, nil
	}
	token, err := svc.Login(context.Background(), "alice@example.com", "senha123")
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if token == "" {
		t.Fatal("token não pode ser vazio")
	}
}

func TestLogin_UserNotFound(t *testing.T) {
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, domain.ErrUserNotFound
		},
	}
	svc := service.NewAuthService(repo, newJWT())
	_, err := svc.Login(context.Background(), "x@x.com", "abc")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("esperava ErrInvalidCredentials, obteve %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			// Hash de "correta", mas tentaremos logar com "errada".
			return nil, domain.ErrUserNotFound
		},
		CreateFn: func(_ context.Context, _ *domain.User) error { return nil },
	}
	jwtM := newJWT()
	svc := service.NewAuthService(repo, jwtM)
	var saved *domain.User
	repo.CreateFn = func(_ context.Context, u *domain.User) error { saved = u; return nil }
	_, _ = svc.Register(context.Background(), "Alice", "a@b.com", "correta")

	repo.FindByEmailFn = func(_ context.Context, _ string) (*domain.User, error) {
		return saved, nil
	}
	_, err := svc.Login(context.Background(), "a@b.com", "errada")
	if !errors.Is(err, domain.ErrInvalidCredentials) {
		t.Fatalf("esperava ErrInvalidCredentials, obteve %v", err)
	}
}

func TestLogin_RepoError(t *testing.T) {
	repo := &mocks.UserRepository{
		FindByEmailFn: func(_ context.Context, _ string) (*domain.User, error) {
			return nil, errors.New("timeout")
		},
	}
	svc := service.NewAuthService(repo, newJWT())
	_, err := svc.Login(context.Background(), "a@b.com", "123456")
	if err == nil {
		t.Fatal("esperava erro")
	}
}
