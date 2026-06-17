// Package consumer_test valida que o adaptador AMQP compila e que
// a assinatura de New corresponde às interfaces esperadas.
// Testes funcionais do processamento estão em worker/pipeline.
package consumer_test

import (
	"testing"

	"github.com/mechmanager/fiapx/worker/consumer"
	"github.com/mechmanager/fiapx/worker/pipeline"
	"github.com/mechmanager/fiapx/worker/processor"
)

// TestConsumer_InterfaceCompliance garante que *processor.Processor satisfaz
// pipeline.VideoProcessor em tempo de compilação (o consumer repassa isso).
func TestConsumer_InterfaceCompliance(t *testing.T) {
	var _ pipeline.VideoProcessor = processor.New()
	// Se compilar, a interface é satisfeita.
	_ = consumer.New // garante que New existe e é exportado
}
