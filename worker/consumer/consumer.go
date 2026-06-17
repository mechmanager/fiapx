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
	conn       *amqp.Connection
	channel    *amqp.Channel
	queueName  string
	maxRetries int
	pipeline   *pipeline.Pipeline
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

	dlq := queueName + ".dlq"
	if _, err = ch.QueueDeclare(dlq, true, false, false, false, nil); err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("erro ao declarar DLQ: %w", err)
	}

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
	ack, _ := c.pipeline.HandleMessage(ctx, msg.Body)

	if ack {
		msg.Ack(false)
		return
	}

	retries := xDeathCount(msg)
	if retries >= c.maxRetries {
		log.Printf("[WARN] mensagem excedeu %d tentativas → DLQ", c.maxRetries)
		msg.Nack(false, false)
	} else {
		msg.Nack(false, true)
	}
}

// xDeathCount lê o número de mortes (tentativas) do header x-death do RabbitMQ.
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

func (c *Consumer) Close() {
	c.channel.Close()
	c.conn.Close()
}

var _ pipeline.VideoProcessor = (*processor.Processor)(nil)
