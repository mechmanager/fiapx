package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/mechmanager/fiapx/worker/config"
	"github.com/mechmanager/fiapx/worker/consumer"
	"github.com/mechmanager/fiapx/worker/notification"
	"github.com/mechmanager/fiapx/worker/processor"
	"github.com/mechmanager/fiapx/worker/repository"
	"github.com/mechmanager/fiapx/worker/storage"
)

func main() {
	// Carrega configurações a partir de variáveis de ambiente.
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	// Pool de conexões com o PostgreSQL.
	pool, err := repository.NewPool(ctx, cfg.DBDSN)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer pool.Close()

	videoRepo := repository.NewVideoRepo(pool)

	// Object storage (MinIO).
	minioStorage, err := storage.NewMinIOStorage(
		cfg.MinIOEndpoint, cfg.MinIOAccessKey, cfg.MinIOSecretKey,
		cfg.MinIOBucket, cfg.MinIOUseSSL,
	)
	if err != nil {
		log.Fatalf("erro ao conectar no MinIO: %v", err)
	}

	proc := processor.New()
	notifier := notification.NewLogNotifier()

	// Consumer RabbitMQ com prefetch configurado.
	c, err := consumer.New(
		cfg.RabbitMQURL, cfg.QueueName, cfg.PrefetchCount,
		videoRepo, minioStorage, proc, notifier,
	)
	if err != nil {
		log.Fatalf("erro ao criar consumer: %v", err)
	}
	defer c.Close()

	// Inicia o loop de consumo bloqueante até SIGTERM/SIGINT.
	if err := c.Run(ctx); err != nil && err != context.Canceled {
		log.Fatalf("consumer encerrado com erro: %v", err)
	}

	log.Println("worker encerrado com sucesso")
}
