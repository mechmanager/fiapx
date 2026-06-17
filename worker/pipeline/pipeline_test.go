package pipeline_test

import (
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/mechmanager/fiapx/worker/mocks"
	"github.com/mechmanager/fiapx/worker/pipeline"
	"github.com/mechmanager/fiapx/worker/processor"
)

// mockProcessor implementa pipeline.VideoProcessor para testes sem ffmpeg real.
type mockProcessor struct {
	result *processor.Result
	err    error
}

func (m *mockProcessor) Process(_ context.Context, _ io.Reader, _ string, workDir string) (*processor.Result, error) {
	if m.err != nil {
		return nil, m.err
	}
	zipPath := workDir + "/frames.zip"
	os.WriteFile(zipPath, []byte("fake-zip"), 0o644)
	result := m.result
	if result == nil {
		result = &processor.Result{FrameCount: 3}
	}
	result.ZipPath = zipPath
	return result, nil
}

func newPipeline(
	videos *mocks.VideoRepository,
	stor *mocks.ObjectStorage,
	proc pipeline.VideoProcessor,
	notifier *mocks.Notifier,
) *pipeline.Pipeline {
	return pipeline.New(videos, stor, proc, notifier)
}

func msgBody(t *testing.T, videoID uuid.UUID, s3Key string) []byte {
	t.Helper()
	body := `{"video_id":"` + videoID.String() + `","user_id":"` + uuid.New().String() + `","s3_key":"` + s3Key + `"}`
	return []byte(body)
}

// --- JSON inválido ---

func TestPipeline_InvalidJSON_Nack(t *testing.T) {
	p := newPipeline(
		&mocks.VideoRepository{},
		&mocks.ObjectStorage{},
		&mockProcessor{},
		&mocks.Notifier{NotifyErrorFn: func(_, _ uuid.UUID, _ string) {}},
	)
	ack, err := p.HandleMessage(context.Background(), []byte("nao-e-json"))
	if ack {
		t.Error("esperava ack=false para JSON inválido")
	}
	if err != nil {
		t.Errorf("esperava err=nil para JSON inválido, obteve %v", err)
	}
}

// --- Erro ao marcar PROCESSING ---

func TestPipeline_UpdateProcessingError(t *testing.T) {
	videoID := uuid.New()
	body := msgBody(t, videoID, "videos/v.mp4")

	videos := &mocks.VideoRepository{
		UpdateStatusFn: func(_ context.Context, _ uuid.UUID, _, _ string) error {
			return errors.New("banco indisponível")
		},
		UpdateDoneFn: func(_ context.Context, _ uuid.UUID, _ string, _ int) error { return nil },
	}
	notified := false
	notifier := &mocks.Notifier{NotifyErrorFn: func(_, _ uuid.UUID, _ string) { notified = true }}
	p := newPipeline(videos, &mocks.ObjectStorage{}, &mockProcessor{}, notifier)
	ack, err := p.HandleMessage(context.Background(), body)

	if ack {
		t.Error("esperava ack=false")
	}
	if err == nil {
		t.Error("esperava erro")
	}
	if !notified {
		t.Error("esperava NotifyError chamado")
	}
}

// --- Erro ao baixar vídeo ---

func TestPipeline_DownloadError(t *testing.T) {
	videoID := uuid.New()
	body := msgBody(t, videoID, "videos/v.mp4")

	videos := &mocks.VideoRepository{
		UpdateStatusFn: func(_ context.Context, _ uuid.UUID, _, _ string) error { return nil },
		UpdateDoneFn:   func(_ context.Context, _ uuid.UUID, _ string, _ int) error { return nil },
	}
	stor := &mocks.ObjectStorage{
		DownloadFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return nil, errors.New("minio indisponível")
		},
	}
	notified := false
	notifier := &mocks.Notifier{NotifyErrorFn: func(_, _ uuid.UUID, _ string) { notified = true }}
	p := newPipeline(videos, stor, &mockProcessor{}, notifier)
	ack, err := p.HandleMessage(context.Background(), body)

	if ack || err == nil || !notified {
		t.Errorf("esperava ack=false, err!=nil, notified=true; obteve ack=%v err=%v notified=%v", ack, err, notified)
	}
}

// --- Erro no processor ---

