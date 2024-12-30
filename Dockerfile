# Build stage
FROM golang:1.23-bullseye AS builder

WORKDIR /app

# Install build dependencies
RUN apt-get update && apt-get install -y \
    gcc \
    sqlite3 \
    libsqlite3-dev \
    && rm -rf /var/lib/apt/lists/*

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -tags sqlite_foreign_keys -o pocketjson .

# Final stage
FROM debian:bullseye-slim

WORKDIR /app

# Install runtime dependencies
RUN apt-get update && apt-get install -y \
    sqlite3 \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Copy the binary from builder
COPY --from=builder /app/pocketjson .

# Create data directory
RUN mkdir -p /app/data && \
    chown -R nobody:nogroup /app/data

# Use non-root user (Debian uses nogroup instead of nobody)
USER nobody

# Command to run the application
CMD ["./pocketjson"]