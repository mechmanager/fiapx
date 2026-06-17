package notification_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/worker/mocks"
)

func TestNotifier_NotifyError_NoOp(t *testing.T) {
	// Garante que o mock do Notifier não entra em pânico quando NotifyErrorFn é nil.
	n := &mocks.Notifier{}
	n.NotifyError(context.Background(), uuid.New(), uuid.New(), "video.mp4", "ffmpeg falhou")
}
