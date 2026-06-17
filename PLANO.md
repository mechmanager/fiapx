# Plano de Execução — Hackathon FIAP X (Sistema de Processamento de Vídeos)

## Contexto

A FIAP X possui um **monolito Go** (`main.go`, ~500 linhas, Gin) que recebe upload de
vídeo, extrai frames com ffmpeg (`fps=1`) e devolve um `.zip`. Não tem autenticação,
banco, fila nem escala. O desafio é reescrever isso como uma **arquitetura de
microsserviços escalável, segura e com boas práticas**, atendendo:

- Processar vários vídeos simultaneamente sem perder requisições em picos.
- Proteção por usuário/senha.
- Listagem de status dos vídeos por usuário.
- Notificação ao usuário em caso de erro.
- Persistência, escalabilidade, versionamento, testes (≥80%) e CI/CD.

O código reaproveitável do monolito está mapeado:

| Função no `main.go` | Onde vai parar |
|---|---|
| `processVideo()` (ffmpeg `fps=1`) | `worker/processor` |
| `createZipFile()` + `addFileToZip()` | `worker/processor` |
| `isValidVideoFile()` | `api-gateway/service` (validação no upload) |
| `handleVideoUpload()` (parse multipart) | `api-gateway/handler` |
| `handleDownload()` | `api-gateway/handler` (agora via MinIO) |

---

## Decisões a confirmar (recomendações marcadas com ✅)

1. **Estrutura do repo** — ✅ **Monorepo no repo atual `fiapx`**: mover o monolito
   para `legacy/` e criar `api-gateway/` e `worker/` na raiz. Entrega 1 link GitHub.
2. **Notificação de erro** — ✅ **Log estruturado** via interface `Notifier`
   (impl. `LogNotifier`), com a interface pronta para trocar por SMTP depois.
   Atende o requisito sem dependência externa frágil de demonstrar.
3. **Orquestração** — ✅ **Docker Compose** (o requisito é "Compose OU K8s").
   Opcional: adicionar `k8s/` depois se sobrar tempo.
4. **Frontend** — ✅ **Frontend HTML simples** (login + upload + lista de status)
   servido pela api-gateway, para facilitar a demonstração no vídeo de 10 min.

> Se você discordar de qualquer item, me avise e eu ajusto o plano antes de codar.

---

## Arquitetura alvo

```
                          ┌─────────────┐
   Browser / curl  ─────▶ │ api-gateway │ ──┐ JWT, upload, status, download
                          │   (Gin)     │   │
                          └─────────────┘   │
                            │   │   │        │
              ┌─────────────┘   │   └────────┴──────────┐
              ▼                 ▼                        ▼
        ┌──────────┐      ┌──────────┐            ┌──────────┐
        │PostgreSQL│      │  MinIO   │            │ RabbitMQ │
        │ (users,  │      │ (vídeos, │            │  (fila   │
        │  videos) │      │  zips)   │            │ video.   │
        └──────────┘      └──────────┘            │ process) │
              ▲                 ▲                  └────┬─────┘
              │                 │                       │ consume (prefetch=5)
              │                 │                       ▼
              │                 │                 ┌──────────┐
              └─────────────────┴─────────────────│  worker  │ ffmpeg + zip
                  atualiza status / lê e grava     └──────────┘
```

**Resiliência a picos**: o upload só publica na fila e responde rápido (`PENDING`).
A fila RabbitMQ é **durável** com mensagens persistentes + `ack` manual, então nada
se perde se o worker cair. Escala horizontal = subir N réplicas do worker.

---

## Stack

Go 1.22 · Gin · pgx/v5 · amqp091-go · minio-go/v7 · golang-jwt/v5 · bcrypt ·
testify + mockery · ffmpeg · Docker Compose · GitHub Actions + SonarCloud.

---

## Estrutura de pastas

```
fiapx/
├── legacy/                      # monolito original (referência)
├── api-gateway/
│   ├── cmd/api/main.go
│   ├── config/                  # carga de env vars
│   ├── domain/                  # structs + interfaces (User, Video, ports)
│   ├── handler/                 # auth, video, health
│   ├── service/                 # auth (jwt+bcrypt), video (orquestra repo/storage/queue)
│   ├── repository/              # postgres (pgx)
│   ├── middleware/              # JWT auth
│   ├── storage/                 # MinIO client (upload/presign)
│   ├── queue/                   # publisher RabbitMQ
│   ├── web/                     # HTML embutido (//go:embed)
│   ├── mocks/                   # gerados por mockery
│   ├── Dockerfile
│   └── go.mod
├── worker/
│   ├── cmd/worker/main.go
│   ├── config/
│   ├── domain/
│   ├── consumer/                # consumer RabbitMQ (prefetch=5, ack manual)
│   ├── processor/               # ffmpeg fps=1 + zip (reaproveita monolito)
│   ├── storage/                 # MinIO (download vídeo / upload zip)
│   ├── repository/              # postgres (update status)
│   ├── notification/            # interface Notifier + LogNotifier
│   ├── mocks/
│   ├── Dockerfile
│   └── go.mod
├── scripts/init.sql
├── .github/workflows/ci.yml
├── docker-compose.yml
├── ARCHITECTURE.md
├── README.md
└── sonar-project.properties
```

---

## Banco de dados (`scripts/init.sql`)

Tabelas `users` e `videos` conforme spec, com:
- `gen_random_uuid()` (extensão `pgcrypto`).
- Índice em `videos(user_id)` para a listagem por usuário.
- Trigger `updated_at` ou atualização explícita no repositório.
- Status: `PENDING` → `PROCESSING` → `DONE` | `ERROR`.

