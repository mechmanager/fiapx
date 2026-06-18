package main

import (
	"log"
	"net/http"
	"time"

	"github.com/mechmanager/fiapx/api-gateway/config"
	"github.com/mechmanager/fiapx/api-gateway/middleware"
	"github.com/mechmanager/fiapx/api-gateway/proxy"
	apiweb "github.com/mechmanager/fiapx/api-gateway/web"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	authProxy, err := proxy.New(cfg.AuthServiceURL)
	if err != nil {
		log.Fatalf("proxy auth inválido: %v", err)
	}
	uploadProxy, err := proxy.New(cfg.UploadServiceURL)
	if err != nil {
		log.Fatalf("proxy upload inválido: %v", err)
	}
	statusProxy, err := proxy.New(cfg.StatusServiceURL)
	if err != nil {
		log.Fatalf("proxy status inválido: %v", err)
	}

	jwtChecker := middleware.NewJWTChecker(cfg.JWTSecret)
	rateLimiter := middleware.NewIPRateLimiter(cfg.RateLimitRPS)
	metricsMW := middleware.NewMetrics()

	mux := http.NewServeMux()

	mux.Handle("GET /metrics", promhttp.Handler())

	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if _, err := w.Write([]byte(`{"status":"ok","service":"api-gateway"}`)); err != nil {
			log.Printf("health write error: %v", err)
		}
	})

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		data, err := apiweb.FS.ReadFile("index.html")
		if err != nil {
			http.Error(w, "frontend não encontrado", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if _, err := w.Write(data); err != nil {
			log.Printf("frontend write error: %v", err)
		}
	})

	mux.Handle("POST /auth/register", authProxy)
	mux.Handle("POST /auth/login", authProxy)
	mux.Handle("POST /videos", jwtChecker.Middleware(uploadProxy))
	mux.Handle("GET /videos", jwtChecker.Middleware(statusProxy))
	mux.Handle("GET /videos/{id}/download", jwtChecker.Middleware(statusProxy))

	srv := &http.Server{
		Addr:         ":" + cfg.GatewayPort,
		Handler:      rateLimiter.Middleware(metricsMW.Middleware(mux)),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("api-gateway escutando na porta %s", cfg.GatewayPort)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
