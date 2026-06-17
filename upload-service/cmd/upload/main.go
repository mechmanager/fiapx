package main

import (
	"context"
	"log"
	"net/http"
	"time"

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

	r := gin.New()
	r.Use(gin.Recovery())

	r.GET("/health", handler.Health)
	r.POST("/videos", uploadHandler.Upload)

	srv := &http.Server{
		Addr:         ":" + cfg.UploadPort,
		Handler:      r,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("upload-service listening on :%s", cfg.UploadPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
