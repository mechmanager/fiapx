// Package middleware contém os middlewares HTTP do api-gateway.
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// ContextUserIDKey é a chave usada para guardar o ID do usuário no contexto do Gin.
const ContextUserIDKey = "userID"

// Authenticator define o que o middleware precisa para validar tokens.
// O JWTManager satisfaz esta interface, permitindo mocks nos testes.
type Authenticator interface {
	Validate(tokenString string) (uuid.UUID, error)
}

// JWTAuth devolve um middleware que exige um token JWT válido no header Authorization.
func JWTAuth(authenticator Authenticator) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Lê o header Authorization no formato "Bearer <token>".
		header := c.GetHeader("Authorization")
		if header == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token de autenticação ausente"})
			return
		}

		// Separa o prefixo "Bearer" do token propriamente dito.
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "formato de token inválido"})
			return
		}

		// Valida o token e extrai o ID do usuário.
		userID, err := authenticator.Validate(parts[1])
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token inválido ou expirado"})
			return
		}

		// Disponibiliza o ID do usuário para os handlers seguintes.
		c.Set(ContextUserIDKey, userID)
		c.Next()
	}
}
