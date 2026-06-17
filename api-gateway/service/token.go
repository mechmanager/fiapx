// Package service contém a lógica de negócio do api-gateway.
package service

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// ErrInvalidToken indica que o token JWT é inválido ou expirou.
var ErrInvalidToken = errors.New("token inválido")

// JWTManager gera e valida tokens JWT assinados com HMAC.
type JWTManager struct {
	secret     []byte        // segredo usado na assinatura
	expiration time.Duration // tempo de validade do token
}

// NewJWTManager cria um gerenciador de tokens com o segredo e validade informados.
func NewJWTManager(secret string, expirationHours int) *JWTManager {
	return &JWTManager{
		secret:     []byte(secret),
		expiration: time.Duration(expirationHours) * time.Hour,
	}
}

// Generate cria um token JWT contendo o ID do usuário no campo subject.
func (manager *JWTManager) Generate(userID uuid.UUID) (string, error) {
	// Define as claims padrão com expiração e identificação do usuário.
	claims := jwt.RegisteredClaims{
		Subject:   userID.String(),
		IssuedAt:  jwt.NewNumericDate(time.Now()),
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(manager.expiration)),
	}

	// Cria o token usando HMAC-SHA256.
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// Assina o token com o segredo configurado.
	signed, err := token.SignedString(manager.secret)
	if err != nil {
		return "", err
	}

	return signed, nil
}

// Validate verifica a assinatura e a validade do token e devolve o ID do usuário.
func (manager *JWTManager) Validate(tokenString string) (uuid.UUID, error) {
	// Faz o parse validando o método de assinatura esperado.
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		// Garante que o algoritmo de assinatura é o esperado (HMAC).
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return manager.secret, nil
	})

	// Rejeita tokens com erro de parse ou inválidos.
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}

	// Converte o subject (string) de volta para UUID.
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}

	return userID, nil
}
