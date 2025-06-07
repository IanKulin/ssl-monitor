# Build stage
FROM golang:1.24-alpine AS builder

# Install ca-certificates for SSL/TLS connections during build
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy go mod file for better caching
COPY go.mod ./
RUN go mod download

# Copy source code
COPY *.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssl-monitor .

# Final stage
FROM alpine:latest

# Install ca-certificates for SSL/TLS connections to external sites
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Create a non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Copy the binary from builder stage
COPY --from=builder /app/ssl-monitor .

# Copy example settings as default configuration
COPY example-settings.json data/settings.json
COPY example-sites.json data/sites.json

# Create data directory with proper permissions
RUN mkdir -p data && \
    chown -R appuser:appgroup /app

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the binary
CMD ["./ssl-monitor"]