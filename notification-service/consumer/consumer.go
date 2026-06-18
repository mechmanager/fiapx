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

const emailSubject = "[FIAP X] Erro ao processar seu vídeo"

const emailBodyTpl = `<!DOCTYPE html>
<html>
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width,initial-scale=1"></head>
<body style="font-family:Inter,Arial,sans-serif;background:#f4f4f8;margin:0;padding:24px;">
  <div style="max-width:520px;margin:0 auto;background:#ffffff;border-radius:16px;overflow:hidden;box-shadow:0 4px 24px rgba(0,0,0,.08);">
    <div style="background:#0b0b14;padding:28px 32px;text-align:center;">
      <span style="color:#fff;font-size:1.5rem;font-weight:800;letter-spacing:-.5px;">FIAP<span style="color:#ed0c6e;">X</span></span>
      <span style="background:#ed0c6e;color:#fff;font-size:.65rem;font-weight:700;padding:2px 8px;border-radius:20px;letter-spacing:.5px;text-transform:uppercase;margin-left:8px;vertical-align:middle;">Pos Tech</span>
    </div>
    <div style="padding:36px 32px;">
      <h2 style="color:#111827;font-size:1.1rem;font-weight:700;margin:0 0 12px;">Falha no processamento do vídeo</h2>
      <p style="color:#6b7280;font-size:.92rem;line-height:1.6;margin:0 0 24px;">Olá! Identificamos um problema ao processar o seu vídeo. Veja os detalhes abaixo.</p>
      <div style="background:#fff5f7;border:1px solid rgba(237,12,110,.2);border-radius:10px;padding:16px 20px;margin-bottom:24px;">
        <p style="margin:0 0 4px;font-size:.75rem;color:#9ca3af;text-transform:uppercase;letter-spacing:.5px;font-weight:600;">Arquivo</p>
        <p style="margin:0;font-size:1rem;font-weight:700;color:#111827;">%s</p>
      </div>
      <p style="color:#6b7280;font-size:.9rem;line-height:1.6;margin:0;">
        Nosso sistema encontrou um problema técnico durante a extração de frames deste arquivo.
        Isso pode acontecer se o arquivo estiver corrompido ou em um formato não suportado.
      </p>
      <div style="margin:20px 0;padding:16px;background:#f9fafb;border-radius:8px;">
        <p style="margin:0 0 8px;font-size:.85rem;font-weight:700;color:#374151;">O que fazer:</p>
        <ul style="margin:0;padding-left:18px;color:#6b7280;font-size:.88rem;line-height:1.8;">
          <li>Verifique se o arquivo de vídeo não está corrompido</li>
          <li>Certifique-se de que está em um formato suportado (MP4, AVI, MOV, MKV)</li>
          <li>Tente enviar o arquivo novamente pela plataforma</li>
        </ul>
      </div>
    </div>
    <div style="background:#f9fafb;border-top:1px solid #e5e7eb;padding:20px 32px;text-align:center;">
      <p style="margin:0;font-size:.75rem;color:#9ca3af;">© FIAP X — Pos Tech &nbsp;·&nbsp; Você recebeu este e-mail porque está cadastrado em nossa plataforma.</p>
    </div>
  </div>
</body>
</html>`

// NotificationMessage representa a mensagem publicada na fila de notificação.
type NotificationMessage struct {
	UserID   uuid.UUID `json:"user_id"`
	VideoID  uuid.UUID `json:"video_id"`
	Filename string    `json:"filename"`
	Error    string    `json:"error"`
}

// NotificationConsumer consome mensagens da fila de notificação.
type NotificationConsumer struct {
	conn     *amqp.Connection
	channel  *amqp.Channel
	queue    string
	userRepo domain.UserRepository
	mailer   mailer.Mailer
}

// New cria e configura o consumidor de notificações.
func New(rabbitURL, queueName string, _ int, userRepo domain.UserRepository, m mailer.Mailer) (*NotificationConsumer, error) {
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
		conn:     conn,
		channel:  ch,
		queue:    queueName,
		userRepo: userRepo,
		mailer:   m,
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

	body := fmt.Sprintf(emailBodyTpl, notification.Filename)
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
	// requeue=false envia para DLQ; requeue=true bypassa o dead-letter exchange
	// e nunca incrementa x-death, causando loop infinito.
	log.Printf("[WARN] falha transitória na notificação → enviando para DLQ")
	if err := msg.Nack(false, false); err != nil {
		log.Printf("[WARN] nack error: %v", err)
	}
}
