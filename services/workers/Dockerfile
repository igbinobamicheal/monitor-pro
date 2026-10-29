# Multi-stage Dockerfile for MonitoPro Background Worker
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

COPY services/workers/go.mod services/workers/go.sum ./
RUN go mod download

COPY services/workers/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /worker ./cmd/worker/main.go

# Production runtime stage
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /worker /app/worker

USER nobody:nobody

ENTRYPOINT ["/app/worker"]
