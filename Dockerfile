# Build stage
FROM golang:1.23.2-alpine AS builder

# Install git and ca-certificates (needed for go mod download)
RUN apk add --no-cache git ca-certificates

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o docs-mcp .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests to Google APIs
RUN apk --no-cache add ca-certificates

# Create a non-root user
RUN adduser -D -s /bin/sh appuser

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/docs-mcp .

# Change ownership to appuser
RUN chown appuser:appuser docs-mcp
USER appuser

# Expose port for HTTP mode (optional)
EXPOSE 8080

# Run the binary
ENTRYPOINT ["./docs-mcp"]
