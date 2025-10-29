# Multi-stage build for Yummy Recipe Manager
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o yummy .

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user
RUN addgroup -g 1001 -S yummy && \
    adduser -u 1001 -S yummy -G yummy

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/yummy .

# Copy assets
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/readme.md .
COPY --from=builder /app/LICENSE .

# Change ownership
RUN chown -R yummy:yummy /app

# Switch to non-root user
USER yummy

# Expose port (if needed for future web features)
EXPOSE 8080

# Set environment variables
ENV GIN_MODE=release

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD ./yummy --version || exit 1

# Default command
ENTRYPOINT ["./yummy"]
