# FIAP X — Como executar localmente

> Guia completo para rodar o sistema de processamento de vídeos em ambiente local com Docker Compose.

---

## Pré-requisitos

| Ferramenta | Versão mínima |
|------------|--------------|
| Docker Desktop | 24+ |
| Docker Compose | v2 (incluído no Docker Desktop) |

**Portas que precisam estar livres:**

| Porta | Serviço |
|-------|---------|
| 8080 | API Gateway (frontend + API) |
| 5432 | PostgreSQL |
| 5672 / 15672 | RabbitMQ (AMQP / Management UI) |
| 9000 / 9001 | MinIO (API / Console) |
| 6379 | Redis |
| 9090 | Prometheus |
| 3000 | Grafana |

---

## 1. Configuração do ambiente

O arquivo `.env` já está presente na raiz com todas as variáveis preenchidas (inclusive SMTP).  
Edite-o se precisar trocar senhas ou configurar outro servidor de e-mail.

```bash
# Variáveis principais — valores padrão já funcionam localmente
POSTGRES_USER=fiapx
POSTGRES_PASSWORD=fiapx123
JWT_SECRET=fiapx-super-secret-jwt-key-minimum-32-chars-ok
MINIO_ROOT_USER=minioadmin
MINIO_ROOT_PASSWORD=minioadmin123
SMTP_HOST=smtp.gmail.com
```

---

## 2. Subir todos os serviços

```bash
docker compose up --build
```

Na **primeira execução** o Docker irá:
1. Fazer o build das 6 imagens Go
2. Subir PostgreSQL, RabbitMQ, MinIO e Redis
3. Criar o bucket `videos` no MinIO automaticamente
4. Rodar `scripts/init.sql` criando as tabelas `users` e `videos`
5. Iniciar api-gateway, auth-service, upload-service, status-service, worker e notification-service

Aguarde todos os containers ficarem `healthy` (leva ~30–60 s na primeira vez).

Para subir em background:
```bash
docker compose up --build -d
```

---

## 3. Verificar que tudo subiu

```bash
docker compose ps
```

Todos os serviços devem aparecer com status `Up` ou `healthy`.

---

## 4. Acessar o frontend

```
http://localhost:8080
```

O frontend React carrega direto no browser com design FIAP PosTech.

**Funcionalidades:**
- Cadastro e login de usuário
- Upload de um ou múltiplos vídeos por vez (drag-and-drop ou seleção no finder)
- Atualização automática do status a cada 4 s enquanto houver vídeos processando
- Download do ZIP de frames com autenticação JWT embutida
- Painel de estatísticas: total, concluídos, processando e frames extraídos

---

## 5. Fluxo de uso via API (curl)

### 5.1 Criar conta

```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Seu Nome","email":"voce@email.com","password":"senha123"}' | jq
```

### 5.2 Fazer login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"voce@email.com","password":"senha123"}' | jq -r '.token')
```

### 5.3 Enviar um vídeo

```bash
curl -s -X POST http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" \
  -F "video=@/caminho/para/video.mp4" | jq
# {"id":"<uuid>","status":"PENDING"}
```

### 5.4 Consultar status

```bash
curl -s http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" | jq
```

O status evolui: `PENDING` → `PROCESSING` → `DONE` (ou `ERROR`).

### 5.5 Baixar o ZIP de frames

```bash
curl -OJ http://localhost:8080/videos/<uuid>/download \
  -H "Authorization: Bearer $TOKEN"
```

### 5.6 Processar múltiplos vídeos em paralelo

```bash
for video in samples/sample1.mp4 samples/sample2.mp4 samples/sample3.mp4; do
  curl -s -X POST http://localhost:8080/videos \
    -H "Authorization: Bearer $TOKEN" \
    -F "video=@$video" &
done
wait
curl -s http://localhost:8080/videos -H "Authorization: Bearer $TOKEN" | jq '.[].status'
```

---

## 6. Observabilidade

### Prometheus
```
http://localhost:9090
```
Coleta métricas de todos os serviços a cada 15 s.

### Grafana
```
http://localhost:3000
```

### Kibana (logs)
```
http://localhost:5601
```

Acesso anônimo — nenhum login necessário.

Na primeira vez, crie o **index pattern** para visualizar os logs:
1. Acesse **Management → Stack Management → Index Patterns**
2. Crie um index pattern com o valor `fiapx-logs-*`
3. Selecione `@timestamp` como campo de tempo
4. Acesse **Discover** para explorar os logs em tempo real

Você pode filtrar por serviço usando o campo `service` (ex.: `service: worker` mostra só os logs do worker).
Login: qualquer usuário (acesso anônimo habilitado) ou `admin` / `admin`.

O dashboard **"FIAP X — Monitoramento"** já aparece automaticamente na pasta *FIAP X* com 10 painéis:

| Painel | O que mostra |
|--------|-------------|
| Requests/s — API Gateway | Taxa de requisições por rota |
| Taxa de Erros 5xx | Percentual de erros no gateway |
| Latência p50/p95/p99 | Histograma de tempo de resposta |
| Fila RabbitMQ | Mensagens aguardando processamento |
| Vídeos Concluídos / com Erro | Contadores do worker |
| Throughput do Worker | Vídeos processados por minuto |
| Duração de Processamento | p50/p95 do tempo de extração de frames |
| Goroutines por Serviço | Saúde do runtime Go em cada serviço |

---

## 7. UIs de administração

| Serviço | URL | Credenciais |
|---------|-----|-------------|
| RabbitMQ Management | http://localhost:15672 | `fiapx` / `fiapx123` |
| MinIO Console | http://localhost:9001 | `minioadmin` / `minioadmin123` |
| Grafana | http://localhost:3000 | anônimo ou `admin` / `admin` |
| Prometheus | http://localhost:9090 | — |
| Kibana (logs) | http://localhost:5601 | — |

---

## 8. Escalar workers

Para processar mais vídeos em paralelo (ex.: 3 workers simultâneos):

```bash
docker compose up --scale worker=3 -d
```

Cada worker consome da mesma fila `video.upload` com prefetch 5.

---

## 9. Ver logs de um serviço

```bash
docker compose logs -f worker
docker compose logs -f api-gateway
docker compose logs -f notification-service
```

---

## 10. Parar tudo

```bash
# Mantém os dados (banco, MinIO, Grafana)
docker compose down

# Reset completo — apaga todos os volumes
docker compose down -v
```

---

## Vídeos de exemplo

A pasta `samples/` contém 4 vídeos prontos para teste (clips do Big Buck Bunny — CC BY 3.0):

| Arquivo | Cena |
|---------|------|
| `sample1.mp4` | Floresta — cena tranquila |
| `sample2.mp4` | Perseguição com a águia |
| `sample3.mp4` | Esquilo correndo |
| `sample4.mp4` | Borboletas — cena final |

---

## Referência de portas

| Serviço | Porta |
|---------|-------|
| API Gateway + Frontend | 8080 |
| Auth Service | 8081 (interno) |
| Upload Service | 8082 (interno) |
| Status Service | 8083 (interno) |
| Worker Metrics | 9100 (interno) |
| PostgreSQL | 5432 |
| RabbitMQ AMQP | 5672 |
| RabbitMQ Management | 15672 |
| MinIO API | 9000 |
| MinIO Console | 9001 |
| Redis | 6379 |
| Prometheus | 9090 |
| Grafana | 3000 |
| Kibana | 5601 |
| Elasticsearch | 9200 (interno) |
