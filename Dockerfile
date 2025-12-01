# Step 1: Use Go 1.24 Alpine image for building
FROM golang:1.24-alpine AS builder

# Set working directory
WORKDIR /app

# Copy go.mod and go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code
COPY . .

# Build the Go binary
RUN go build -o gocrypt .

# Step 2: Use a minimal image for runtime
FROM alpine:latest

# Install ca-certificates if HTTPS needed (optional)
RUN apk --no-cache add ca-certificates

# Set working directory
WORKDIR /root/

# Copy the built binary from the builder stage
COPY --from=builder /app/gocrypt .

# Copy static files if any (like index.html)
COPY --from=builder /app/index.html .

# Expose port 8080 (or your app's port)
EXPOSE 8080

# Command to run the binary
CMD ["./gocrypt", "serve"]
