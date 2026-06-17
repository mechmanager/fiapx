# FIAP X — Guia de Execução e Testes

## Pré-requisitos

| Ferramenta | Versão mínima |
|---|---|
| Docker Desktop | 24+ |
| Docker Compose | v2 (embutido no Docker Desktop) |
| Go | 1.22+ (apenas para testes unitários locais) |
| curl / httpie | qualquer (para testes de API) |

---

## 1. Subindo o ambiente completo

```bash
# Clone e entre no diretório
git clone <url-do-repo> fiapx && cd fiapx

# Copie as variáveis de ambiente (o .env já está pronto para dev local)
cp .env.example .env   # edite JWT_SECRET e senhas se desejar

# Build de todos os serviços e subida completa
docker compose up --build
```

Aguarde todos os health checks ficarem verdes. Você verá logs de:
- `fiapx-postgres` — banco pronto
- `fiapx-rabbitmq` — broker pronto
- `fiapx-minio` + `fiapx-minio-setup` — bucket `videos` criado
- `fiapx-redis` — cache pronto
- Todos os 6 serviços Go inicializados

### Portas expostas

| Serviço | URL |
|---|---|
| API Gateway (entrada única) | http://localhost:8080 |
| RabbitMQ Management | http://localhost:15672 (fiapx/fiapx123) |
| MinIO Console | http://localhost:9001 (minioadmin/minioadmin123) |
| Prometheus | http://localhost:9090 |
| Grafana | http://localhost:3000 (admin/admin) |
| PostgreSQL | localhost:5432 (fiapx/fiapx123) |

### Escalar workers horizontalmente

```bash
docker compose up --scale worker=3
```

---

## 2. Fluxo de teste manual (happy path)

### 2.1 Registrar um usuário

```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"João Silva","email":"joao@example.com","password":"senha123"}' | jq
```

Resposta esperada — `201 Created`:
```json
{"id":"<uuid>","email":"joao@example.com","name":"João Silva"}
```

### 2.2 Fazer login e obter token

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"joao@example.com","password":"senha123"}' | jq -r '.token')

echo "Token: $TOKEN"
```

### 2.3 Fazer upload de um vídeo

O projeto inclui 3 vídeos de exemplo em `samples/`:

```bash
# Upload do sample1.mp4
curl -s -X POST http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" \
  -F "video=@samples/sample1.mp4" | jq
```

Resposta esperada — `202 Accepted`:
```json
{"video_id":"<uuid>","status":"PENDING","message":"video enfileirado para processamento"}
```

### 2.4 Verificar status dos vídeos

```bash
curl -s http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" | jq
```

Resposta esperada:
```json
[
  {
    "id":"<uuid>",
    "filename":"sample1.mp4",
    "status":"DONE",
    "created_at":"...",
    "updated_at":"..."
  }
]
```

Status possíveis: `PENDING` → `PROCESSING` → `DONE` | `ERROR`

### 2.5 Download do arquivo processado (ZIP com frames)

```bash
# Substitua <video_id> pelo UUID retornado no upload
curl -s http://localhost:8080/videos/<video_id>/download \
  -H "Authorization: Bearer $TOKEN" \
  -o resultado.zip

unzip -l resultado.zip
```

---

## 3. Testes com os 3 vídeos de exemplo

```bash
# Faça login e exporte o token (veja seção 2.2)

# Upload simultâneo dos 3 vídeos
for i in 1 2 3; do
  echo "Enviando sample${i}.mp4..."
  curl -s -X POST http://localhost:8080/videos \
    -H "Authorization: Bearer $TOKEN" \
    -F "video=@samples/sample${i}.mp4" | jq '.video_id'
done

# Aguarde ~10s e verifique todos os status
curl -s http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" | jq '.[].status'
```

---

## 4. Testes de autenticação e segurança

### Token ausente deve retornar 401

```bash
curl -s -o /dev/null -w "%{http_code}" \
  http://localhost:8080/videos
