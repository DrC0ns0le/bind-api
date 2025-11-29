# syntax=docker/dockerfile:1.4

# Stage 1: Build the Go application
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go module files first for better caching
COPY go.mod go.sum ./

# Download dependencies with cache mount for faster rebuilds
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

# Copy the rest of the source code
COPY . .

# Build with cache mounts and optimized flags
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main .


# Stage 2: Final minimal image
FROM alpine:latest AS final

# Install runtime dependencies
RUN apk add --no-cache ansible openssh-client sshpass ca-certificates && \
    rm -rf /root/.cache /var/cache/apk/*

WORKDIR /app/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy templates folder
COPY --from=builder /app/render/templates /app/render/templates

# Copy Ansible configuration files
COPY --from=builder /app/ansible/deploy_config.yaml /app/ansible/deploy_config.yaml

ENTRYPOINT ["./main"]
