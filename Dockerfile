# Stage 1: Build the Go application
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy only necessary files for building, avoiding unnecessary rebuilds
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source code
COPY . .

# Download dependencies
RUN go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .


# Stage 2: Final minimal image
FROM alpine:latest AS final

RUN apk add --no-cache ansible openssh-client sshpass && \
    rm -rf /root/.cache /var/cache/apk/*

WORKDIR /app/

# Copy the binary from the builder stage
COPY --from=builder /app/main .

# Copy templates folder
COPY --from=builder /app/render/templates /app/render/templates

# Copy Ansible configuration files
COPY --from=builder /app/ansible/deploy_config.yaml /app/ansible/deploy_config.yaml

ENTRYPOINT ["./main"]