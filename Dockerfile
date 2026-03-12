# ============================================
# Stage 1: Build Go Backend
# ============================================
FROM golang:1.22-alpine AS backend-builder

WORKDIR /app/backend

RUN apk add --no-cache gcc musl-dev

COPY backend/go.mod backend/go.sum ./
RUN go mod download

COPY backend/ .
RUN CGO_ENABLED=1 CGO_CFLAGS="-D_LARGEFILE64_SOURCE" GOOS=linux go build -o server .

# ============================================
# Stage 2: Build Vue Frontend
# ============================================
FROM node:20-alpine AS frontend-builder

WORKDIR /app/frontend

COPY frontend/package*.json ./
RUN npm ci

COPY frontend/ .
ENV VITE_API_URL=/api
RUN npm run build

# ============================================
# Stage 3: Final Runtime (Go serves everything)
# ============================================
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata sqlite-libs

WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/backend/server .

# Copy frontend build - Go will serve these as static files
COPY --from=frontend-builder /app/frontend/dist ./static

# Create directories
RUN mkdir -p /app/data /app/uploads

# Environment
ENV GIN_MODE=release
ENV DB_PATH=/app/data/eval.db
ENV UPLOAD_PATH=/app/uploads
ENV STATIC_DIR=/app/static
ENV JWT_SECRET=change-this-in-production
ENV PORT=8080

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget -qO- http://localhost:8080/api/health || exit 1

CMD ["./server"]
