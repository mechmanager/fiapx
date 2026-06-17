# ---- Estágio de build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /notification-service ./cmd/notification

# ---- Estágio final ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates && addgroup -S appgroup && adduser -S -G appgroup appuser

WORKDIR /app
COPY --from=builder /notification-service /app/notification-service

USER appuser
ENTRYPOINT ["/app/notification-service"]
