# --- Build Stage ---
FROM golang:1.22-alpine AS builder

# Install git for go mod download (useful if you have private modules)
RUN apk add --no-cache git

WORKDIR /app

# Copy go.mod and go.sum first to leverage Docker caching
COPY go.mod go.sum ./
RUN go mod download

# Copy the entire source
COPY . .

# Build the Go binary
RUN go build -o server api.go

# --- Runtime Stage ---
FROM alpine:3.20

WORKDIR /app

# Copy binary from builder
COPY --from=builder /app/server .

# Copy .env file (optional, for local testing)
COPY .env .env

# Expose the API port
EXPOSE 8081

# Run the binary
CMD ["./server"]
