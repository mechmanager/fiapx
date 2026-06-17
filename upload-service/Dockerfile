# Stage 1: build
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o upload-service ./cmd/upload

# Stage 2: final image
FROM alpine:3.20

RUN apk --no-cache add ca-certificates tzdata && addgroup -S appgroup && adduser -S -G appgroup appuser

WORKDIR /app

COPY --from=builder /app/upload-service .

USER appuser
EXPOSE 8082

ENTRYPOINT ["./upload-service"]
