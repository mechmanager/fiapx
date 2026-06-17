# Arquitetura — FIAP X

## 1. Visão geral de componentes

<img width="1520" height="1240" alt="image" src="https://github.com/user-attachments/assets/9075dd3f-c7c9-48f1-8210-714195f6f786" />

---

## 2. Diagrama de sequência — Upload e processamento

<img width="1560" height="1040" alt="image" src="https://github.com/user-attachments/assets/fb9f37ac-84d0-4b36-8de3-dbfe5bac9c91" />


---

## 3. Decisões técnicas

### API Gateway sem domínio

O gateway é um **proxy puro** (`net/http/httputil.ReverseProxy`). Não acessa banco,
não conhece usuário nem vídeo. Responsabilidades exclusivas: validar JWT, injetar
`X-User-ID` no header e fazer rate limiting por IP. Cada microsserviço downstream
confia no header sem re-validar o token.

### Separação de responsabilidades por serviço

| Serviço | Escreve em | Lê de |
|---------|-----------|-------|
| Auth | `users` (PostgreSQL) | `users` |
| Upload | `videos` INSERT (PostgreSQL) · MinIO · `video.upload` | — |
| Status | — | `videos` (PostgreSQL + Redis) · MinIO |
| Worker | `videos` UPDATE (PostgreSQL) · MinIO · `notification` | MinIO · `video.upload` |
| Notification | — | `users` (PostgreSQL) |

### Hexagonal / Ports & Adapters

Cada serviço define interfaces de domínio (ports) em `domain/`. Adapters concretos
ficam em subpacotes (`repository/`, `storage/`, `queue/`, `cache/`), substituíveis
sem alterar a lógica de negócio.

```
domain.UserRepository    ←  repository.UserRepo      (pgx)
domain.ObjectStorage     ←  storage.MinIOStorage     (minio-go)
domain.QueuePublisher    ←  queue.RabbitMQPublisher  (amqp091)
domain.StatusCache       ←  cache.RedisCache         (go-redis)
pipeline.VideoProcessor  ←  processor.Processor      (ffmpeg)
```

### Resiliência da fila com Dead-Letter Queue

Mensagens que falham 3 vezes seguidas (`MAX_RETRIES=3`) recebem `nack` sem requeue
e são roteadas automaticamente para a DLQ (`video.upload.dlq`). Isso evita que uma
mensagem corrompida bloqueie workers indefinidamente.

### Cache de status (Redis)

O Status Service implementa **cache-aside**: tenta Redis primeiro (TTL 30 s); em miss
busca no PostgreSQL e repopula o cache. O Worker atualiza o cache após cada `UPDATE`
no banco.

### Escalabilidade horizontal

O Worker não tem estado local — N réplicas consomem a mesma fila em paralelo com
`prefetch=5`. Em Kubernetes, o HPA escala de 2 a 20 réplicas por CPU; para escalar
por profundidade de fila usar KEDA `ScaledObject` com RabbitMQ scaler.

### Notificação desacoplada

O Worker publica em `notification` apenas em caso de erro. O Notification Service
consome essa fila de forma independente e envia e-mail via SMTP. Os dois serviços
evoluem sem acoplamento.

---

## 4. Estrutura de pastas

```
fiapx/
├── api-gateway/              # proxy puro — sem domínio
│   ├── cmd/gateway/main.go
│   ├── config/
│   ├── middleware/           # jwt_check · rate_limit · metrics
│   ├── proxy/                # reverse_proxy.go · router.go
│   └── go.mod
│
├── auth-service/
│   ├── cmd/auth/main.go
│   ├── config/
│   ├── domain/               # User · interfaces
│   ├── handler/              # register · login
│   ├── service/              # AuthService (bcrypt + JWT)
│   ├── repository/           # UserRepo (pgx)
│   ├── mocks/
│   └── go.mod
│
├── upload-service/
│   ├── cmd/upload/main.go
│   ├── config/
│   ├── domain/               # Video · interfaces
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
│   ├── handler/              # list · download
│   ├── service/              # StatusService
│   ├── repository/           # VideoRepo (pgx)
│   ├── cache/                # RedisCache
│   ├── storage/              # MinIOStorage
│   ├── mocks/
│   └── go.mod
│
├── worker/
│   ├── cmd/worker/main.go
│   ├── config/
│   ├── consumer/             # AMQP adapter (DLQ · prefetch · ack)
│   ├── pipeline/             # lógica testável
│   ├── processor/            # ffmpeg fps=1 + createZip
│   ├── repository/           # UpdateStatus · UpdateDone (pgx)
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
├── k8s/                      # Deployment por serviço + Redis
├── terraform/                # AWS EKS
├── legacy/                   # monolito original (referência)
├── scripts/
│   ├── init.sql
│   └── coverage.sh           # gate ≥ 80% por módulo
├── .github/workflows/ci.yml
├── docker-compose.yml
├── sonar-project.properties
└── .env.example
```
