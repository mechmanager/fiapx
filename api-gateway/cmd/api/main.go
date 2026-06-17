package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/mechmanager/fiapx/api-gateway/config"
	"github.com/mechmanager/fiapx/api-gateway/handler"
	"github.com/mechmanager/fiapx/api-gateway/middleware"
	"github.com/mechmanager/fiapx/api-gateway/queue"
	"github.com/mechmanager/fiapx/api-gateway/repository"
	"github.com/mechmanager/fiapx/api-gateway/service"
	"github.com/mechmanager/fiapx/api-gateway/storage"
	apiweb "github.com/mechmanager/fiapx/api-gateway/web"
)

func main() {
	// Carrega configurações a partir de variáveis de ambiente.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	ctx := context.Background()

	// Pool de conexões com o PostgreSQL.
	pool, err := repository.NewPool(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer pool.Close()

	// Repositórios.
	userRepo := repository.NewUserRepo(pool)
	videoRepo := repository.NewVideoRepo(pool)

	// Object storage (MinIO).
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
		cfg.MinIOBucket, cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Fatalf("erro ao conectar no MinIO: %v", err)
	}

	// Fila (RabbitMQ).
	publisher, err := queue.NewRabbitMQPublisher(cfg.RabbitMQURL, cfg.QueueName)
	if err != nil {
		log.Fatalf("erro ao conectar no RabbitMQ: %v", err)
	}
	defer publisher.Close()

	// Serviços.
	jwtManager := service.NewJWTManager(cfg.JWTSecret, cfg.JWTExpirationHours)
	authService := service.NewAuthService(userRepo, jwtManager)
	videoService := service.NewVideoService(videoRepo, minioStorage, publisher)

	// Handlers.
	authHandler := handler.NewAuthHandler(authService)
	videoHandler := handler.NewVideoHandler(videoService)

	// Roteador Gin.
	r := gin.Default()

	// Aumenta o limite de upload para 500 MB.
	r.MaxMultipartMemory = 500 << 20

	// Frontend estático embutido no binário.
	r.GET("/", func(c *gin.Context) {
		c.FileFromFS("index.html", http.FS(apiweb.FS))
	})

	// Rotas públicas.
	r.GET("/health", handler.Health)
	r.POST("/auth/register", authHandler.Register)
	r.POST("/auth/login", authHandler.Login)

	// Rotas protegidas por JWT.
	protected := r.Group("/", middleware.JWTAuth(jwtManager))
	{
		protected.POST("/videos", videoHandler.Upload)
		protected.GET("/videos", videoHandler.List)
		protected.GET("/videos/:id/download", videoHandler.Download)
	}

	log.Printf("api-gateway escutando na porta %s", cfg.APIPort)
	if err := r.Run(":" + cfg.APIPort); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
