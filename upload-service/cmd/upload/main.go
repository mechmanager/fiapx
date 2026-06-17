// Command upload is the entry point for the upload-service.
package main

import (
	"context"
	"log"

	"github.com/gin-gonic/gin"

	"github.com/mechmanager/fiapx/upload-service/config"
	"github.com/mechmanager/fiapx/upload-service/handler"
	"github.com/mechmanager/fiapx/upload-service/queue"
	"github.com/mechmanager/fiapx/upload-service/repository"
	"github.com/mechmanager/fiapx/upload-service/service"
	"github.com/mechmanager/fiapx/upload-service/storage"
)

func main() {
	cfg := config.Load()

	ctx := context.Background()

	pool, err := repository.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("failed to connect to postgres: %v", err)
	}
	defer pool.Close()

	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint,
		cfg.MinIOAccessKey,
		cfg.MinIOSecretKey,
		cfg.MinioBucket,
		cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Fatalf("failed to create minio storage: %v", err)
	}

	publisher, err := queue.NewRabbitMQPublisher(cfg.RabbitMQURL, cfg.QueueName)
	if err != nil {
		log.Fatalf("failed to create rabbitmq publisher: %v", err)
	}
	defer publisher.Close()

	videoRepo := repository.NewVideoRepo(pool)
	uploadSvc := service.NewUploadService(videoRepo, minioStorage, publisher)
	uploadHandler := handler.NewUploadHandler(uploadSvc)

	r := gin.Default()

	r.GET("/health", handler.Health)
	r.POST("/videos", uploadHandler.Upload)

	log.Printf("upload-service listening on :%s", cfg.UploadPort)
	if err := r.Run(":" + cfg.UploadPort); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
