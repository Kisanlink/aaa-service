# Multi-stage build for AAA Service
# Stage 1: Builder
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata make gcc musl-dev

# Set working directory
WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./

# Download dependencies (cached if go.mod/go.sum unchanged)
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o aaa-service \
    ./cmd/server/main.go

# Verify binary was created
RUN test -f /app/aaa-service

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies and security updates
RUN apk --no-cache add ca-certificates tzdata curl && \
    apk upgrade --no-cache

# Create non-root user and group
RUN addgroup -g 1001 -S aaa && \
    adduser -u 1001 -S aaa -G aaa

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder --chown=aaa:aaa /app/aaa-service .

# Copy additional required files
COPY --from=builder --chown=aaa:aaa /app/migrations ./migrations
COPY --from=builder --chown=aaa:aaa /app/docs ./docs

# Ensure binary is executable
RUN chmod +x ./aaa-service

# Switch to non-root user
USER aaa

# Expose HTTP and gRPC ports
EXPOSE 8080 50051

# Health check for HTTP endpoint
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/health || exit 1

# Set environment variables with defaults
ENV GIN_MODE=release \
    LOG_LEVEL=info \
    PORT=8080 \
    GRPC_PORT=50051 \
    APP_ENV=production

# Run the application
ENTRYPOINT ["./aaa-service"]
