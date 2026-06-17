// Testes internos (mesmo pacote) para cobrir as funções privadas de zip.
package processor

import (
	"io"
	"os"
	"path/filepath"
	"testing"
)

func TestCreateZip_HappyPath(t *testing.T) {
	dir := t.TempDir()

	// Cria dois arquivos de frame falsos.
	f1 := filepath.Join(dir, "frame_0001.png")
	f2 := filepath.Join(dir, "frame_0002.png")
	os.WriteFile(f1, []byte("png-data-1"), 0o644)
	os.WriteFile(f2, []byte("png-data-2"), 0o644)

	zipPath := filepath.Join(dir, "out.zip")
	err := createZip([]string{f1, f2}, zipPath)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}

	info, err := os.Stat(zipPath)
	if err != nil {
		t.Fatalf("zip não foi criado: %v", err)
	}
	if info.Size() == 0 {
		t.Error("zip está vazio")
	}
}

func TestCreateZip_FileNotFound(t *testing.T) {
	dir := t.TempDir()
	err := createZip([]string{"/nao/existe.png"}, filepath.Join(dir, "out.zip"))
	if err == nil {
		t.Fatal("esperava erro para arquivo inexistente")
	}
}

func TestCreateZip_InvalidZipPath(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "frame.png")
	os.WriteFile(f, []byte("data"), 0o644)

	err := createZip([]string{f}, "/dir/nao/existe/out.zip")
	if err == nil {
		t.Fatal("esperava erro para caminho de zip inválido")
	}
}

func TestAddToZip_FileNotFound(t *testing.T) {
	// Testa via createZip com arquivo inexistente (percurso que chama addToZip internamente).
	dir := t.TempDir()
	err := createZip([]string{"/nao/existe/frame.png"}, filepath.Join(dir, "out.zip"))
	if err == nil {
		t.Fatal("esperava erro")
	}
}

func TestSaveToFile_HappyPath(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "video.mp4")
	import_reader := &fakeReader{data: []byte("conteudo")}
	err := saveToFile(import_reader, path)
	if err != nil {
		t.Fatalf("esperava nil, obteve %v", err)
	}
	data, _ := os.ReadFile(path)
	if string(data) != "conteudo" {
		t.Errorf("conteúdo incorreto: %q", string(data))
	}
}

func TestSaveToFile_InvalidPath(t *testing.T) {
	err := saveToFile(&fakeReader{}, "/nao/existe/video.mp4")
	if err == nil {
		t.Fatal("esperava erro para caminho inválido")
	}
}

// --- helpers ---

type fakeReader struct {
	data []byte
	pos  int
}

func (r *fakeReader) Read(p []byte) (int, error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n := copy(p, r.data[r.pos:])
	r.pos += n
	return n, nil
}
