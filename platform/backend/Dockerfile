# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy the entire project (since we need the poker-engine from root)
COPY . .

# Change to backend directory
WORKDIR /app/platform/backend

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o poker-server ./cmd/server

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