---

## Endpoints

| Método | Rota | Auth | Descrição |
|---|---|---|---|
| GET | `/health` | público | `{status:"ok"}` |
| POST | `/auth/register` | público | cria usuário (bcrypt) |
| POST | `/auth/login` | público | retorna JWT |
| POST | `/videos` | JWT | upload → MinIO + `PENDING` + publica na fila |
| GET | `/videos` | JWT | lista vídeos do usuário com status |
| GET | `/videos/:id/download` | JWT | stream do zip do MinIO (valida dono) |

---

## Fases (com dependências)

```
Fase 0  ──▶ Fase 1 ──▶ Fase 2 ──▶ Fase 3 ──┐
   (setup)   (infra/    (api auth) (api      │
             docker)               videos)   │
                          └──────────────────┼──▶ Fase 4 (worker)
                                             │
                          Fases 2,3,4 ───────┴──▶ Fase 5 (testes 80%)
                                                      └──▶ Fase 6 (CI/CD)
                                                              └──▶ Fase 7 (docs)
```

### Fase 0 — Setup do monorepo
- Mover `main.go`, `go.mod`, `go.sum`, `Dockerfile` → `legacy/`.
- Criar os dois módulos Go (`api-gateway`, `worker`) e a árvore de pastas.
- **Entregável**: estrutura compila vazia (`go build ./...`).

### Fase 1 — Infra local (Docker Compose + init.sql)
- `docker-compose.yml`: postgres:16, rabbitmq:3-management, minio, api-gateway, worker.
- Healthchecks + `depends_on` para ordem de subida.
- `scripts/init.sql` montado no postgres; bucket MinIO criado na subida.
- **Entregável**: `docker compose up` sobe a infra (sem app ainda) e o banco inicializa.

### Fase 2 — api-gateway: autenticação
- `domain` (User, interfaces), `config`, `repository/postgres` (users).
- `service/auth` (bcrypt + JWT), `handler/auth` (`/register`, `/login`).
- `middleware/jwt`, `handler/health`.
- **Entregável**: registrar e logar usuário, receber token, rota protegida responde 401 sem token.

### Fase 3 — api-gateway: upload, listagem e download
- `storage/minio` (upload vídeo, presign/stream zip).
- `queue/publisher` (publica `{video_id, user_id, s3_key}` na fila durável).
- `service/video` + `handler/video` (`POST /videos`, `GET /videos`, `GET /videos/:id/download`).
- Frontend HTML embutido (`//go:embed`).
- **Entregável**: upload grava no MinIO + `PENDING` no banco + mensagem na fila; listagem e download funcionam.

### Fase 4 — worker
- `consumer` (RabbitMQ, `prefetch=5`, ack manual, requeue em falha transitória).
- `processor` (download do MinIO → ffmpeg `fps=1` → zip → upload zip no MinIO) — reusa lógica do monolito.
- `repository` (update `PROCESSING`/`DONE`/`ERROR`, `frame_count`, `zip_s3_key`).
- `notification` (LogNotifier em caso de `ERROR`).
- **Entregável**: ponta a ponta — upload vira zip baixável; vários vídeos em paralelo.

### Fase 5 — Testes (cobertura ≥80%)
- `mockery` para gerar mocks de todas as interfaces.
- Testes `package _test` cobrindo happy path + erros em service/handler/processor/consumer.
- Script de cobertura excluindo `mocks/`, `cmd/`, `docs`.
- **Entregável**: `go test ./... -cover` ≥ 80% nos dois módulos.

### Fase 6 — CI/CD (GitHub Actions + SonarCloud)
- `ci.yml`: checkout (`fetch-depth: 0`) → setup-go → `gofmt` check → `go vet` →
  testes com cobertura → gate ≥80% → SonarCloud scan → build das imagens Docker.
- `sonar-project.properties` (você precisará configurar `SONAR_TOKEN` no repo).
- **Entregável**: pipeline verde no push.

### Fase 7 — Documentação final
- `ARCHITECTURE.md`: diagrama Mermaid de componentes + diagrama de sequência + decisões.
- `README.md`: setup (`docker compose up`), exemplos `curl` de cada endpoint, decisões técnicas.
- **Entregável**: docs completas para o vídeo de 10 min.

---

## Padrões de código (aplicados em todas as fases)

- Dependências externas atrás de **interfaces** (ports) → testáveis com mocks.
- Mocks via **mockery**; testes em `package _test`.
- **Comentários em português** nas linhas executáveis; variáveis descritivas, sem abreviações.
- Tratamento de erro explícito em todos os pontos.

---

## Como verificar (end-to-end)

1. `docker compose up --build` sobe tudo.
2. `curl POST /auth/register` e `/auth/login` → obter token.
3. `curl POST /videos` com um `.mp4` e o token → recebe `video_id` + `PENDING`.
4. `curl GET /videos` → status muda `PENDING`→`PROCESSING`→`DONE`.
5. `curl GET /videos/:id/download` → baixa o zip com os frames.
6. Subir 3+ uploads juntos → confirmar processamento paralelo (logs do worker / RabbitMQ UI).
7. Enviar arquivo inválido para forçar `ERROR` → confirmar notificação no log.
8. `go test ./... -cover` nos dois módulos ≥ 80%.

---

## Por onde começar

**Fase 0 (setup do monorepo)** — é pré-requisito de tudo e não tem risco.
Assim que você validar este plano (e confirmar as 4 decisões acima), eu começo por ela.
```
```
