# Multi-stage build for Image Processor

# Stage 1: Build stage
FROM golang:1.25-alpine AS builder

# Install required packages
RUN apk add --no-cache git build-base

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

RUN go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy source code
COPY backend/ ./backend/

# Build API service
FROM builder AS api-builder
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/api ./backend/cmd/api/main.go

# Build Worker service
FROM builder AS worker-builder
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/worker ./backend/cmd/worker/main.go

# Stage 2: API Runtime
FROM alpine:latest AS api

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary and config
COPY --from=api-builder /app/api .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY backend/internal/config/config.yaml ./backend/internal/config/
COPY --from=builder /app/backend/migrations /root/backend/migrations


EXPOSE 8080

CMD ["./api"]

# Stage 3: Worker Runtime
FROM alpine:latest AS worker

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary and config
COPY --from=worker-builder /app/worker .
COPY --from=builder /go/bin/migrate /usr/local/bin/migrate
COPY backend/internal/config/config.yaml ./backend/internal/config/
COPY --from=builder /app/backend/migrations /root/backend/migrations


CMD ["./worker"]

