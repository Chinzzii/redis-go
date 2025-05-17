# ────────────────
# Builder Stage
# ────────────────
FROM golang:1.23.5-alpine AS builder

# Install git (needed if any modules are pulled from git repos)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod
COPY go.mod ./
RUN go mod download

# Copy the rest of the source tree
COPY . .

# Build a statically‐linked binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o redis-server ./cmd/redis-server

# ────────────────
# Runtime Stage
# ────────────────
FROM alpine:latest

WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/redis-server .

# Expose Redis default port
EXPOSE 6379

# Create a directory for persistence files
RUN mkdir -p /root/data

# mount point for appendonly.aof and dump.rdb
VOLUME ["/root/data"]

# Run the server, setting working dir to /root so persistence ends up in /root/data
CMD ["./redis-server"]
