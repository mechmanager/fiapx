package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	httpRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gateway_http_requests_total",
			Help: "Total number of HTTP requests handled by the gateway.",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "gateway_http_request_duration_seconds",
			Help:    "Duration of HTTP requests handled by the gateway.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)
)

// uuidPattern matches UUID v4 segments in a URL path.
var uuidPattern = regexp.MustCompile(`[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}`)

// normalizePath strips query parameters and replaces UUID segments with {id}
// to avoid high-cardinality label values.
func normalizePath(rawPath string) string {
	// Strip query string
	if idx := strings.IndexByte(rawPath, '?'); idx != -1 {
		rawPath = rawPath[:idx]
	}
	// Replace UUID-like segments with {id}
	return uuidPattern.ReplaceAllString(rawPath, "{id}")
}

// responseWriter wraps http.ResponseWriter to capture the status code.
type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// Metrics holds Prometheus middleware state. Metrics are registered once via
// promauto at package init time, so this struct is lightweight.
type Metrics struct{}

// NewMetrics creates a new Metrics middleware instance.
func NewMetrics() *Metrics {
	return &Metrics{}
}

// Middleware wraps next, recording request count and duration for every request
// except GET /metrics (to avoid recursive instrumentation).
func (m *Metrics) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Skip instrumentation for the /metrics endpoint itself.
		if r.URL.Path == "/metrics" {
			next.ServeHTTP(w, r)
			return
		}

		path := normalizePath(r.URL.Path)
		method := r.Method

		rw := newResponseWriter(w)
		start := time.Now()

		next.ServeHTTP(rw, r)

		duration := time.Since(start).Seconds()
		status := strconv.Itoa(rw.statusCode)

		httpRequestsTotal.WithLabelValues(method, path, status).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	})
}
