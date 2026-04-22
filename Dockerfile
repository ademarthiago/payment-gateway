# =============================================================================
# Stage 1: Builder
# =============================================================================
FROM golang:1.24-bookworm AS builder

WORKDIR /app

# Install dependencies first (layer cache optimization)
COPY go.mod go.sum ./
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build binary
# CGO_ENABLED=0 for static binary
# -ldflags reduces binary size
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.version=$(git describe --tags --always --dirty 2>/dev/null || echo dev)" \
    -o /app/bin/api \
    ./cmd/api

# =============================================================================
# Stage 2: Runtime
# =============================================================================
FROM debian:bookworm-slim AS runtime

# Security: non-root user
RUN groupadd --gid 1001 appgroup && \
    useradd --uid 1001 --gid appgroup --shell /bin/bash --create-home appuser

# Install ca-certificates for HTTPS calls and wget for healthcheck
RUN apt-get update && apt-get install -y \
    ca-certificates \
    wget \
    --no-install-recommends && \
    rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/bin/api .

# Copy migrations (needed at runtime)
COPY --from=builder /app/migrations ./migrations

# Ownership
RUN chown -R appuser:appgroup /app

USER appuser

EXPOSE 8088

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8088/health || exit 1

ENTRYPOINT ["./api"]