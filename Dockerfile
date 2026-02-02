# Multi-stage build for Image Processor

# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

# Install required packages
RUN apk add --no-cache git make gcc musl-dev

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./backend/

# Build API service
FROM builder AS api-builder
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/api ./backend/cmd/api

# Build Worker service
FROM builder AS worker-builder
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o /app/worker ./backend/cmd/worker

# Stage 2: API Runtime
FROM alpine:latest AS api

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary and config
COPY --from=api-builder /app/api .
COPY backend/internal/config/config.yaml ./backend/internal/config/

EXPOSE 8080

CMD ["./api"]

# Stage 3: Worker Runtime
FROM alpine:latest AS worker

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary and config
COPY --from=worker-builder /app/worker .
COPY backend/internal/config/config.yaml ./backend/internal/config/

CMD ["./worker"]

