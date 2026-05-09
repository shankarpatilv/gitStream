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
    -o api ./cmd/api

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS connections to databases
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S gitstream && \
    adduser -u 1001 -S gitstream -G gitstream

# Set working directory
WORKDIR /home/gitstream

# Copy binary from builder stage
COPY --from=builder /app/api .

# Change ownership to non-root user
RUN chown gitstream:gitstream ./api

# Switch to non-root user
USER gitstream

# Expose port (configurable via API_PORT env var, default 8090)
EXPOSE 8090

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8090/health || exit 1

# Run the binary
CMD ["./api"]