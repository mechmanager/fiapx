# FIAP X — Plataforma de Processamento de Vídeos

Sistema de processamento de vídeos em arquitetura de microsserviços, construído como
reescrita escalável de um monolito Go. Recebe vídeos via API REST, extrai frames com
ffmpeg e disponibiliza um `.zip` para download.

## Arquitetura

Veja [ARCHITECTURE.md](ARCHITECTURE.md) para diagramas de componentes e sequência.

**Serviços:**
- **api-gateway** — autenticação JWT, upload, listagem e download (Gin)
- **worker** — consome fila RabbitMQ, processa vídeo com ffmpeg, armazena zip no MinIO
- **PostgreSQL** — usuários e metadados de vídeo
- **RabbitMQ** — fila durável `video.process` (prefetch=5, ack manual)
- **MinIO** — armazenamento S3-compatível de vídeos e zips

## Pré-requisitos

- Docker e Docker Compose v2
- `ffmpeg` instalado nos contêineres (incluído no Dockerfile do worker)
- (Opcional) `go 1.25+` para rodar testes locais

## Configuração

```bash
# 1. Copiar e editar variáveis de ambiente
cp .env.example .env
# Edite .env e troque todos os valores TROQUE_*

# 2. Subir todos os serviços
docker compose up --build
```

A aplicação estará disponível em `http://localhost:8080`.

## Interface Web

Acesse `http://localhost:8080` para usar o frontend HTML integrado:
- Cadastro e login de usuário
- Upload de vídeo (`.mp4`, `.avi`, `.mov`, `.mkv`, `.webm`)
- Listagem de status em tempo real
- Download do zip de frames

## API — exemplos com curl

### Saúde

```bash
curl http://localhost:8080/health
# {"status":"ok"}
```

### Cadastro

```bash
curl -s -X POST http://localhost:8080/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"usuario@exemplo.com","password":"senha123"}' | jq
```

### Login

```bash
TOKEN=$(curl -s -X POST http://localhost:8080/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"usuario@exemplo.com","password":"senha123"}' \
  | jq -r '.token')
echo "Token: $TOKEN"
```

### Upload de vídeo

```bash
curl -s -X POST http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" \
  -F "file=@/caminho/para/video.mp4" | jq
# {"id":"<uuid>","status":"PENDING"}
```

### Listagem de vídeos

```bash
curl -s http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" | jq
# [{"id":"...","filename":"video.mp4","status":"DONE","frame_count":42,...}]
```

### Download do zip de frames

```bash
curl -OJ http://localhost:8080/videos/<uuid>/download \
  -H "Authorization: Bearer $TOKEN"
# Salva frames_<uuid>.zip no diretório atual
```

### Teste de processamento paralelo

```bash
# Enviar 3 vídeos simultaneamente
for i in 1 2 3; do
  curl -s -X POST http://localhost:8080/videos \
    -H "Authorization: Bearer $TOKEN" \
    -F "file=@video.mp4" &
done
wait

# Acompanhar status
curl -s http://localhost:8080/videos \
  -H "Authorization: Bearer $TOKEN" | jq '.[].status'
```

## Testes

```bash
# Rodar testes com gate de cobertura ≥ 80% (requer Go instalado localmente)
bash scripts/coverage.sh

# Ou por módulo
cd api-gateway && go test ./... -cover
cd worker      && go test ./... -cover
```

## Kubernetes

```bash
# 1. Criar secrets a partir do template
cp k8s/secrets.yaml.example k8s/secrets.yaml
# Edite k8s/secrets.yaml com valores base64 reais:
#   echo -n "minhasenha" | base64

# 2. Aplicar todos os manifestos
kubectl apply -k k8s/

# 3. Acompanhar pods
kubectl get pods -n fiapx -w
```

## Infraestrutura AWS (Terraform)

```bash
cd terraform

# Inicializar providers
terraform init

# Planejar (revise antes de aplicar)
terraform plan

# Criar cluster EKS (~15 min)
terraform apply

# Configurar kubectl
$(terraform output -raw kubeconfig_command)

# Destruir quando não precisar mais
terraform destroy
```

> **Custo estimado:** t3.medium × 3 nós ≈ US$ 0,12/h. Lembre de destruir o cluster
> após a apresentação para evitar cobranças.

## CI/CD

O pipeline `.github/workflows/ci.yml` executa em todo push e PR:

| Job | O que faz |
|-----|-----------|
| `lint` | `gofmt` + `go vet` nos dois módulos |
| `test` | `go test` com cobertura, gate ≥ 80%, artefatos de coverage |
| `sonar` | SonarCloud scan (requer secret `SONAR_TOKEN` no repositório) |
| `docker` | Build e push das imagens para `ghcr.io` (apenas na branch `main`) |

### Secrets necessários no GitHub

| Secret | Descrição |
|--------|-----------|
| `SONAR_TOKEN` | Token de autenticação do SonarCloud |
| `GITHUB_TOKEN` | Automático — usado para push no GHCR |

## Variáveis de ambiente

Veja `.env.example` para a lista completa. Nunca comite o arquivo `.env`.

## Estado dos vídeos

```
PENDING → PROCESSING → DONE
                    ↘ ERROR
```

- `PENDING`: upload recebido, na fila aguardando worker
- `PROCESSING`: worker iniciou o processamento
- `DONE`: zip gerado e disponível para download
- `ERROR`: falha no processamento (detalhes no log do worker)
