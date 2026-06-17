// Package consumer implementa o consumidor de mensagens de notificação do RabbitMQ.
package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/notification-service/domain"
	"github.com/mechmanager/fiapx/notification-service/mailer"
)

const (
	emailSubject = "[FIAP X] Erro ao processar seu vídeo"
	emailBodyFmt = "Olá! Houve um erro ao processar o vídeo \"%s\".\n\nDetalhes: %s"
)

// NotificationMessage representa a mensagem publicada na fila de notificação.
type NotificationMessage struct {
	UserID   uuid.UUID `json:"user_id"`
	VideoID  uuid.UUID `json:"video_id"`
	Filename string    `json:"filename"`
	Error    string    `json:"error"`
}

// NotificationConsumer consome mensagens da fila de notificação.
type NotificationConsumer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	queue      string
	maxRetries int
	userRepo   domain.UserRepository
	mailer     mailer.Mailer
}

// New cria e configura o consumidor de notificações.
func New(rabbitURL, queueName string, maxRetries int, userRepo domain.UserRepository, m mailer.Mailer) (*NotificationConsumer, error) {
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao abrir canal RabbitMQ: %w", err)
	}

	// Declara a DLQ primeiro.
	_, err = ch.QueueDeclare(
		queueName+".dlq",
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar DLQ: %w", err)
	}

	// Declara a fila principal com DLQ configurada.
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": queueName + ".dlq",
		},
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar fila: %w", err)
	}

	// Prefetch de 5 mensagens.
	if err := ch.Qos(5, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao configurar QoS: %w", err)
	}

	return &NotificationConsumer{
		conn:       conn,
		channel:    ch,
		queue:      queueName,
		maxRetries: maxRetries,
		userRepo:   userRepo,
		mailer:     m,
	}, nil
}

// Close libera os recursos do consumidor.
func (c *NotificationConsumer) Close() {
	if c.channel != nil {
		c.channel.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

// Run inicia o loop de consumo de mensagens.
func (c *NotificationConsumer) Run(ctx context.Context) error {
	msgs, err := c.channel.Consume(
		c.queue,
		"",    // consumer tag
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,
	)
	if err != nil {
		return fmt.Errorf("erro ao registrar consumidor: %w", err)
	}

	log.Printf("notification-service aguardando mensagens na fila %q", c.queue)

	for {
		select {
		case <-ctx.Done():
			log.Println("notification-service encerrando consumo")
			return nil
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal de mensagens fechado")
			}
			c.handle(ctx, msg)
		}
	}
}

// handle processa uma mensagem individual.
func (c *NotificationConsumer) handle(ctx context.Context, msg amqp.Delivery) {
	var notification NotificationMessage
	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		log.Printf("erro ao deserializar mensagem: %v — enviando para DLQ", err)
		msg.Nack(false, false)
		return
	}

	email, err := c.userRepo.FindEmailByUserID(ctx, notification.UserID)
	if err != nil {
		log.Printf("erro ao buscar e-mail do usuário %s: %v", notification.UserID, err)
		c.nackWithRetry(msg)
		return
	}

	if email == "" {
		log.Printf("usuário %s não encontrado, descartando notificação para vídeo %s",
			notification.UserID, notification.VideoID)
		msg.Ack(false)
		return
	}

	body := fmt.Sprintf(emailBodyFmt, notification.Filename, notification.Error)
	if err := c.mailer.Send(email, emailSubject, body); err != nil {
		log.Printf("erro ao enviar e-mail para %s: %v", email, err)
		c.nackWithRetry(msg)
		return
	}

	log.Printf("notificação enviada para %s (vídeo %s)", email, notification.VideoID)
	msg.Ack(false)
}

// nackWithRetry faz nack verificando o contador de retentativas.
func (c *NotificationConsumer) nackWithRetry(msg amqp.Delivery) {
	retries := int64(0)
	if xDeath, ok := msg.Headers["x-death"]; ok {
		if deathList, ok := xDeath.([]interface{}); ok && len(deathList) > 0 {
			if deathMap, ok := deathList[0].(amqp.Table); ok {
				if count, ok := deathMap["count"].(int64); ok {
					retries = count
				}
			}
		}
	}

	if int(retries) >= c.maxRetries {
		log.Printf("mensagem excedeu %d tentativas, enviando para DLQ", c.maxRetries)
		msg.Nack(false, false)
		return
	}

	msg.Nack(false, true) // requeue
}
