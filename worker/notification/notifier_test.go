package notification_test

import (
	"testing"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/worker/notification"
)

func TestLogNotifier_NotifyError(t *testing.T) {
	n := notification.NewLogNotifier()
	// Apenas garante que não entra em pânico.
	n.NotifyError(uuid.New(), uuid.New(), "ffmpeg falhou")
}
