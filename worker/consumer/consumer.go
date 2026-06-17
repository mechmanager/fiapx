// Package consumer implementa o adaptador RabbitMQ que integra o Pipeline ao broker.
// A lógica de processamento fica em worker/pipeline, isolada e totalmente testável.
package consumer

import (
	"context"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mechmanager/fiapx/worker/domain"
	"github.com/mechmanager/fiapx/worker/pipeline"
	"github.com/mechmanager/fiapx/worker/processor"
)

// Consumer consome mensagens da fila RabbitMQ e delega ao Pipeline.
type Consumer struct {
	conn      *amqp.Connection
	channel   *amqp.Channel
	queueName string
	pipeline  *pipeline.Pipeline
}

// New conecta ao RabbitMQ, configura o prefetch e devolve o consumer pronto.
func New(
	url, queueName string,
	prefetchCount int,
	videos domain.VideoRepository,
	stor domain.ObjectStorage,
	proc pipeline.VideoProcessor,
	notifier domain.Notifier,
) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar no RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("erro ao abrir canal: %w", err)
	}

	if err := ch.Qos(prefetchCount, 0, false); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao configurar QoS: %w", err)
	}

	_, err = ch.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar fila: %w", err)
	}

	return &Consumer{
		conn: conn, channel: ch, queueName: queueName,
		pipeline: pipeline.New(videos, stor, proc, notifier),
	}, nil
}

// Run inicia o loop de consumo bloqueante. Retorna apenas quando ctx é cancelado.
func (c *Consumer) Run(ctx context.Context) error {
	msgs, err := c.channel.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar consumo: %w", err)
	}

	log.Printf("worker aguardando mensagens na fila %q", c.queueName)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal RabbitMQ fechado inesperadamente")
			}
			ack, _ := c.pipeline.HandleMessage(ctx, msg.Body)
			if ack {
				msg.Ack(false)
			} else {
				msg.Nack(false, false)
			}
		}
	}
}

// Close encerra a conexão com o RabbitMQ.
func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}

// Garante que *processor.Processor satisfaz VideoProcessor (checagem em compilação).
var _ pipeline.VideoProcessor = (*processor.Processor)(nil)
