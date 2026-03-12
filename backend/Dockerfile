# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache gcc musl-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=1 GOOS=linux go build -o server .

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata sqlite-libs

# Copy binary from builder
COPY --from=builder /app/server .

# Create uploads directory
RUN mkdir -p /app/uploads /app/data

# Set environment variables
ENV GIN_MODE=release
ENV DB_PATH=/app/data/eval.db
ENV UPLOAD_PATH=/app/uploads
ENV JWT_SECRET=change-this-in-production
ENV PORT=8080

EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["./server"]
