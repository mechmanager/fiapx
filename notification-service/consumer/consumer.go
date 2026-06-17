package consumer

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

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
		_ = conn.Close()
		return nil, fmt.Errorf("erro ao abrir canal RabbitMQ: %w", err)
	}

	_, err = ch.QueueDeclare(queueName+".dlq", true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("erro ao declarar DLQ: %w", err)
	}

	_, err = ch.QueueDeclare(
		queueName,
		true, false, false, false,
		amqp.Table{
			"x-dead-letter-exchange":    "",
			"x-dead-letter-routing-key": queueName + ".dlq",
		},
	)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("erro ao declarar fila: %w", err)
	}

	if err := ch.Qos(5, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
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
		if err := c.channel.Close(); err != nil {
			log.Printf("[WARN] channel close: %v", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			log.Printf("[WARN] connection close: %v", err)
		}
	}
}

// Run inicia o loop de consumo de mensagens.
func (c *NotificationConsumer) Run(ctx context.Context) error {
	msgs, err := c.channel.Consume(c.queue, "", false, false, false, false, nil)
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

func (c *NotificationConsumer) handle(ctx context.Context, msg amqp.Delivery) {
	var notification NotificationMessage
	if err := json.Unmarshal(msg.Body, &notification); err != nil {
		log.Printf("erro ao deserializar mensagem: %v — enviando para DLQ", err)
		if err := msg.Nack(false, false); err != nil {
			log.Printf("[WARN] nack error: %v", err)
		}
		return
	}

	email, err := c.userRepo.FindEmailByUserID(ctx, notification.UserID)
	if err != nil {
		log.Printf("erro ao buscar e-mail do usuário %s: %v", notification.UserID, err)
		c.nackWithRetry(msg)
		return
	}

	if email == "" {
		log.Printf("usuário %s não encontrado, descartando notificação", notification.UserID)
		if err := msg.Ack(false); err != nil {
			log.Printf("[WARN] ack error: %v", err)
		}
		return
	}

	body := fmt.Sprintf(emailBodyFmt, notification.Filename, notification.Error)
	if err := c.mailer.Send(email, emailSubject, body); err != nil {
		log.Printf("erro ao enviar e-mail: %v", err)
		c.nackWithRetry(msg)
		return
	}

	log.Printf("notificação enviada para vídeo %s", notification.VideoID)
	if err := msg.Ack(false); err != nil {
		log.Printf("[WARN] ack error: %v", err)
	}
}

func (c *NotificationConsumer) nackWithRetry(msg amqp.Delivery) {
	retries := xDeathCount(msg)
	if retries >= c.maxRetries {
		log.Printf("mensagem excedeu %d tentativas, enviando para DLQ", c.maxRetries)
		if err := msg.Nack(false, false); err != nil {
			log.Printf("[WARN] nack error: %v", err)
		}
		return
	}
	if err := msg.Nack(false, true); err != nil {
		log.Printf("[WARN] nack error: %v", err)
	}
}

func xDeathCount(msg amqp.Delivery) int {
	xDeath, ok := msg.Headers["x-death"]
	if !ok {
		return 0
	}
	deathList, ok := xDeath.([]interface{})
	if !ok || len(deathList) == 0 {
		return 0
	}
	deathMap, ok := deathList[0].(amqp.Table)
	if !ok {
		return 0
	}
	count, _ := deathMap["count"].(int64)
	return int(count)
}