func TestPipeline_ProcessorError(t *testing.T) {
	videoID := uuid.New()
	body := msgBody(t, videoID, "videos/fake.mp4")

	videos := &mocks.VideoRepository{
		UpdateStatusFn: func(_ context.Context, _ uuid.UUID, _, _ string) error { return nil },
		UpdateDoneFn:   func(_ context.Context, _ uuid.UUID, _ string, _ int) error { return nil },
	}
	stor := &mocks.ObjectStorage{
		DownloadFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("conteudo")), nil
		},
	}
	notified := false
	notifier := &mocks.Notifier{NotifyErrorFn: func(_, _ uuid.UUID, _ string) { notified = true }}
	proc := &mockProcessor{err: errors.New("ffmpeg falhou")}
	p := newPipeline(videos, stor, proc, notifier)
	ack, err := p.HandleMessage(context.Background(), body)

	if ack || err == nil || !notified {
		t.Errorf("esperava ack=false, err!=nil, notified=true; obteve ack=%v err=%v notified=%v", ack, err, notified)
	}
}

// --- Erro ao fazer upload do zip ---

func TestPipeline_UploadZipError(t *testing.T) {
	videoID := uuid.New()
	body := msgBody(t, videoID, "videos/v.mp4")

	videos := &mocks.VideoRepository{
		UpdateStatusFn: func(_ context.Context, _ uuid.UUID, _, _ string) error { return nil },
		UpdateDoneFn:   func(_ context.Context, _ uuid.UUID, _ string, _ int) error { return nil },
	}
	stor := &mocks.ObjectStorage{
		DownloadFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("conteudo")), nil
		},
		UploadFileFn: func(_ context.Context, _, _ string) error {
			return errors.New("minio cheio")
		},
	}
	notifier := &mocks.Notifier{NotifyErrorFn: func(_, _ uuid.UUID, _ string) {}}
	proc := &mockProcessor{result: &processor.Result{FrameCount: 2}}
	p := newPipeline(videos, stor, proc, notifier)
	ack, err := p.HandleMessage(context.Background(), body)

	if ack || err == nil {
		t.Errorf("esperava ack=false, err!=nil; obteve ack=%v err=%v", ack, err)
	}
}

// --- Erro no UpdateDone ---

func TestPipeline_UpdateDoneError(t *testing.T) {
	videoID := uuid.New()
	body := msgBody(t, videoID, "videos/v.mp4")

	notified := false
	videos := &mocks.VideoRepository{
		UpdateStatusFn: func(_ context.Context, _ uuid.UUID, _, _ string) error { return nil },
		UpdateDoneFn: func(_ context.Context, _ uuid.UUID, _ string, _ int) error {
			return errors.New("constraint")
		},
	}
	stor := &mocks.ObjectStorage{
		DownloadFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("conteudo")), nil
		},
		UploadFileFn: func(_ context.Context, _, _ string) error { return nil },
	}
	notifier := &mocks.Notifier{NotifyErrorFn: func(_, _ uuid.UUID, _ string) { notified = true }}
	proc := &mockProcessor{result: &processor.Result{FrameCount: 2}}
	p := newPipeline(videos, stor, proc, notifier)
	ack, err := p.HandleMessage(context.Background(), body)

	if ack || err == nil || !notified {
		t.Errorf("esperava ack=false, err!=nil, notified=true; obteve ack=%v err=%v notified=%v", ack, err, notified)
	}
}

// --- Sucesso completo ---

func TestPipeline_Success(t *testing.T) {
	videoID := uuid.New()
	body := msgBody(t, videoID, "videos/v.mp4")

	var doneCalled bool
	videos := &mocks.VideoRepository{
		UpdateStatusFn: func(_ context.Context, _ uuid.UUID, _, _ string) error { return nil },
		UpdateDoneFn: func(_ context.Context, _ uuid.UUID, _ string, _ int) error {
			doneCalled = true
			return nil
		},
	}
	stor := &mocks.ObjectStorage{
		DownloadFn: func(_ context.Context, _ string) (io.ReadCloser, error) {
			return io.NopCloser(strings.NewReader("conteudo")), nil
		},
		UploadFileFn: func(_ context.Context, _, _ string) error { return nil },
	}
	notifier := &mocks.Notifier{NotifyErrorFn: func(_, _ uuid.UUID, _ string) {}}
	proc := &mockProcessor{result: &processor.Result{FrameCount: 5}}
	p := newPipeline(videos, stor, proc, notifier)
	ack, err := p.HandleMessage(context.Background(), body)

	if !ack || err != nil || !doneCalled {
		t.Errorf("esperava ack=true, err=nil, doneCalled=true; obteve ack=%v err=%v doneCalled=%v", ack, err, doneCalled)
	}
}
