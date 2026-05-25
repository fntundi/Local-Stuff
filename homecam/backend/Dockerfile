# Sentinel NOC - Go Backend Dockerfile
# =====================================
# Multi-stage build for optimized production image
#
# Build:
#   docker build -t sentinel-backend ./backend
#
# Run:
#   docker run -p 8001:8001 sentinel-backend

# Stage 1: Build the application
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# go.sum is optional: copied when present, generated when absent.
# Once go.sum is committed, the build is fully reproducible.
COPY go.mod go.su[m] ./

# Download dependencies (creates go.sum if missing)
RUN go mod download

# Copy source code
COPY . .

# Build the application with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/sentinel-noc \
    ./cmd/api

# Stage 2: Create minimal runtime image
FROM alpine:3.19

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata curl

# Create non-root user for security
RUN adduser -D -g '' appuser

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/sentinel-noc .

# Copy timezone data
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo

# Set ownership
RUN chown -R appuser:appuser /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8001

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8001/api/health || exit 1

# Run the application
CMD ["./sentinel-noc"]
