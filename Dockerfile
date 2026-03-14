# Build stage
FROM golang:1.18-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o driftguard ./cmd/main.go

# Final stage
FROM alpine:latest

WORKDIR /app

# Install ca-certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1000 driftguard && \
    adduser -D -u 1000 -G driftguard driftguard

# Copy binary from builder
COPY --from=builder /app/driftguard .
COPY --from=builder /app/config.example.yaml ./config.yaml

# Set ownership
RUN chown -R driftguard:driftguard /app

USER driftguard

# Expose API port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["./driftguard"]
CMD ["--config", "./config.yaml"]
