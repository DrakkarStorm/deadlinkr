# ==============================================================================
# Build Stage
# ==============================================================================
FROM golang:1.24.6-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files first (better layer caching)
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download && go mod verify

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o deadlinkr \
    .

# ==============================================================================
# Final Stage - Distroless for security and minimal size
# ==============================================================================
FROM gcr.io/distroless/static:nonroot

# Copy timezone data and ca-certificates from builder
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

# Copy the binary
COPY --from=builder /app/deadlinkr /usr/local/bin/deadlinkr

# Create directories for configuration and logs
USER nonroot:nonroot

# Set environment variables
ENV LOG_LEVEL=info
ENV TIMEOUT=10
ENV CONCURRENCY=20
ENV RATE_LIMIT=2.0
ENV CACHE_ENABLED=true

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ["/usr/local/bin/deadlinkr", "--help"]

# Default command
ENTRYPOINT ["/usr/local/bin/deadlinkr"]
CMD ["--help"]

# Metadata
LABEL maintainer="DrakkarStorm <tavernetech@gmail.com>"
LABEL description="Deadlinkr - Advanced dead link checker with authentication support"
LABEL version="1.0.0"
LABEL org.opencontainers.image.source="https://github.com/DrakkarStorm/deadlinkr"
LABEL org.opencontainers.image.documentation="https://github.com/DrakkarStorm/deadlinkr/blob/master/README.md"
LABEL org.opencontainers.image.licenses="MIT"