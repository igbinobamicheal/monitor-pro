FROM golang:alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Cache dependencies
COPY services/api/go.mod services/api/go.sum ./
RUN go mod download

# Copy source
COPY services/api/ ./

# Build statically linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/monitopro-api ./cmd/server

# Final minimal runtime image
FROM alpine:3.20

RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app
COPY --from=builder /bin/monitopro-api /app/monitopro-api

EXPOSE 8080

ENTRYPOINT ["/app/monitopro-api"]
