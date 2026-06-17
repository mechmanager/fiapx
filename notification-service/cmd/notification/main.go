package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"

	"github.com/mechmanager/fiapx/notification-service/config"
	"github.com/mechmanager/fiapx/notification-service/consumer"
	"github.com/mechmanager/fiapx/notification-service/mailer"
	"github.com/mechmanager/fiapx/notification-service/repository"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Pool de conexões com o PostgreSQL.
	pool, err := repository.NewPool(ctx, cfg.PostgresDSN)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}
	defer pool.Close()

	// Repositório de usuários.
	userRepo := repository.NewUserRepo(pool)

	// Mailer.
	m := mailer.NewMailer(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPFrom)

	// Consumidor de notificações.
	c, err := consumer.New(cfg.RabbitMQURL, cfg.NotificationQueue, cfg.MaxRetries, userRepo, m)
	if err != nil {
		log.Fatalf("erro ao criar consumidor: %v", err)
	}
	defer c.Close()

	if err := c.Run(ctx); err != nil {
		log.Fatalf("erro no consumidor: %v", err)
	}
}
