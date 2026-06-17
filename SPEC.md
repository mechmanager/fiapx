# FIAP X — Spec v2: Arquitetura de Microsserviços

> **Status:** proposta para revisão do time  
> **Data:** 2026-06-16  
> **Motivação:** extrair o domínio do api-gateway monolítico em serviços independentes e
> introduzir um gateway puro de roteamento, cache de status, notificação desacoplada e
> observabilidade.

---

## 1. Visão geral

```
                        ┌─────────────────────────────────┐
  Browser / curl ──────▶│         API Gateway             │
                        │  routing · rate limit · JWT check│
                        └──────────┬──────────────────────┘
                                   │ injeta X-User-ID header
              ┌────────────────────┼────────────────────┐
              ▼                    ▼                    ▼
       ┌────────────┐     ┌──────────────┐     ┌──────────────┐
       │    Auth    │     │    Upload    │     │    Status    │
       │  Service   │     │   Service   │     │   Service   │
       │            │     │             │     │             │
       │/auth/reg   │     │POST /videos │     │GET /videos  │
       │/auth/login │     │             │     │GET /:id/dl  │
       └─────┬──────┘     └──────┬──────┘     └──────┬──────┘
             │                   │                    │
             ▼                   ▼                    ▼
       ┌──────────┐       ┌────────────┐       ┌──────────┐
       │PostgreSQL│       │   MinIO    │       │  Redis   │
       │  users   │       │  vídeos   │       │  cache   │
       └──────────┘       └─────┬──────┘       └──────────┘
                                │ publica
                                ▼
                    ┌──────────────────────┐
                    │  RabbitMQ            │
                    │  video.upload  ──────┼──▶ Worker (N réplicas)
                    │  notification  ◀─────┼─── Worker (erro)
                    │  *.dlq               │    Notification Service
                    └──────────────────────┘
                                │
                    ┌───────────┴───────────┐
                    ▼                       ▼
             ┌────────────┐       ┌──────────────────┐
             │   Worker   │       │  Notification    │
             │  Service   │       │    Service       │
             │ ffmpeg+zip │       │  e-mail SMTP     │
             └─────┬──────┘       └──────────────────┘
                   │
                   ▼
              ┌──────────┐
              │PostgreSQL│  UPDATE status
              │  videos  │  UPDATE zip_s3_key
              └──────────┘

  Observabilidade: Prometheus + Grafana (scrape de todos os serviços)
```

---

## 2. Serviços

### 2.1 API Gateway

**Responsabilidade única:** roteamento, rate limiting e validação de JWT. Não tem domínio próprio — não acessa banco nem fila.

| Item | Decisão |
|------|---------|
| Implementação | Go `net/http/httputil.ReverseProxy` |
| Porta | 8080 (ponto de entrada único) |
| JWT | valida `Authorization: Bearer <token>` com `JWT_SECRET` compartilhado; injeta `X-User-ID` no header para downstream |
| Rate limit | sliding-window in-memory por IP (`golang.org/x/time/rate`) |
| Rotas públicas | `/health`, `/auth/*` — bypass do JWT check |
| Health check | `GET /health` → `{"status":"ok"}` |

**Tabela de roteamento:**

| Método | Path | Upstream | JWT obrigatório |
|--------|------|----------|-----------------|
| `*` | `/health` | próprio | não |
| `POST` | `/auth/register` | auth-service | não |
| `POST` | `/auth/login` | auth-service | não |
| `POST` | `/videos` | upload-service | sim |
| `GET` | `/videos` | status-service | sim |
| `GET` | `/videos/:id/download` | status-service | sim |

---

### 2.2 Auth Service

**Responsabilidade:** cadastro de usuário, login e emissão de JWT.

| Item | Decisão |
|------|---------|
| Porta | 8081 (interna — só o gateway acessa) |
| Banco | PostgreSQL, tabela `users` |
| JWT | HS256, expira em `JWT_EXPIRATION_HOURS` (padrão 24 h) |
| Senha | bcrypt cost 12 |

**Endpoints:**

```
POST /auth/register
Body:  { "email": "string", "password": "string (min 8 chars)" }
201:   { "id": "uuid", "email": "string" }
409:   { "error": "email already registered" }

POST /auth/login
Body:  { "email": "string", "password": "string" }
200:   { "token": "jwt_string" }
401:   { "error": "invalid credentials" }
```

---

### 2.3 Upload Service

**Responsabilidade:** receber o arquivo de vídeo, persistir no MinIO e publicar mensagem na fila `video.upload`.

