package main

import (
	"log"
	"net/http"

	"github.com/mechmanager/fiapx/api-gateway/config"
	"github.com/mechmanager/fiapx/api-gateway/middleware"
	"github.com/mechmanager/fiapx/api-gateway/proxy"
	apiweb "github.com/mechmanager/fiapx/api-gateway/web"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("configuração inválida: %v", err)
	}

	jwtChecker := middleware.NewJWTChecker(cfg.JWTSecret)
	rateLimiter := middleware.NewIPRateLimiter(cfg.RateLimitRPS)

	authProxy := proxy.New(cfg.AuthServiceURL)
	uploadProxy := proxy.New(cfg.UploadServiceURL)
	statusProxy := proxy.New(cfg.StatusServiceURL)

	mux := http.NewServeMux()

	// Saúde do próprio gateway.
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"status":"ok","service":"api-gateway"}`))
	})

	// Frontend estático embutido no binário.
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
		w.Write(data)
	})

	// Rotas públicas → auth-service (sem JWT).
	mux.Handle("POST /auth/register", authProxy)
	mux.Handle("POST /auth/login", authProxy)

	// Rotas protegidas → upload-service e status-service (JWT obrigatório).
	mux.Handle("POST /videos", jwtChecker.Middleware(uploadProxy))
	mux.Handle("GET /videos", jwtChecker.Middleware(statusProxy))
	mux.Handle("GET /videos/{id}/download", jwtChecker.Middleware(statusProxy))

	// Aplica rate limiter sobre todo o mux.
	handler := rateLimiter.Middleware(mux)

	log.Printf("api-gateway escutando na porta %s", cfg.GatewayPort)
	if err := http.ListenAndServe(":"+cfg.GatewayPort, handler); err != nil {
		log.Fatalf("erro ao iniciar servidor: %v", err)
	}
}
