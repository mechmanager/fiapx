// Package queue implements RabbitMQ message publishing for the upload-service.
package queue

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher implements domain.QueuePublisher using RabbitMQ.
type RabbitMQPublisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

// NewRabbitMQPublisher connects to RabbitMQ, declares the durable queue with a DLQ,
// and also declares the DLQ itself.
func NewRabbitMQPublisher(url, queueName string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open RabbitMQ channel: %w", err)
	}

	// Declare the dead-letter queue first.
	_, err = ch.QueueDeclare(
		queueName+".dlq",
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare DLQ: %w", err)
	}

	// Declare the main queue with dead-letter routing arguments.
	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": queueName + ".dlq",
	}
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		args,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	return &RabbitMQPublisher{conn: conn, channel: ch, queueName: queueName}, nil
}

// Publish sends a persistent JSON message to the configured queue.
func (p *RabbitMQPublisher) Publish(_ context.Context, body []byte) error {
	return p.channel.Publish(
		"",          // default exchange
		p.queueName, // routing key
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

// Close gracefully closes the channel and connection.
func (p *RabbitMQPublisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
