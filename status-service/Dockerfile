# ---- Estágio de build ----
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /status-service ./cmd/status

# ---- Estágio final ----
FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /app
COPY --from=builder /status-service /app/status-service

EXPOSE 8083

ENTRYPOINT ["/app/status-service"]
