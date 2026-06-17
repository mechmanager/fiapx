# ---- Estágio de build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY cache/ ./cache/
COPY cmd/ ./cmd/
COPY config/ ./config/
COPY domain/ ./domain/
COPY handler/ ./handler/
COPY repository/ ./repository/
COPY service/ ./service/
COPY storage/ ./storage/
RUN CGO_ENABLED=0 GOOS=linux go build -o /status-service ./cmd/status

# ---- Estágio final ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && addgroup -S appgroup && adduser -S -G appgroup appuser

WORKDIR /app
COPY --from=builder /status-service /app/status-service

USER appuser
EXPOSE 8083

ENTRYPOINT ["/app/status-service"]
