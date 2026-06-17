package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/google/uuid"
	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mechmanager/fiapx/worker/domain"
)

// RabbitMQNotifier publica mensagens de erro na fila de notificação.
type RabbitMQNotifier struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

func NewRabbitMQNotifier(url, queueName string) (*RabbitMQNotifier, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("notifier: erro ao conectar no RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("notifier: erro ao abrir canal: %w", err)
	}
	dlq := queueName + ".dlq"
	_, err = ch.QueueDeclare(dlq, true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("notifier: erro ao declarar DLQ: %w", err)
	}
	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": dlq,
	}
	_, err = ch.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("notifier: erro ao declarar fila: %w", err)
	}
	return &RabbitMQNotifier{conn: conn, channel: ch, queueName: queueName}, nil
}

func (n *RabbitMQNotifier) NotifyError(_ context.Context, videoID, userID uuid.UUID, filename, reason string) {
	msg := domain.NotificationMessage{
		UserID:   userID,
		VideoID:  videoID,
		Filename: filename,
		Error:    reason,
	}
	body, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[WARN] notifier: falha ao serializar mensagem: %v", err)
		return
	}
	err = n.channel.Publish("", n.queueName, false, false, amqp.Publishing{
		DeliveryMode: amqp.Persistent,
		ContentType:  "application/json",
		Body:         body,
	})
	if err != nil {
		log.Printf("[WARN] notifier: falha ao publicar: %v", err)
	}
}

func (n *RabbitMQNotifier) Close() {
	if err := n.channel.Close(); err != nil {
		log.Printf("[WARN] notifier channel close: %v", err)
	}
	if err := n.conn.Close(); err != nil {
		log.Printf("[WARN] notifier connection close: %v", err)
	}
}