| Item | Decisão |
|------|---------|
| Porta | 8082 (interna) |
| Banco | PostgreSQL, tabela `videos` (INSERT com status `PENDING`) |
| Storage | MinIO — objeto `videos/<video_id>.<ext>` |
| Fila | publica em `video.upload` (durable, persistent) |
| Formatos aceitos | `.mp4 .avi .mov .mkv .webm` |
| User ID | lido do header `X-User-ID` (injetado pelo gateway) |

**Endpoint:**

```
POST /videos
Header: X-User-ID: <uuid>          (injetado pelo gateway)
Body:   multipart/form-data; file=<video>
202:    { "id": "uuid", "status": "PENDING" }
400:    { "error": "unsupported file type" }
```

**Mensagem publicada em `video.upload`:**

```json
{
  "video_id":  "uuid",
  "user_id":   "uuid",
  "s3_key":    "videos/<video_id>.mp4",
  "filename":  "nome_original.mp4"
}
```

---

### 2.4 Status Service

**Responsabilidade:** listar vídeos do usuário e entregar o zip para download.

| Item | Decisão |
|------|---------|
| Porta | 8083 (interna) |
| Banco | PostgreSQL, tabela `videos` (SELECT) |
| Cache | Redis — chave `video:<video_id>:status`, TTL 30 s |
| Storage | MinIO — stream do objeto `zips/<video_id>.zip` |
| User ID | lido do header `X-User-ID` |

**Endpoints:**

```
GET /videos
Header: X-User-ID: <uuid>
200: [
  {
    "id":          "uuid",
    "filename":    "video.mp4",
    "status":      "PENDING | PROCESSING | DONE | ERROR",
    "frame_count": 42,
    "error":       "string | null",
    "created_at":  "RFC3339"
  }
]

GET /videos/:id/download
Header: X-User-ID: <uuid>
200: application/octet-stream  (stream do zip)
Content-Disposition: attachment; filename="frames_<id>.zip"
403: { "error": "forbidden" }           (dono diferente)
409: { "error": "video not ready" }     (status != DONE)
404: { "error": "not found" }
```

---

### 2.5 Worker Service

**Responsabilidade:** consumir a fila `video.upload`, processar o vídeo com ffmpeg, gerar o zip, atualizar o status no banco e publicar em `notification` em caso de erro.

| Item | Decisão |
|------|---------|
| Concorrência | prefetch = 5, ack manual |
| Retry | nack sem requeue após 3 falhas → DLQ `video.upload.dlq` |
| Banco | PostgreSQL — `UPDATE status = PROCESSING / DONE / ERROR` |
| Storage | MinIO — download do vídeo original, upload do zip |
| Fila entrada | `video.upload` |
| Fila saída | `notification` (apenas em caso de `ERROR`) |

**Fluxo interno:**

```
1. Consume message → parse JSON
2. UPDATE videos SET status='PROCESSING'
3. GET MinIO videos/<video_id>
4. ffmpeg -i input.mp4 -vf fps=1 frames/%04d.png
5. createZip(frames/) → frames_<video_id>.zip
6. PUT MinIO zips/<video_id>.zip
7. UPDATE videos SET status='DONE', zip_s3_key=..., frame_count=...
8. ack
   ↳ erro em qualquer step:
     UPDATE videos SET status='ERROR', error_message=...
     publish → notification queue
     nack (sem requeue se tentativas >= 3)
```

**Mensagem publicada em `notification`:**

```json
{
  "user_id":   "uuid",
  "video_id":  "uuid",
  "filename":  "video.mp4",
  "error":     "descrição do erro"
}
```

---

### 2.6 Notification Service

**Responsabilidade:** consumir a fila `notification` e enviar e-mail ao usuário via SMTP.

| Item | Decisão |
|------|---------|
| Fila entrada | `notification` (durable, persistent) |
| DLQ | `notification.dlq` |
| SMTP | `net/smtp` com variáveis `SMTP_HOST`, `SMTP_PORT`, `SMTP_USER`, `SMTP_PASS`, `SMTP_FROM` |
| Banco | PostgreSQL — busca `users.email` pelo `user_id` da mensagem |

**Template de e-mail:**

```
Assunto: [FIAP X] Erro ao processar seu vídeo
Corpo:   Olá, houve um erro ao processar "<filename>".
         Detalhes: <error>
```

---

## 3. Banco de dados

### Schema (PostgreSQL)

