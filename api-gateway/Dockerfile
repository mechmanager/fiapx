FROM golang:1.25-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY cmd/ ./cmd/
COPY config/ ./config/
COPY middleware/ ./middleware/
COPY proxy/ ./proxy/
COPY web/ ./web/
RUN CGO_ENABLED=0 GOOS=linux go build -o /api-gateway ./cmd/gateway

FROM alpine:3.20
RUN apk add --no-cache ca-certificates && addgroup -S appgroup && adduser -S -G appgroup appuser
WORKDIR /app
COPY --from=builder /api-gateway /app/api-gateway
USER appuser
EXPOSE 8080
ENTRYPOINT ["/app/api-gateway"]
