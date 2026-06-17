// Package processor implementa a extração de frames e a criação do zip.
// A lógica central (ffmpeg fps=1 + zip) é reaproveitada do monolito legacy/.
package processor

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// Result contém o caminho local do zip gerado e o total de frames extraídos.
type Result struct {
	ZipPath    string
	FrameCount int
}

// ffmpegExecutor é o tipo da função que executa o ffmpeg.
// Injetável para testes.
type ffmpegExecutor func(videoPath, framePattern string) error

// Processor executa o pipeline ffmpeg → zip em um diretório temporário.
type Processor struct {
	execute ffmpegExecutor
}

// New cria um Processor usando o ffmpeg do sistema.
func New() *Processor {
	return &Processor{execute: runFFmpeg}
}

// NewWithExecutor cria um Processor com executor personalizado (para testes).
func NewWithExecutor(exec ffmpegExecutor) *Processor {
	return &Processor{execute: exec}
}

// Process baixa o vídeo para disco, extrai frames com ffmpeg e compacta em zip.
func (p *Processor) Process(_ context.Context, videoReader io.Reader, videoFilename, workDir string) (*Result, error) {
	videoPath := filepath.Join(workDir, videoFilename)
	if err := saveToFile(videoReader, videoPath); err != nil {
		return nil, fmt.Errorf("erro ao salvar vídeo temporário: %w", err)
	}

	framesDir := filepath.Join(workDir, "frames")
	if err := os.MkdirAll(framesDir, 0o755); err != nil {
		return nil, fmt.Errorf("erro ao criar diretório de frames: %w", err)
	}

	framePattern := filepath.Join(framesDir, "frame_%04d.png")
	if err := p.execute(videoPath, framePattern); err != nil {
		return nil, fmt.Errorf("ffmpeg falhou: %w", err)
	}

	frames, err := filepath.Glob(filepath.Join(framesDir, "*.png"))
	if err != nil || len(frames) == 0 {
		return nil, fmt.Errorf("nenhum frame extraído do vídeo")
	}

	zipPath := filepath.Join(workDir, "frames.zip")
	if err := createZip(frames, zipPath); err != nil {
		return nil, fmt.Errorf("erro ao criar zip: %w", err)
	}

	return &Result{ZipPath: zipPath, FrameCount: len(frames)}, nil
}

// runFFmpeg executa o ffmpeg extraindo 1 frame por segundo.
func runFFmpeg(videoPath, framePattern string) error {
	cmd := exec.Command("ffmpeg",
		"-i", videoPath,
		"-vf", "fps=1",
		"-y",
		framePattern,
	)
	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("%s — %w", string(output), err)
	}
	return nil
}

func saveToFile(reader io.Reader, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = io.Copy(f, reader)
	return err
}

func createZip(files []string, zipPath string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	w := zip.NewWriter(zipFile)
	defer w.Close()

	for _, f := range files {
		if err := addToZip(w, f); err != nil {
			return err
		}
	}
	return nil
}

func addToZip(w *zip.Writer, filePath string) error {
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}
	header.Name = filepath.Base(filePath)
	header.Method = zip.Deflate

	writer, err := w.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(writer, f)
	return err
}
