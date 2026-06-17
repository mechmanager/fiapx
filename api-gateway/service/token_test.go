package service_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/api-gateway/service"
)

func TestJWTManager_GenerateAndValidate(t *testing.T) {
	m := service.NewJWTManager("segredo", 1)
	id := uuid.New()

	token, err := m.Generate(id)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}

	got, err := m.Validate(token)
	if err != nil {
		t.Fatalf("Validate: %v", err)
	}
	if got != id {
		t.Errorf("esperava %v, obteve %v", id, got)
	}
}

func TestJWTManager_InvalidToken(t *testing.T) {
	m := service.NewJWTManager("segredo", 1)
	_, err := m.Validate("token-invalido")
	if err == nil {
		t.Fatal("esperava erro para token inválido")
	}
}

func TestJWTManager_WrongSecret(t *testing.T) {
	m1 := service.NewJWTManager("segredo1", 1)
	m2 := service.NewJWTManager("segredo2", 1)

	token, _ := m1.Generate(uuid.New())
	_, err := m2.Validate(token)
	if err == nil {
		t.Fatal("esperava erro ao validar com segredo diferente")
	}
}