```sql
-- Extensão para UUIDs
CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE users (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  email         TEXT        UNIQUE NOT NULL,
  password_hash TEXT        NOT NULL,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE videos (
  id            UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id       UUID        NOT NULL REFERENCES users(id),
  filename      TEXT        NOT NULL,
  s3_key        TEXT        NOT NULL,
  zip_s3_key    TEXT,
  status        TEXT        NOT NULL DEFAULT 'PENDING'
                            CHECK (status IN ('PENDING','PROCESSING','DONE','ERROR')),
  frame_count   INT,
  error_message TEXT,
  created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_videos_user_id ON videos(user_id);
```

**Ownership por serviço:**

| Tabela | Escrita | Leitura |
|--------|---------|---------|
| `users` | auth-service | auth-service, notification-service |
| `videos` | upload-service (INSERT), worker (UPDATE) | status-service, worker |

---

## 4. Mensageria (RabbitMQ)

### Filas

| Fila | Tipo | Produtores | Consumidores |
|------|------|------------|--------------|
| `video.upload` | durable | upload-service | worker |
| `video.upload.dlq` | durable | RabbitMQ (após 3 nacks) | — (monitoramento) |
| `notification` | durable | worker | notification-service |
| `notification.dlq` | durable | RabbitMQ | — (monitoramento) |

### Configuração das DLQs

```
x-dead-letter-exchange: ""
x-dead-letter-routing-key: "<fila>.dlq"
x-message-ttl: 86400000   (24 h — mensagens expiram da DLQ após 1 dia)
```

---

## 5. Cache (Redis)

| Chave | Valor | TTL | Escrita | Leitura |
|-------|-------|-----|---------|---------|
| `video:<id>:status` | `PENDING\|PROCESSING\|DONE\|ERROR` | 30 s | worker (após UPDATE) | status-service |

O status-service tenta o cache primeiro; em miss busca no PostgreSQL e repopula.

---

## 6. Variáveis de ambiente

### Compartilhadas por mais de um serviço

| Variável | Serviços |
|----------|---------|
| `POSTGRES_DSN` | auth, upload, status, worker, notification |
| `JWT_SECRET` | api-gateway (validar), auth (emitir) |
| `RABBITMQ_URL` | upload, worker, notification |
| `MINIO_ENDPOINT` | upload, worker, status |
| `MINIO_ACCESS_KEY` | upload, worker, status |
| `MINIO_SECRET_KEY` | upload, worker, status |
| `MINIO_BUCKET` | upload, worker, status |
| `REDIS_URL` | status, worker |

### Por serviço

| Serviço | Variável | Exemplo |
|---------|---------|---------|
| api-gateway | `AUTH_SERVICE_URL` | `http://auth-service:8081` |
| api-gateway | `UPLOAD_SERVICE_URL` | `http://upload-service:8082` |
| api-gateway | `STATUS_SERVICE_URL` | `http://status-service:8083` |
| api-gateway | `RATE_LIMIT_RPS` | `100` |
| auth | `JWT_EXPIRATION_HOURS` | `24` |
| worker | `QUEUE_NAME` | `video.upload` |
| worker | `PREFETCH_COUNT` | `5` |
| worker | `MAX_RETRIES` | `3` |
| notification | `SMTP_HOST` | `smtp.gmail.com` |
| notification | `SMTP_PORT` | `587` |
| notification | `SMTP_USER` | `noreply@fiapx.com` |
| notification | `SMTP_PASS` | `...` |
| notification | `SMTP_FROM` | `FIAP X <noreply@fiapx.com>` |

---

## 7. Observabilidade

Cada serviço Go expõe `/metrics` (Prometheus format) via `promhttp.Handler()`.

| Métrica | Serviço | Tipo |
|---------|---------|------|
| `http_requests_total{method,path,status}` | gateway, auth, upload, status | Counter |
| `http_request_duration_seconds{path}` | gateway, auth, upload, status | Histogram |
| `queue_messages_consumed_total{queue,result}` | worker, notification | Counter |
| `video_processing_duration_seconds` | worker | Histogram |

Grafana: dashboards pré-configurados em `monitoring/grafana/dashboards/`.

---

## 8. Estrutura de pastas (alvo)

