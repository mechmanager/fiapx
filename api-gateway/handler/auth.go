// Package handler contém os handlers HTTP (camada de entrada) do api-gateway.
package handler

import (
	"context"
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mechmanager/fiapx/api-gateway/domain"
)

// AuthUseCase descreve as operações de autenticação usadas pelo handler.
// Implementado por service.AuthService e mockado nos testes.
type AuthUseCase interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

// AuthHandler expõe os endpoints de cadastro e login.
type AuthHandler struct {
	auth AuthUseCase // lógica de autenticação
}

// NewAuthHandler cria o handler de autenticação.
func NewAuthHandler(auth AuthUseCase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

// registerRequest representa o corpo da requisição de cadastro.
type registerRequest struct {
	Name     string `json:"name" binding:"required"`           // nome do usuário
	Email    string `json:"email" binding:"required,email"`    // email válido
	Password string `json:"password" binding:"required,min=6"` // senha com no mínimo 6 caracteres
}

// loginRequest representa o corpo da requisição de login.
type loginRequest struct {
	Email    string `json:"email" binding:"required,email"` // email do usuário
	Password string `json:"password" binding:"required"`    // senha em texto puro
}

// Register trata POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	// Faz o bind e a validação do corpo da requisição.
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delega o cadastro ao serviço de autenticação.
	user, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		// Email duplicado vira 409 Conflict; demais erros viram 500.
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao cadastrar usuário"})
		return
	}

	// Devolve os dados públicos do usuário criado.
	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

// Login trata POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	// Faz o bind e a validação do corpo da requisição.
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Delega a autenticação ao serviço.
	token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		// Credenciais inválidas viram 401; demais erros viram 500.
		if errors.Is(err, domain.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "erro ao realizar login"})
		return
	}

	// Devolve o token JWT.
	c.JSON(http.StatusOK, gin.H{"token": token})
}
