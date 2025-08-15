# Dockerfile for the Go application

# --- Build Stage ---
FROM golang:1.21-alpine AS builder

# Set the working directory inside the container
WORKDIR /app

# Copy go.mod and go.sum to download dependencies first
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the rest of the source code
COPY . .

# Build the application
# -o /app/server: output binary to /app/server
# CGO_ENABLED=0: disable CGO for a static binary
# GOOS=linux: build for Linux
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/server ./cmd/server

# --- Final Stage ---
FROM alpine:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the built binary from the builder stage
COPY --from=builder /app/server .

# Copy the migrations directory
COPY migrations ./migrations

# Copy the .env file
# Note: For production, it's better to manage secrets using Docker secrets or other secret management tools.
COPY .env.example ./.env

# Expose the application port
EXPOSE 8080

# Set the command to run the application
CMD ["./server"]
