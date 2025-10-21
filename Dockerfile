# Build stage
FROM golang:1.23-alpine AS builder

WORKDIR /app

# Install build dependencies
# - gcc, musl-dev: Required for CGO
# - sqlite-dev: SQLite development headers
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Copy go mod files and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
# Use libsqlite3 tag to link against system SQLite instead of bundled version
RUN CGO_ENABLED=1 GOOS=linux go build \
    -a \
    -tags "sqlite_foreign_keys libsqlite3" \
    -ldflags="-s -w" \
    -trimpath \
    -o pocketjson \
    ./cmd/pocketjson

# Final stage
FROM alpine:latest

WORKDIR /app

# Install only runtime dependencies
# - sqlite-libs: SQLite shared library (required at runtime)
# - ca-certificates: For HTTPS support
# - curl: For healthcheck
RUN apk add --no-cache sqlite-libs ca-certificates curl && \
    # Remove apk cache to save space
    rm -rf /var/cache/apk/*

# Copy the binary from builder
COPY --from=builder /app/pocketjson .

# Copy templates directory for HTML rendering
COPY --from=builder /app/templates ./templates

# Create data directory with proper permissions
RUN mkdir -p /app/data && \
    chown -R nobody:nobody /app/data

# Use non-root user for security
USER nobody

# Expose port
EXPOSE 9819

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD curl -f http://localhost:9819/health || exit 1

# Command to run the application
CMD ["./pocketjson"]
