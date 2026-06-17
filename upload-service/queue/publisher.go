package queue

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher implements domain.QueuePublisher using RabbitMQ.
type RabbitMQPublisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

// NewRabbitMQPublisher connects to RabbitMQ, declares the durable queue with a DLQ.
func NewRabbitMQPublisher(url, queueName string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	_, err = ch.QueueDeclare(queueName+".dlq", true, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": queueName + ".dlq",
	}
	_, err = ch.QueueDeclare(queueName, true, false, false, false, args)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQPublisher{conn: conn, channel: ch, queueName: queueName}, nil
}

// Publish sends a persistent JSON message to the configured queue.
func (p *RabbitMQPublisher) Publish(_ context.Context, body []byte) error {
	return p.channel.Publish(
		"",
		p.queueName,
		false,
		false,
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

// Close gracefully closes the channel and connection.
func (p *RabbitMQPublisher) Close() {
	if err := p.channel.Close(); err != nil {
		log.Printf("[WARN] publisher channel close: %v", err)
	}
	if err := p.conn.Close(); err != nil {
		log.Printf("[WARN] publisher connection close: %v", err)
	}
}
