# FIAP X — Como executar localmente

## Pré-requisitos

- [Docker](https://docs.docker.com/get-docker/) + Docker Compose v2
- Porta `8080` (gateway), `5432` (postgres), `5672`/`15672` (rabbitmq), `9000`/`9001` (minio), `6379` (redis) livres

---

## 1. Configuração do ambiente

O arquivo `.env` já está pronto na raiz do projeto com todas as variáveis configuradas (incluindo SMTP).

Caso precise ajustar algo (ex.: trocar senha do banco), edite o `.env` antes de subir.

---

## 2. Subir todos os serviços

```bash
docker compose up --build
```

Na primeira execução o Docker vai:
1. Fazer o build das imagens de cada serviço Go
2. Subir PostgreSQL, RabbitMQ, MinIO e Redis
3. Criar o bucket `videos` no MinIO automaticamente
4. Executar o `scripts/init.sql` criando as tabelas `users` e `videos`
5. Subir api-gateway, auth-service, upload-service, status-service, worker e notification-service

Aguarde todos os containers ficarem `healthy` antes de usar (leva ~30–60s na primeira vez).

---

## 3. Verificar que tudo subiu

```bash
docker compose ps
```

Todos os serviços devem estar com status `running` ou `Up`.

---

## 4. Acessar o frontend

Abra no navegador:

```
http://localhost:8080
```

A página de login do FIAP X vai aparecer.

---

## 5. Fluxo de uso

### 5.1 Criar conta

```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Teste","email":"teste@fiapx.com","password":"senha123"}' | jq
```

### 5.2 Fazer login

```bash
curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"teste@fiapx.com","password":"senha123"}' | jq
```

Copie o `token` retornado.

### 5.3 Enviar um vídeo

```bash
curl -s -X POST http://localhost:8080/videos \
  -H "Authorization: Bearer SEU_TOKEN_AQUI" \
  -F "video=@/caminho/para/video.mp4" | jq
```

### 5.4 Consultar status do vídeo

```bash
curl -s http://localhost:8080/videos \
  -H "Authorization: Bearer SEU_TOKEN_AQUI" | jq
```

O status vai evoluir: `PENDING` → `PROCESSING` → `DONE` (ou `ERROR`).

Quando `DONE`, o campo `zip_url` contém o link para baixar o ZIP com os frames extraídos.

---

## 6. UIs de administração

| Serviço    | URL                          | Usuário / Senha          |
|------------|------------------------------|--------------------------|
| RabbitMQ   | http://localhost:15672       | `fiapx` / `fiapx123`     |
| MinIO      | http://localhost:9001        | `minioadmin` / `minioadmin123` |

---

## 7. Escalar workers

Para processar mais vídeos em paralelo:

```bash
docker compose up --scale worker=3
```

---

## 8. Ver logs de um serviço específico

```bash
docker compose logs -f worker
docker compose logs -f notification-service
docker compose logs -f api-gateway
```

---

## 9. Parar tudo

```bash
# Para os containers mas mantém os dados
docker compose down

# Para e apaga volumes (banco, minio, grafana) — reset completo
docker compose down -v
```

---

## Portas resumidas

| Serviço             | Porta |
|---------------------|-------|
| API Gateway         | 8080  |
| Auth Service        | 8081  |
| Upload Service      | 8082  |
| Status Service      | 8083  |
| PostgreSQL          | 5432  |
| RabbitMQ AMQP       | 5672  |
| RabbitMQ Management | 15672 |
| MinIO API           | 9000  |
| MinIO Console       | 9001  |
| Redis               | 6379  |
| Prometheus          | 9090  |
| Grafana             | 3000  |
