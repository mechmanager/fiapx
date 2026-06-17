# ---- Estágio de build ----
# Compila um binário estático a partir do código-fonte.
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copia os manifests primeiro para aproveitar o cache de camadas do Docker.
COPY go.mod go.sum ./
RUN go mod download

# Copia o restante do código e compila o binário.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./cmd/api

# ---- Estágio final ----
# Imagem enxuta apenas com o binário e certificados TLS.
FROM alpine:3.20

# ca-certificates é necessário para chamadas TLS (ex.: MinIO/SMTP futuros).
RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /api-gateway /app/api-gateway

EXPOSE 8080

ENTRYPOINT ["/app/api-gateway"]
