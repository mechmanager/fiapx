// Package handler contains HTTP handlers for the auth-service.
package handler

import (
	"context"
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mechmanager/fiapx/auth-service/domain"
)

// AuthUseCase describes the auth operations used by the handler.
type AuthUseCase interface {
	Register(ctx context.Context, name, email, password string) (*domain.User, error)
	Login(ctx context.Context, email, password string) (string, error)
}

// AuthHandler exposes registration and login endpoints.
type AuthHandler struct {
	auth AuthUseCase
}

// NewAuthHandler creates an AuthHandler.
func NewAuthHandler(auth AuthUseCase) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Name     string `json:"name"     binding:"required"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// Register handles POST /auth/register.
func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("service=auth-service event=cadastro msg=\"cadastrando novo usuário\" email=%s name=%q", req.Email, req.Name)

	user, err := h.auth.Register(c.Request.Context(), req.Name, req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrEmailAlreadyExists) {
			log.Printf("service=auth-service event=cadastro status=conflito msg=\"e-mail já cadastrado\" email=%s", req.Email)
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		log.Printf("service=auth-service event=cadastro status=erro msg=\"falha ao cadastrar usuário\" email=%s err=%v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to register user"})
		return
	}

	log.Printf("service=auth-service event=cadastro status=ok msg=\"usuário cadastrado com sucesso\" user_id=%s email=%s", user.ID, user.Email)
	c.JSON(http.StatusCreated, gin.H{
		"id":    user.ID,
		"name":  user.Name,
		"email": user.Email,
	})
}

// Login handles POST /auth/login.
func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	log.Printf("service=auth-service event=login msg=\"realizando login\" email=%s", req.Email)

	token, err := h.auth.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrInvalidCredentials) {
			log.Printf("service=auth-service event=login status=falha msg=\"credenciais inválidas\" email=%s", req.Email)
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		log.Printf("service=auth-service event=login status=erro msg=\"falha interna no login\" email=%s err=%v", req.Email, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to login"})
		return
	}

	log.Printf("service=auth-service event=login status=ok msg=\"login realizado com sucesso\" email=%s", req.Email)
	c.JSON(http.StatusOK, gin.H{"token": token})
}
