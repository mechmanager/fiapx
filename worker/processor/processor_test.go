package processor_test

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mechmanager/fiapx/worker/processor"
)

// makeFakeExecutor retorna um executor que cria n frames PNG falsos.
func makeFakeExecutor(n int) func(videoPath, framePattern string) error {
	return func(_, framePattern string) error {
		dir := filepath.Dir(framePattern)
		for i := 1; i <= n; i++ {
			name := filepath.Join(dir, fmt.Sprintf("frame_%04d.png", i))
			if err := os.WriteFile(name, []byte("png"), 0o644); err != nil {
				return err
			}
		}
		return nil
	}
}

func TestProcess_HappyPath(t *testing.T) {
	p := processor.NewWithExecutor(makeFakeExecutor(3))
	workDir := t.TempDir()

	result, err := p.Process(context.Background(), strings.NewReader("fake-video"), "v.mp4", workDir)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	if result.FrameCount != 3 {
		t.Errorf("esperava 3 frames, obteve %d", result.FrameCount)
	}
	if _, err := os.Stat(result.ZipPath); err != nil {
		t.Errorf("zip não encontrado: %v", err)
	}
}

func TestProcess_FFmpegError(t *testing.T) {
	p := processor.NewWithExecutor(func(_, _ string) error {
		return fmt.Errorf("ffmpeg não instalado")
	})
	workDir := t.TempDir()
	_, err := p.Process(context.Background(), strings.NewReader("data"), "v.mp4", workDir)
	if err == nil {
		t.Fatal("esperava erro do ffmpeg")
	}
}

func TestProcess_NoFrames(t *testing.T) {
	// Executor bem-sucedido mas não cria nenhum frame PNG.
	p := processor.NewWithExecutor(func(_, _ string) error { return nil })
	workDir := t.TempDir()
	_, err := p.Process(context.Background(), strings.NewReader("data"), "v.mp4", workDir)
	if err == nil {
		t.Fatal("esperava erro: nenhum frame extraído")
	}
}

func TestProcess_SaveFileError(t *testing.T) {
	// Diretório inexistente → erro ao salvar o arquivo de vídeo.
	p := processor.New()
	_, err := p.Process(context.Background(), strings.NewReader("data"), "v.mp4", "/nao/existe")
	if err == nil {
		t.Fatal("esperava erro ao salvar em diretório inexistente")
	}
}
