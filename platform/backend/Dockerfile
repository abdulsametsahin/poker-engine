# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod files first for better layer caching
COPY platform/backend/go.mod platform/backend/go.sum ./platform/backend/

# Change to backend directory and download dependencies
WORKDIR /app/platform/backend
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

# Copy poker-engine module files (from root)
COPY go.mod go.sum ./
COPY engine ./engine
COPY models ./models
COPY server ./server

# Copy backend application
COPY platform/backend ./platform/backend

# Build the application with cache mounts
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o poker-server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata mysql-client bash

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/platform/backend/poker-server .

# Copy migrations directory
COPY --from=builder /app/platform/backend/migrations ./migrations

# Copy database scripts (optional, for reference)
COPY --from=builder /app/platform/backend/scripts ./scripts

# Copy entrypoint script
COPY --from=builder /app/platform/backend/entrypoint.sh .
RUN chmod +x entrypoint.sh

# Expose port
EXPOSE 8080

# Set entrypoint to run migrations before starting the app
ENTRYPOINT ["./entrypoint.sh"]
