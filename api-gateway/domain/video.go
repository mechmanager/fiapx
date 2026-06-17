package domain

import (
	"context"
	"io"
	"time"

	"github.com/google/uuid"
)

// Status possíveis de um vídeo ao longo do processamento.
const (
	StatusPending    = "PENDING"    // aguardando processamento na fila
	StatusProcessing = "PROCESSING" // sendo processado pelo worker
	StatusDone       = "DONE"       // processado com sucesso
	StatusError      = "ERROR"      // falhou no processamento
)

// Video representa um vídeo enviado e o resultado do seu processamento.
type Video struct {
	ID               uuid.UUID // identificador único
	UserID           uuid.UUID // dono do vídeo
	OriginalFilename string    // nome original do arquivo enviado
	S3Key            string    // chave do vídeo no object storage
	Status           string    // PENDING, PROCESSING, DONE ou ERROR
	ErrorMessage     string    // mensagem de erro quando Status = ERROR
	ZipS3Key         string    // chave do zip de frames no object storage
	FrameCount       int       // quantidade de frames extraídos
	CreatedAt        time.Time // data de criação
	UpdatedAt        time.Time // data da última atualização
}

// VideoRepository abstrai o acesso aos dados de vídeos.
type VideoRepository interface {
	// Create persiste um novo vídeo no banco.
	Create(ctx context.Context, video *Video) error
	// ListByUser lista os vídeos de um usuário ordenados por data de criação.
	ListByUser(ctx context.Context, userID uuid.UUID) ([]*Video, error)
	// FindByID busca um vídeo pelo seu identificador.
	FindByID(ctx context.Context, id uuid.UUID) (*Video, error)
}

// ObjectStorage abstrai o armazenamento de objetos (MinIO/S3).
type ObjectStorage interface {
	// Upload envia um objeto para o storage a partir de um leitor.
	Upload(ctx context.Context, objectKey string, reader io.Reader, size int64, contentType string) error
	// Download devolve um leitor para o conteúdo de um objeto.
	Download(ctx context.Context, objectKey string) (io.ReadCloser, error)
}

// QueuePublisher abstrai a publicação de mensagens na fila.
type QueuePublisher interface {
	// Publish envia uma mensagem (payload JSON) para a fila de processamento.
	Publish(ctx context.Context, body []byte) error
}
