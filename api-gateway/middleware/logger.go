package middleware

import (
	"log"
	"net/http"
	"time"
)

// Logger é o middleware de log de requisições HTTP do api-gateway.
type Logger struct{}

// NewLogger cria um Logger.
func NewLogger() *Logger { return &Logger{} }

// Middleware registra method, path, status e duração de cada requisição.
// Ignora /metrics e /health para não poluir os logs com polling.
func (l *Logger) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/metrics" || r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		rw := newResponseWriter(w)
		start := time.Now()

		next.ServeHTTP(rw, r)

		log.Printf("service=api-gateway method=%s path=%s status=%d duration=%s ip=%s",
			r.Method, r.URL.Path, rw.statusCode,
			time.Since(start).Round(time.Millisecond), r.RemoteAddr)
	})
}
