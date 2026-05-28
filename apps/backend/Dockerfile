# Build stage
FROM golang:1.25.7-bookworm AS builder

# Install build dependencies for go-fitz (MuPDF)
RUN apt-get update && apt-get install -y \
    libmupdf-dev \
    build-essential \
    pkg-config

WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build API binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o api ./cmd/api/main.go

# Build Worker binary
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o worker ./cmd/worker/main.go

# Final stage
FROM debian:bookworm-slim

# Install runtime dependencies for go-fitz and CA certificates
RUN apt-get update && apt-get install -y \
    libmupdf-dev \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy binaries from builder
COPY --from=builder /app/api /app/api
COPY --from=builder /app/worker /app/worker

# Default command (will be overridden in fly.toml for worker)
CMD ["/app/api"]
