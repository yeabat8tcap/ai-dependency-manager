# Frontend build stage
FROM node:20-alpine AS frontend-builder

# Set working directory for frontend
WORKDIR /app/web

# Copy package files
COPY web/package*.json ./

# Install dependencies (including dev dependencies for build tools)
RUN npm ci

# Copy frontend source (excluding node_modules)
COPY web/src ./src
COPY web/public ./public
COPY web/angular.json ./
COPY web/tsconfig*.json ./
COPY web/scripts ./scripts

# Build frontend for production
RUN npm run build:production

# Backend build stage
FROM golang:1.21-alpine AS backend-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /app

# Set SQLite compilation environment variables for Alpine Linux
ENV CGO_CFLAGS="-D_LARGEFILE64_SOURCE"

# Copy Go module files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Copy built frontend assets from frontend-builder
COPY --from=frontend-builder /app/web/dist ./web/dist

# Build the application with embedded frontend (SQLite compatible flags for Alpine)
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -tags="sqlite_omit_load_extension" -o ai-dep-manager .

# Production runtime stage
FROM alpine:latest

# Install runtime dependencies for full-stack application
RUN apk --no-cache add \
    ca-certificates \
    git \
    curl \
    nodejs \
    npm \
    python3 \
    py3-pip \
    openjdk11 \
    maven \
    gradle \
    sqlite \
    tzdata \
    && rm -rf /var/cache/apk/*

# Create non-root user
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from backend-builder stage
COPY --from=backend-builder /app/ai-dep-manager .

# Copy configuration files
COPY --from=backend-builder /app/config.yaml.example ./config.yaml.example

# Create necessary directories with proper permissions
RUN mkdir -p /data /app/logs /app/db && \
    chown -R appuser:appgroup /data /app

# Switch to non-root user
USER appuser

# Expose ports for web server and API
EXPOSE 8080 8081

# Set environment variables for production deployment
ENV AI_DEP_MANAGER_DATA_DIR=/data
ENV AI_DEP_MANAGER_CONFIG_FILE=/data/config.yaml
ENV AI_DEP_MANAGER_DB_PATH=/data/ai-dep-manager.db
ENV AI_DEP_MANAGER_LOG_LEVEL=info
ENV AI_DEP_MANAGER_LOG_FORMAT=json
ENV AI_DEP_MANAGER_WEB_PORT=8080
ENV AI_DEP_MANAGER_API_PORT=8081
ENV GIN_MODE=release

# Health check for unified web server
HEALTHCHECK --interval=30s --timeout=10s --start-period=30s --retries=3 \
    CMD curl -f http://localhost:8080/api/health || exit 1

# Create startup script for unified deployment
RUN echo '#!/bin/sh' > /app/start.sh && \
    echo 'echo "🚀 Starting AI Dependency Manager - Production Deployment"' >> /app/start.sh && \
    echo 'echo "🌐 Web Server: http://localhost:8080"' >> /app/start.sh && \
    echo 'echo "📊 Frontend: http://localhost:8080"' >> /app/start.sh && \
    echo 'echo "🔗 API: http://localhost:8080/api"' >> /app/start.sh && \
    echo 'echo "📝 Logs: /app/logs/"' >> /app/start.sh && \
    echo 'echo "💾 Database: /data/ai-dep-manager.db"' >> /app/start.sh && \
    echo 'echo "✨ Unified Full-Stack Application with Comprehensive Logging"' >> /app/start.sh && \
    echo 'exec ./ai-dep-manager serve-simple' >> /app/start.sh && \
    chmod +x /app/start.sh

# Default command - start unified web server
CMD ["/app/start.sh"]
