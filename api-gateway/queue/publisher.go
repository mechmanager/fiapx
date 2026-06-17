// Package queue implementa a publicação de mensagens no RabbitMQ.
package queue

import (
	"context"
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitMQPublisher implementa domain.QueuePublisher usando RabbitMQ.
type RabbitMQPublisher struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
}

// NewRabbitMQPublisher conecta ao RabbitMQ e declara a fila durável.
func NewRabbitMQPublisher(url, queueName string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao abrir canal RabbitMQ: %w", err)
	}

	// Declara a fila como durável para sobreviver a reinicializações do broker.
	_, err = ch.QueueDeclare(
		queueName,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar fila: %w", err)
	}

	return &RabbitMQPublisher{conn: conn, channel: ch, queueName: queueName}, nil
}

// Publish envia uma mensagem persistente para a fila de processamento.
func (p *RabbitMQPublisher) Publish(_ context.Context, body []byte) error {
	return p.channel.Publish(
		"",          // exchange padrão
		p.queueName, // routing key = nome da fila
		false,       // mandatory
		false,       // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent, // mensagem sobrevive a reinicializações
			ContentType:  "application/json",
			Body:         body,
		},
	)
}

// Close encerra a conexão com o RabbitMQ.
func (p *RabbitMQPublisher) Close() {
	p.channel.Close()
	p.conn.Close()
}
