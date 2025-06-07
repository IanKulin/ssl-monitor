# Build stage
FROM golang:1.24-alpine AS builder

# Install ca-certificates for SSL/TLS connections during build
RUN apk --no-cache add ca-certificates

WORKDIR /app

# Copy go mod file for better caching
COPY go.mod ./
RUN go mod download

# Copy source code
COPY src/*.go ./

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ssl-monitor .

# Final stage
FROM alpine:latest

# Install ca-certificates for SSL/TLS connections to external sites
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/ssl-monitor .

# Create data directory
RUN mkdir data

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/ || exit 1

# Run the binary
CMD ["./ssl-monitor"]