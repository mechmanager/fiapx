# FIAP X — Plataforma de Processamento de Vídeos

Sistema de processamento de vídeos em arquitetura de microsserviços Go.  
Recebe vídeos via API REST, extrai frames com ffmpeg e disponibiliza um `.zip` para download.

> **Como executar localmente → [COMO_EXECUTAR.md](COMO_EXECUTAR.md)**

---

## Arquitetura

Veja [ARCHITECTURE.md](ARCHITECTURE.md) para diagramas de componentes e sequência.

### Microsserviços

| Serviço | Responsabilidade | Porta |
|---------|-----------------|-------|
| **api-gateway** | Proxy reverso, JWT, rate limit, serve o frontend React | 8080 |
| **auth-service** | Cadastro, login, geração de JWT | 8081 |
| **upload-service** | Recebe o vídeo, salva no MinIO, publica na fila | 8082 |
| **status-service** | Listagem de vídeos e download do ZIP de frames | 8083 |
| **worker** | Consome a fila, extrai frames com ffmpeg, compacta em ZIP | — |
| **notification-service** | Consome notificações de erro e envia e-mail via SMTP | — |

### Infraestrutura

| Componente | Função |
|------------|--------|
| **PostgreSQL** | Persistência de usuários e metadados de vídeos |
| **RabbitMQ** | Fila durável `video.upload` com DLQ, prefetch 5, ack manual |
| **MinIO** | Object storage S3-compatível para vídeos e ZIPs |
| **Redis** | Cache de metadados no status-service |
| **Prometheus** | Coleta de métricas de todos os serviços |
| **Grafana** | Dashboard de monitoramento (provisionado automaticamente) |

---

## Fluxo de processamento

```
Usuário → api-gateway → upload-service → MinIO (vídeo)
                                       → RabbitMQ (mensagem)
                                               ↓
                                           worker
                                       → ffmpeg (frames)
                                       → ZIP → MinIO
                                       → PostgreSQL (status DONE)
                                       → notification-service (em caso de erro)
```

**Estados do vídeo:**
```
PENDING → PROCESSING → DONE
                    ↘ ERROR
```

---

## API REST

Todos os endpoints passam pelo api-gateway em `http://localhost:8080`.

### Autenticação

```bash
# Cadastro
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Seu Nome","email":"voce@email.com","password":"senha123"}'

# Login — retorna JWT
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"voce@email.com","password":"senha123"}' | jq -r '.token')
```

### Vídeos

```bash
# Upload
curl -s -X POST http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" \
  -F "video=@video.mp4"

# Listar com status
curl -s http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" | jq

# Baixar ZIP de frames
curl -OJ http://localhost:8080/videos/<uuid>/download \
  -H "Authorization: Bearer $TOKEN"
```

---

## Qualidade de Software

| Aspecto | Ferramenta | Gate |
|---------|-----------|------|
| Cobertura de testes | `go test -coverprofile` | ≥ 80% por serviço |
| Análise estática | SonarCloud | Reliability A, Duplication ≤ 3% |
| Formatação | `gofmt` | Zero diff |
| Lint | `go vet` | Zero warnings |

```bash
# Rodar testes localmente (requer Go 1.25+)
cd api-gateway && go test ./... -cover
cd worker      && go test ./config/... ./pipeline/... ./processor/... -cover
```

---

## CI/CD

Pipeline GitHub Actions em `.github/workflows/` — um workflow por serviço:

| Job | O que faz |
|-----|-----------|
| `lint-and-test` | `gofmt` + `go vet` + `go test` com gate de cobertura |
| `docker` | Build e push da imagem para `ghcr.io` (apenas na `main`) |
| `sonar` | Análise SonarCloud (requer secret `SONAR_TOKEN`) |

**Secrets necessários no repositório:**

| Secret | Descrição |
|--------|-----------|
| `SONAR_TOKEN` | Token de autenticação do SonarCloud |
| `GITHUB_TOKEN` | Automático — para push no GHCR |

---

## Observabilidade

```
http://localhost:9090   → Prometheus
http://localhost:3000   → Grafana (dashboard auto-provisionado)
```

Métricas expostas:
- `gateway_http_requests_total` — requisições por rota/método/status
- `gateway_http_request_duration_seconds` — histograma de latência
- `worker_videos_processed_total` — contagem por resultado (done/error)
- `worker_processing_duration_seconds` — histograma de duração do processamento
- Métricas Go runtime (goroutines, GC, memória) em todos os serviços
- Métricas RabbitMQ (fila, consumers)

---

## Escalabilidade horizontal

```bash
# Escalar workers para processar vídeos em paralelo
docker compose up --scale worker=3 -d
```

Cada instância do worker consome independentemente da mesma fila RabbitMQ.  
O sistema não perde requisições em picos — mensagens ficam enfileiradas até um worker estar disponível.

---

## Variáveis de ambiente

Veja o arquivo `.env` na raiz para a configuração completa.  
**Nunca versione o arquivo `.env`.**

---

## Vídeos de exemplo

A pasta `samples/` contém clips do **Big Buck Bunny** (Blender Foundation — CC BY 3.0)  
prontos para testar o pipeline de processamento.
