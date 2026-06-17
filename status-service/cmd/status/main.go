package main

import (
	"context"
	"log"

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

	// Pool de conexões com o PostgreSQL.
	pool, err := repository.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer pool.Close()

	// Repositório de vídeos.
	videoRepo := repository.NewVideoRepo(pool)

	// Object storage (MinIO).
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
		cfg.MinIOBucket, cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Fatalf("erro ao conectar no MinIO: %v", err)
	}

	// Cache Redis.
	redisCache, err := cache.NewRedisCache(cfg.RedisURL)
	if err != nil {
		log.Fatalf("erro ao conectar no Redis: %v", err)
	}

	// Serviço e handler.
	statusSvc := service.NewStatusService(videoRepo, minioStorage, redisCache)
	statusHandler := handler.NewStatusHandler(statusSvc)

	// Roteador Gin.
	r := gin.Default()

	r.GET("/health", handler.Health)
	r.GET("/videos", statusHandler.List)
	r.GET("/videos/:id/download", statusHandler.Download)

	log.Printf("status-service escutando na porta %s", cfg.StatusPort)
	if err := r.Run(":" + cfg.StatusPort); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
