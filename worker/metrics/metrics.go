package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	VideosProcessed = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "worker_videos_processed_total",
		Help: "Total de vídeos processados pelo worker.",
	}, []string{"status"})

	ProcessingDuration = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "worker_processing_duration_seconds",
		Help:    "Duração do processamento de cada vídeo em segundos.",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
	})
)