```
fiapx/
├── api-gateway/              # proxy puro — sem domínio
│   ├── cmd/gateway/main.go
│   ├── config/
│   ├── middleware/           # jwt_check, rate_limit, metrics
│   ├── proxy/                # reverse_proxy.go, router.go
│   └── go.mod
│
├── auth-service/
│   ├── cmd/auth/main.go
│   ├── config/
│   ├── domain/               # User, interfaces
│   ├── handler/              # register, login
│   ├── service/              # AuthService (bcrypt, jwt)
│   ├── repository/           # UserRepo (pgx)
│   ├── mocks/
│   └── go.mod
│
├── upload-service/
│   ├── cmd/upload/main.go
│   ├── config/
│   ├── domain/               # Video, interfaces
│   ├── handler/              # upload
│   ├── service/              # UploadService
│   ├── repository/           # VideoRepo (pgx)
│   ├── storage/              # MinIOStorage
│   ├── queue/                # RabbitMQPublisher
│   ├── mocks/
│   └── go.mod
│
├── status-service/
│   ├── cmd/status/main.go
│   ├── config/
│   ├── domain/
│   ├── handler/              # list, download
│   ├── service/              # StatusService
│   ├── repository/           # VideoRepo (pgx)
│   ├── cache/                # RedisCache
│   ├── storage/              # MinIOStorage
│   ├── mocks/
│   └── go.mod
│
├── worker/                   # refatorar: DLQ + publish notification
│   ├── cmd/worker/main.go
│   ├── config/
│   ├── consumer/             # AMQP adapter (DLQ aware)
│   ├── pipeline/             # lógica testável
│   ├── processor/            # ffmpeg + zip
│   ├── repository/           # VideoRepo
│   ├── storage/              # MinIOStorage
│   ├── notification/         # RabbitMQNotifier (publica em fila)
│   ├── mocks/
│   └── go.mod
│
├── notification-service/
│   ├── cmd/notification/main.go
│   ├── config/
│   ├── consumer/             # AMQP consumer
│   ├── mailer/               # SMTPMailer
│   ├── repository/           # UserRepo (busca email)
│   ├── mocks/
│   └── go.mod
│
├── monitoring/
│   ├── prometheus.yml
│   └── grafana/dashboards/
│
├── k8s/                      # um Deployment por serviço
├── terraform/
├── scripts/
│   ├── init.sql
│   └── coverage.sh           # gate ≥ 80% por módulo
├── .github/workflows/ci.yml
├── docker-compose.yml
└── .env.example
```

---

## 9. Plano de implementação

### Dependências entre fases

```
Fase 0 ──▶ Fase 1 ──▶ Fase 2 ──▶ Fase 3 ──▶ Fase 4
(setup)    (auth)     (upload)   (status)   (gateway)
                                               │
             Fases 1–4 ──────────────────────▶ Fase 5 (worker refactor)
                                               │
                                    Fase 5 ──▶ Fase 6 (notification)
                                               │
                          Fases 1–6 ──────────▶ Fase 7 (observability)
                                               │
                          Fases 1–7 ──────────▶ Fase 8 (testes 80%)
                                               │
                                    Fase 8 ──▶ Fase 9 (CI/CD + K8s)
```

### Fase 0 — Reestruturação do monorepo

- Mover `api-gateway/` → `legacy-api-gateway/` (preservar como referência)
- Criar pastas vazias: `api-gateway/`, `auth-service/`, `upload-service/`, `status-service/`, `notification-service/`, `monitoring/`
- Inicializar `go.mod` em cada novo serviço
- Entregável: `go build ./...` em todos os módulos (vazio, sem erros de compilação)

### Fase 1 — Auth Service

- `domain/`: struct `User`, interfaces `UserRepository`, `AuthUseCase`
- `repository/postgres.go`: `Create`, `FindByEmail`
- `service/auth.go`: `Register` (bcrypt), `Login` (bcrypt + JWT)
- `handler/auth.go`: `POST /auth/register`, `POST /auth/login`
- Entregável: `curl POST /auth/register` e `/auth/login` funcionando diretamente (sem gateway ainda)

### Fase 2 — Upload Service

- `domain/`: struct `Video`, interfaces `VideoRepository`, `ObjectStorage`, `QueuePublisher`
- `repository/postgres.go`: `Create`
- `storage/minio.go`: `Upload`
- `queue/publisher.go`: `Publish` em `video.upload`
- `service/upload.go`: valida extensão → MinIO → INSERT PENDING → publish
- `handler/upload.go`: `POST /videos` (lê `X-User-ID` do header)
- Entregável: upload funciona diretamente no serviço

### Fase 3 — Status Service

- `domain/`: interfaces `VideoRepository`, `ObjectStorage`, `StatusCache`
- `repository/postgres.go`: `ListByUser`, `FindByID`
- `cache/redis.go`: `Get`, `Set`
- `storage/minio.go`: `Download`
- `service/status.go`: lista com cache-aside; download valida dono + status
- `handler/status.go`: `GET /videos`, `GET /videos/:id/download`
- Entregável: listagem e download funcionam diretamente no serviço

