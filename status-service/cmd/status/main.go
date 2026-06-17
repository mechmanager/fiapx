package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/mechmanager/fiapx/status-service/cache"
	"github.com/mechmanager/fiapx/status-service/config"
	"github.com/mechmanager/fiapx/status-service/handler"
	"github.com/mechmanager/fiapx/status-service/repository"
	"github.com/mechmanager/fiapx/status-service/service"
	"github.com/mechmanager/fiapx/status-service/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	ctx := context.Background()

	pool, err := repository.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer pool.Close()

	videoRepo := repository.NewVideoRepo(pool)

	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
		cfg.MinIOBucket, cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Fatalf("erro ao conectar no MinIO: %v", err)
	}

	redisCache, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		log.Fatalf("erro ao conectar no Redis: %v", err)
	}

	statusSvc := service.NewStatusService(videoRepo, minioStorage, redisCache)
	statusHandler := handler.NewStatusHandler(statusSvc)

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", handler.Health)
	r.GET("/videos", statusHandler.List)
	r.GET("/videos/:id/download", statusHandler.Download)

	srv := &http.Server{
		Addr:         ":" + cfg.StatusPort,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 120 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("status-service escutando na porta %s", cfg.StatusPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
