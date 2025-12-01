# Use a minimal base image with Go installed
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./
# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN go build -o gocrypt .

# Final stage: minimal image to run the binary
FROM alpine:latest

WORKDIR /root/

COPY --from=builder /app/gocrypt .

# Expose the port
EXPOSE 8080

# Command to run the executable
CMD ["./gocrypt", "serve", "-port", "8080"]
