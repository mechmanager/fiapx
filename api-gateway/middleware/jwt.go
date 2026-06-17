package middleware

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrInvalidToken = errors.New("token inválido")

// JWTChecker valida o token e injeta X-User-ID no request antes de prosseguir.
type JWTChecker struct {
	secret []byte
}

func NewJWTChecker(secret string) *JWTChecker {
	return &JWTChecker{secret: []byte(secret)}
}

func (j *JWTChecker) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		header := r.Header.Get("Authorization")
		if header == "" {
			http.Error(w, `{"error":"token de autenticação ausente"}`, http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			http.Error(w, `{"error":"formato de token inválido"}`, http.StatusUnauthorized)
			return
		}

		userID, err := j.validate(parts[1])
		if err != nil {
			http.Error(w, `{"error":"token inválido ou expirado"}`, http.StatusUnauthorized)
			return
		}

		r.Header.Set("X-User-ID", userID.String())
		next.ServeHTTP(w, r)
	})
}

func (j *JWTChecker) validate(tokenString string) (uuid.UUID, error) {
	claims := &jwt.RegisteredClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return j.secret, nil
	})
	if err != nil || !token.Valid {
		return uuid.Nil, ErrInvalidToken
	}
	if claims.ExpiresAt != nil && claims.ExpiresAt.Before(time.Now()) {
		return uuid.Nil, ErrInvalidToken
	}
	userID, err := uuid.Parse(claims.Subject)
	if err != nil {
		return uuid.Nil, ErrInvalidToken
	}
	return userID, nil
}
