# Stage 1: Build the application
FROM golang:1.24-alpine AS builder

# Install git and build essentials
RUN apk add --no-cache git build-base

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /rail-connect ./cmd/rail-connect/main.go
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o /rail-client ./client/example.go

# Stage 2: Create a minimal runtime image
FROM alpine:3.21

# Install CA certificates for HTTPS
RUN apk --no-cache add ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /rail-connect .
COPY --from=builder /rail-client .

# Copy configuration files
COPY config/config.yaml ./config/

# Expose the gRPC port
EXPOSE 50051

# Run the service
CMD ["./rail-connect"]