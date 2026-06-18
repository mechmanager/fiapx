package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", handler.Health)
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	srv := &http.Server{
		Addr:         ":" + cfg.AuthPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("auth-service listening on :%s", cfg.AuthPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
