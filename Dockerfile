# Build stage
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git protobuf-dev

# Install buf
RUN wget -O /usr/local/bin/buf https://github.com/bufbuild/buf/releases/download/v1.28.1/buf-Linux-x86_64 && \
    chmod +x /usr/local/bin/buf

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Generate protobuf files
RUN buf generate

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o bin/nettestlab cmd/server/main.go

# Runtime stage
FROM alpine:3.18

# Install runtime dependencies
RUN apk add --no-cache \
    iproute2 \
    iptables \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1000 nettestlab && \
    adduser -D -u 1000 -G nettestlab nettestlab

# Copy binary from builder
COPY --from=builder /app/bin/nettestlab /usr/local/bin/nettestlab

# Create directories
RUN mkdir -p /etc/nettestlab /var/log/nettestlab && \
    chown -R nettestlab:nettestlab /etc/nettestlab /var/log/nettestlab

# Switch to non-root user
USER nettestlab

# Expose gRPC port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD grpc_health_probe -addr=localhost:8080 || exit 1

# Set entrypoint
ENTRYPOINT ["/usr/local/bin/nettestlab"]
CMD ["-host", "0.0.0.0", "-port", "8080"]