### Fase 4 — API Gateway

- `proxy/router.go`: tabela de roteamento configurável por env vars
- `proxy/reverse_proxy.go`: `httputil.ReverseProxy` com rewrite de headers
- `middleware/jwt_check.go`: valida token, injeta `X-User-ID`, bypass em rotas públicas
- `middleware/rate_limit.go`: `golang.org/x/time/rate`, sliding window por IP
- Entregável: todos os fluxos funcionam passando pelo gateway; rate limit rejeita com 429

### Fase 5 — Worker (refatorar)

- `notification/rabbitmq.go`: implementação de `Notifier` que publica em `notification` (substituir `LogNotifier`)
- `consumer/consumer.go`: suporte a DLQ — contador de tentativas, nack sem requeue após `MAX_RETRIES`
- Configurar `video.upload.dlq` no RabbitMQ (declaração no startup)
- Atualizar `docker-compose.yml` e K8s para passar `REDIS_URL` e `MAX_RETRIES`
- Entregável: mensagem com erro reiterativo vai para DLQ; notificação publicada na fila

### Fase 6 — Notification Service

- `consumer/consumer.go`: consume `notification` queue
- `repository/postgres.go`: `FindEmailByUserID` (busca email do usuário)
- `mailer/smtp.go`: `Send(to, subject, body string) error`
- `service/notification.go`: orquestra consumer → fetch email → send
- Configurar `notification.dlq`
- Entregável: e-mail enviado ao usuário após erro no worker

### Fase 7 — Observabilidade

- Adicionar `prometheus/client_golang` em cada serviço
- `middleware/metrics.go` no gateway: contador de requests por rota/status
- Métricas de fila no worker e notification-service
- `monitoring/prometheus.yml`: scrape configs
- `monitoring/grafana/dashboards/fiapx.json`: dashboard com request rate, latência p99, queue depth
- Adicionar `prometheus` e `grafana` ao `docker-compose.yml`
- Entregável: `docker compose up` → Grafana em `:3000` com dados em tempo real

### Fase 8 — Testes (≥ 80% por serviço)

- Mocks manuais (struct com `Fn` fields) por serviço
- Testes em `package _test` cobrindo happy path + erros em handlers, services e pipeline
- Atualizar `scripts/coverage.sh` para iterar sobre todos os módulos
- Gate: 80% em cada módulo individualmente
- Entregável: `bash scripts/coverage.sh` verde em todos os 6 serviços

### Fase 9 — CI/CD, K8s e Terraform

- `.github/workflows/ci.yml`: matrix sobre todos os módulos para lint e test; jobs `docker` buildando 6 imagens
- `k8s/`: novos Deployments para `auth-service`, `upload-service`, `status-service`, `notification-service`; atualizar `api-gateway` e `worker`; adicionar `redis.yaml`
- `kustomization.yaml`: atualizar lista de resources
- `terraform/`: sem mudança estrutural (EKS comporta N serviços)
- Entregável: pipeline verde; `kubectl apply -k k8s/` sobe tudo

---

## 10. Pontos de decisão para o time

Antes de iniciar a implementação, o time deve alinhar:

1. **Shared package vs duplicação de tipos de mensagem**
   - Opção A: pasta `shared/` com os schemas JSON (`VideoUploadMessage`, `NotificationMessage`) importada pelos serviços
   - Opção B: cada serviço define seu próprio struct (acoplamento zero, duplicação mínima)
   - Recomendação: **Opção B** — microsserviços independentes toleram duplicação de DTOs

2. **Banco compartilhado vs banco por serviço**
   - Opção A: um PostgreSQL, schemas/tabelas separadas por serviço (acesso via DSN único)
   - Opção B: um PostgreSQL por serviço (isolamento total, mais infra)
   - Recomendação: **Opção A** para hackathon — um PostgreSQL, mas cada serviço acessa só suas tabelas

3. **Redis: serviço dedicado vs sidecar**
   - Um Redis compartilhado por todos os serviços (simples, suficiente para o escopo)

4. **SMTP para e-mail**
   - Usar conta Gmail com app password para demo, ou Mailtrap para testes
   - Variáveis `SMTP_*` no `.env`

5. **Frontend (HTML)**
   - O frontend atual (`web/index.html`) continuará servido pelo API Gateway como arquivo estático
   - O gateway não precisa de lógica Go para isso — pode servir via `http.FileServer` ou mover para um serviço dedicado de frontend (fora do escopo desta spec)
