# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install templ CLI
RUN go install github.com/a-h/templ/cmd/templ@latest

# Copy source code
COPY . .

# Build frontend templates
RUN templ generate

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o rab-maker .

# Runtime stage
FROM alpine:latest

# Install ca-certificates for HTTPS, timezone data, and su-exec for privilege dropping
RUN apk --no-cache add ca-certificates tzdata wget su-exec

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/rab-maker .

# Copy directories needed at runtime
COPY --from=builder /app/backend ./backend
COPY --from=builder /app/frontend ./frontend

# Create database directory and set permissions
RUN mkdir -p backend/databases && \
    chown -R appuser:appuser /app

# Copy entrypoint script
COPY --chown=appuser:appuser docker-entrypoint.sh /usr/local/bin/
RUN chmod +x /usr/local/bin/docker-entrypoint.sh

# Don't switch to appuser here - the entrypoint will do that after fixing permissions

# Environment variables
ENV PORT=3002
ENV ENV=production

EXPOSE 3002

ENTRYPOINT ["docker-entrypoint.sh"]
CMD ["./rab-maker"]
