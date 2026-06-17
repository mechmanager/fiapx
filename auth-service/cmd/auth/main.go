// Command auth is the entry point for the auth-service.
package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/mechmanager/fiapx/auth-service/config"
	"github.com/mechmanager/fiapx/auth-service/handler"
	"github.com/mechmanager/fiapx/auth-service/repository"
	"github.com/mechmanager/fiapx/auth-service/service"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	pool, err := repository.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepo(pool)
	jwtManager := service.NewJWTManager(cfg.JWTSecret, cfg.JWTExpirationHours)
	authSvc := service.NewAuthService(userRepo, jwtManager)
	authHandler := handler.NewAuthHandler(authSvc)

	r := gin.Default()

	r.GET("/health", handler.Health)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	log.Printf("auth-service listening on :%s", cfg.AuthPort)
	if err := r.Run(":" + cfg.AuthPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