# Esperado: 401
```

### Token inválido deve retornar 401

```bash
curl -s -o /dev/null -w "%{http_code}" \
  -H "Authorization: Bearer token_invalido" \
  http://localhost:8080/videos
# Esperado: 401
```

### Rate limit: mais de 100 req/s por IP retorna 429

```bash
# Requer hey ou apache bench instalado
hey -n 200 -c 50 http://localhost:8080/health
```

---

## 5. Testes unitários (sem Docker)

### Rodar todos os módulos de uma vez

```bash
./scripts/coverage.sh
```

### Rodar por módulo individualmente

```bash
cd api-gateway    && go test ./... && cd ..
cd auth-service   && go test ./... && cd ..
cd upload-service && go test ./... && cd ..
cd status-service && go test ./... && cd ..
cd notification-service && go test ./... && cd ..
cd worker         && go test ./... && cd ..
```

### Com flag de cobertura detalhada

```bash
cd worker
go test -coverprofile=coverage.out ./config/... ./pipeline/... ./notification/...
go tool cover -html=coverage.out -o coverage.html
open coverage.html   # macOS
```

---

## 6. Observabilidade

### Prometheus

Acesse http://localhost:9090 e execute queries como:

```promql
# Taxa de requisições por serviço
rate(http_requests_total[1m])

# Latência P99
histogram_quantile(0.99, rate(http_request_duration_seconds_bucket[5m]))
```

### Grafana

1. Acesse http://localhost:3000 (admin/admin)
2. Vá em **Connections → Data Sources → Add data source**
3. Selecione **Prometheus** → URL: `http://prometheus:9090` → **Save & Test**
4. Importe um dashboard na tela **Dashboards → Import** (ex: ID `1860` do grafana.com para Node Exporter)

### RabbitMQ

Monitore filas e mensagens mortas em http://localhost:15672:
- Fila `video.upload` — mensagens pendentes
- Fila `video.upload.dlq` — mensagens com falha após max_retries
- Fila `notification` — notificações pendentes

---

## 7. Parando e limpando

```bash
# Parar todos os serviços (preserva volumes)
docker compose down

# Parar e apagar todos os dados (volumes incluídos)
docker compose down -v

# Rebuild forçado de um serviço específico
docker compose up --build api-gateway
```

---

## 8. Variáveis de ambiente relevantes

| Variável | Padrão | Descrição |
|---|---|---|
| `JWT_SECRET` | — | Segredo HS256 (mín. 32 chars) |
| `JWT_EXPIRATION_HOURS` | `24` | Validade do token |
| `GATEWAY_PORT` | `8080` | Porta do API Gateway |
| `RATE_LIMIT_RPS` | `100` | Req/s por IP |
| `QUEUE_NAME` | `video.upload` | Fila principal de vídeos |
| `NOTIFICATION_QUEUE` | `notification` | Fila de notificações |
| `PREFETCH_COUNT` | `5` | Prefetch do consumer RabbitMQ |
| `MAX_RETRIES` | `3` | Tentativas antes do DLQ |
| `SMTP_HOST` | `""` | Vazio = modo log (sem envio real) |
| `MINIO_BUCKET` | `videos` | Bucket de armazenamento |

---

## 9. Troubleshooting

**`api-gateway` recusa conexão**
```bash
docker compose logs api-gateway
# Verifique se auth-service, upload-service e status-service subiram
```

**Worker não processa mensagens**
```bash
docker compose logs worker
# Verifique RABBITMQ_URL e MINIO_ENDPOINT no .env
```

**Erro de permissão no MinIO**
```bash
docker compose logs minio-setup
# O bucket deve ser criado automaticamente; rode: docker compose restart minio-setup
```

**Banco não inicializa**
```bash
docker compose logs postgres
# O script scripts/init.sql é executado apenas na primeira vez
# Para reiniciar do zero: docker compose down -v && docker compose up
```
