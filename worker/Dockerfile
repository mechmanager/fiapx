# ---- Estágio de build ----
# Compila um binário estático a partir do código-fonte.
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copia os manifests primeiro para aproveitar o cache de camadas do Docker.
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY config/ ./config/
COPY consumer/ ./consumer/
COPY domain/ ./domain/
COPY metrics/ ./metrics/
COPY notification/ ./notification/
COPY pipeline/ ./pipeline/
COPY processor/ ./processor/
COPY repository/ ./repository/
COPY storage/ ./storage/
RUN CGO_ENABLED=0 GOOS=linux go build -o /worker ./cmd/worker

# ---- Estágio final ----
# Imagem enxuta com o binário e o ffmpeg (necessário para extrair frames).
FROM alpine:3.20

# ffmpeg processa os vídeos; ca-certificates habilita chamadas TLS.
RUN apk add --no-cache ffmpeg ca-certificates && addgroup -S appgroup && adduser -S -G appgroup appuser

WORKDIR /app
COPY --from=builder /worker /app/worker

USER appuser
ENTRYPOINT ["/app/worker"]
