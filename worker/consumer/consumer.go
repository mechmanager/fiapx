package consumer

import (
	"context"
	"crypto/md5"
	"fmt"
	"log"
	"sync"

	amqp "github.com/rabbitmq/amqp091-go"

	"github.com/mechmanager/fiapx/worker/domain"
	"github.com/mechmanager/fiapx/worker/pipeline"
	"github.com/mechmanager/fiapx/worker/processor"
)

// Consumer consome mensagens da fila RabbitMQ e delega ao Pipeline.
type Consumer struct {
	conn       *amqp.Connection
	channel    *amqp.Channel
	queueName  string
	maxRetries int
	pipeline   *pipeline.Pipeline

	mu      sync.Mutex
	retries map[[16]byte]int // hash do body → contagem de tentativas
}

func New(
	url, queueName string,
	prefetchCount, maxRetries int,
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

	// Declara a DLQ antes da fila principal.
	dlq := queueName + ".dlq"
	if _, err = ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar DLQ: %w", err)
	}

	// Fila principal configurada com dead-letter para a DLQ.
	args := amqp.Table{
		"x-dead-letter-exchange":    "",
		"x-dead-letter-routing-key": dlq,
	}
	if _, err = ch.QueueDeclare(queueName, true, false, false, false, args); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar fila: %w", err)
	}

	return &Consumer{
		conn: conn, channel: ch, queueName: queueName,
		maxRetries: maxRetries,
		retries:    make(map[[16]byte]int),
		pipeline:   pipeline.New(videos, stor, proc, notifier),
	}, nil
}

func (c *Consumer) Run(ctx context.Context) error {
	msgs, err := c.channel.Consume(c.queueName, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("erro ao iniciar consumo: %w", err)
	}

	log.Printf("worker aguardando mensagens na fila %q (maxRetries=%d)", c.queueName, c.maxRetries)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal RabbitMQ fechado inesperadamente")
			}
			c.handle(ctx, msg)
		}
	}
}

func (c *Consumer) handle(ctx context.Context, msg amqp.Delivery) {
	key := md5.Sum(msg.Body)
	ack, _ := c.pipeline.HandleMessage(ctx, msg.Body)

	if ack {
		c.mu.Lock()
		delete(c.retries, key)
		c.mu.Unlock()
		msg.Ack(false)
		return
	}

	c.mu.Lock()
	c.retries[key]++
	count := c.retries[key]
	c.mu.Unlock()

	if count >= c.maxRetries {
		log.Printf("[WARN] mensagem excedeu %d tentativas → DLQ", c.maxRetries)
		c.mu.Lock()
		delete(c.retries, key)
		c.mu.Unlock()
		msg.Nack(false, false) // sem requeue → vai para DLQ
	} else {
		msg.Nack(false, true) // requeue para nova tentativa
	}
}

func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}

// Garante que *processor.Processor satisfaz VideoProcessor em compilação.
var _ pipeline.VideoProcessor = (*processor.Processor)(nil)
