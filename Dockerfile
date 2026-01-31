# ==========================================
# Dockerfile for AAA Service
# Usage: docker build -t aaa-service .
# Context: Root of aaa-service folder
# ==========================================

# ---------------------------------
# Stage 1: Build
# ---------------------------------
FROM golang:1.24-alpine AS builder
WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git

# Copy dependencies first for caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/server cmd/server/main.go

# ---------------------------------
# Stage 2: Runtime
# ---------------------------------
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Copy binary from builder
COPY --from=builder /bin/server .

# Copy migration files if needed (assuming integrated or separate runner)
# COPY migrations ./migrations

# Expose ports
# 8080: HTTP
# 50051: gRPC
EXPOSE 8080 50051

# Healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD curl -f http://localhost:8080/health || exit 1

# Run the server
CMD ["./server"]