// Package notification implementa as estratégias de notificação do worker.
package notification

import (
	"log"

	"github.com/google/uuid"
)

// LogNotifier implementa domain.Notifier registrando o erro no log estruturado.
// A interface permite trocar por uma implementação SMTP ou webhook sem alterar o consumer.
type LogNotifier struct{}

// NewLogNotifier cria o notificador baseado em log.
func NewLogNotifier() *LogNotifier {
	return &LogNotifier{}
}

// NotifyError registra a falha de processamento com os identificadores envolvidos.
func (n *LogNotifier) NotifyError(videoID uuid.UUID, userID uuid.UUID, reason string) {
	log.Printf("[ERRO-PROCESSAMENTO] video_id=%s user_id=%s motivo=%q", videoID, userID, reason)
}
