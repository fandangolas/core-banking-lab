# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o perf-test ./cmd/loadtest

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl

WORKDIR /app

# Copy the binary from builder stage
COPY --from=builder /app/perf-test .

# Copy web assets
COPY web/ ./web/

# Create reports directory
RUN mkdir -p reports

# Expose port for web UI
EXPOSE 9999

# Default command runs the web server
CMD ["./perf-test", "-mode=server", "-server-port=9999", "-api-url=http://banking-api-service:8080"]