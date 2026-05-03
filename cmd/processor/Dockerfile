# Build stage
FROM golang:1.25-alpine AS builder

# Install git for dependency resolution
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the binary with static linking
# CGO_ENABLED=0 ensures static binary for alpine
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o processor ./cmd/processor

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS/TLS connections to Kafka and databases
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S gitstream && \
    adduser -u 1001 -S gitstream -G gitstream

# Set working directory
WORKDIR /home/gitstream

# Copy binary from builder stage
COPY --from=builder /app/processor .

# Change ownership to non-root user
RUN chown gitstream:gitstream ./processor

# Switch to non-root user
USER gitstream

# Expose metrics port (configurable via PROCESSOR_METRICS_PORT env var, default 8091)
EXPOSE 8091

# Health check via metrics endpoint
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8091/metrics || exit 1

# Run the binary
CMD ["./processor"]